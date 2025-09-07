package internxtclient

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"

	"golang.org/x/crypto/pbkdf2"
)

type AuthService struct {
	client *Client
}

type LoginResponse struct {
	HasKeys      bool   `json:"hasKeys"`
	SKey         string `json:"sKey"`
	TFA          bool   `json:"tfa"`
	HasKyberKeys bool   `json:"hasKyberKeys"`
	HasECCKeys   bool   `json:"hasEccKeys"`
}

type AccessResponse struct {
	User     *User           `json:"user"`
	Token    string          `json:"token"`
	UserTeam json.RawMessage `json:"userTeam"`
	NewToken string          `json:"newToken"`
}

const authPath = "auth"

// Login calls {DRIVE_API_URL}/auth/login with {"email":…}
func (a *AuthService) Login(email string) (*LoginResponse, error) {
	endpoint := path.Join(authPath, "login")
	payload := map[string]string{
		"email": email,
	}
	var loginResponse LoginResponse

	if resp, err := a.client.Post(APITypeDrive, endpoint, &payload, &loginResponse, nil); err != nil {
		return nil, a.client.GetError(endpoint, resp, err)
	}

	return &loginResponse, nil
}

// Logout calls {DRIVE_API_URL}/auth/logout. Returns error if failed.
func (a *AuthService) Logout() error {
	endpoint := path.Join(authPath, "logout")

	if resp, err := a.client.Get(APITypeDrive, endpoint, nil, nil); err != nil {
		return a.client.GetError(endpoint, resp, err)
	}

	return nil
}

// AccessLogin calls {DRIVE_API_URL}/auth/login/access based on our previous LoginResponse
func (a *AuthService) AccessLogin(loginResponse *LoginResponse, password string) (*AccessResponse, error) {
	endpoint := path.Join(authPath, "login", "access")
	encPwd, passwordHash, err := deriveEncryptedPasswordAndHash(password, loginResponse.SKey, a.client.Config.AppCryptoSecret)
	if err != nil {
		return nil, err
	}
	a.client.Config.EncryptedPassword = encPwd
	a.client.Config.PasswordHash = passwordHash

	req := map[string]interface{}{
		"email":    a.client.UserData.AccessData.User.Email,
		"password": encPwd,
	}

	// TODO
	/*
		if lr.TFA && a.client.UserData.LoginData.TFA != "" {
			req["tfa"] = a.client.UserData.LoginData.TFA
		}
	*/

	var accessResponse AccessResponse

	if resp, err := a.client.Post(APITypeDrive, endpoint, &req, &accessResponse, nil); err != nil {
		return nil, a.client.GetError(endpoint, resp, err)
	}

	a.client.UserData.AccessData = &accessResponse

	// 1) SHA256 the raw pass string
	sum := sha256.Sum256([]byte(accessResponse.User.UserID))
	hexPass := hex.EncodeToString(sum[:])

	// 2) build "user:hexPass" and Base64
	creds := fmt.Sprintf("%s:%s", accessResponse.User.BridgeUser, hexPass)
	a.client.UserData.BasicAuthHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))

	a.client.UserData.AccessData.User.Mnemonic, err = decryptTextWithKey(accessResponse.User.Mnemonic, password)
	if err != nil {
		return nil, err
	}

	return &accessResponse, nil
}

func (a *AuthService) AreCredentialsCorrect(hashedPassword string) (bool, error) {
	endpoint := path.Join(authPath, "are-credentials-correct")
	opts := map[string]string{
		"hashedPassword": hashedPassword,
	}

	if resp, err := a.client.doRequestWithQuery(APITypeDrive, http.MethodGet, endpoint, opts, nil, nil, nil); err != nil {
		return false, a.client.GetError(endpoint, resp, err)
	}

	return true, nil
}

// Gets the encrypted password and the PasswordHash
func deriveEncryptedPasswordAndHash(password, hexSalt, secret string) (string, string, error) {
	// decrypt the OpenSSL‐style salt blob to hex salt string
	saltHex, err := decryptTextWithKey(hexSalt, secret)
	if err != nil {
		return "", "", err
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", "", err
	}

	// PBKDF2‐SHA1
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha1.New)
	hashHex := hex.EncodeToString(key)

	// re‐encrypt with OpenSSL style AES‑CBC
	encryptedPass, err := encryptTextWithKey(hashHex, secret)
	if err != nil {
		return "", "", err
	}

	return encryptedPass, hashHex, nil
}

func decryptTextWithKey(hexCipher, secret string) (string, error) {
	data, err := hex.DecodeString(hexCipher)
	if err != nil {
		return "", err
	}
	if len(data) < 16 {
		return "", errors.New("failed to login")
	}
	salt := data[8:16]
	// EVP_BytesToKey with MD5 ×3
	d := append([]byte(secret), salt...)
	var prev = d
	hashes := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		h := md5.Sum(prev)
		hashes[i] = h[:]
		prev = append(hashes[i], d...)
	}
	key := append(hashes[0], hashes[1]...)
	iv := hashes[2]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	ct := data[16:]
	pt := make([]byte, len(ct))
	mode.CryptBlocks(pt, ct)
	// strip PKCS#7
	pad := int(pt[len(pt)-1])
	pt = pt[:len(pt)-pad]
	return string(pt), nil
}

func encryptTextWithKey(plaintext, secret string) (string, error) {
	salt := make([]byte, 8)
	_, _ = rand.Read(salt)
	d := append([]byte(secret), salt...)
	var prev = d
	hashes := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		h := md5.Sum(prev)
		hashes[i] = h[:]
		prev = append(hashes[i], d...)
	}
	key := append(hashes[0], hashes[1]...)
	iv := hashes[2]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	// PKCS#7 pad
	padLen := aes.BlockSize - len(plaintext)%aes.BlockSize
	for i := 0; i < padLen; i++ {
		plaintext += string(byte(padLen))
	}
	ct := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ct, []byte(plaintext))

	out := append([]byte("Salted__"), salt...)
	out = append(out, ct...)
	return hex.EncodeToString(out), nil
}
