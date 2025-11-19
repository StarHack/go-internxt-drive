package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"crypto/sha1"

	"github.com/internxt/rclone/config"

	"golang.org/x/crypto/pbkdf2"
)

type AccessResponse struct {
	User struct {
		Email               string `json:"email"`
		UserID              string `json:"userId"`
		Mnemonic            string `json:"mnemonic"`
		PrivateKey          string `json:"privateKey"`
		PublicKey           string `json:"publicKey"`
		RevocateKey         string `json:"revocateKey"`
		RootFolderID        string `json:"rootFolderId"`
		Name                string `json:"name"`
		Lastname            string `json:"lastname"`
		UUID                string `json:"uuid"`
		Credit              int    `json:"credit"`
		CreatedAt           string `json:"createdAt"`
		Bucket              string `json:"bucket"`
		RegisterCompleted   bool   `json:"registerCompleted"`
		Teams               bool   `json:"teams"`
		Username            string `json:"username"`
		BridgeUser          string `json:"bridgeUser"`
		SharedWorkspace     bool   `json:"sharedWorkspace"`
		HasReferralsProgram bool   `json:"hasReferralsProgram"`
		BackupsBucket       string `json:"backupsBucket"`
		Avatar              string `json:"avatar"`
		EmailVerified       bool   `json:"emailVerified"`
		LastPasswordChanged string `json:"lastPasswordChangedAt"`
	} `json:"user"`
	Token    string          `json:"token"`
	UserTeam json.RawMessage `json:"userTeam"`
	NewToken string          `json:"newToken"`
}

// AccessLogin calls {DRIVE_API_URL}/auth/login/access based on our previous LoginResponse
func AccessLogin(cfg *config.Config, lr *LoginResponse) (*AccessResponse, error) {
	encPwd, err := deriveEncryptedPassword(cfg.Password, lr.SKey, cfg.AppCryptoSecret)
	if err != nil {
		return nil, err
	}
	cfg.EncryptedPassword = encPwd

	req := map[string]interface{}{
		"email":    cfg.Email,
		"password": encPwd,
	}
	if lr.TFA && cfg.TFA != "" {
		req["tfa"] = cfg.TFA
	}

	b, _ := json.Marshal(req)

	httpReq, err := http.NewRequest(http.MethodPost, cfg.DriveAPIURL+"/auth/login/access", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ar AccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, err
	}

	cfg.Token = ar.NewToken
	cfg.RootFolderID = ar.User.RootFolderID
	cfg.Bucket = ar.User.Bucket

	// 1) SHA256 the raw pass string
	sum := sha256.Sum256([]byte(ar.User.UserID))
	hexPass := hex.EncodeToString(sum[:])

	// 2) build "user:hexPass" and Base64
	creds := fmt.Sprintf("%s:%s", ar.User.BridgeUser, hexPass)
	cfg.BasicAuthHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))

	cfg.Mnemonic, err = decryptTextWithKey(ar.User.Mnemonic, cfg.Password)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}

func AreCredentialsCorrect(cfg *config.Config, hashedPassword string) (bool, error) {
	endpoint := fmt.Sprintf("%s/auth/are-credentials-correct?hashedPassword=%s", cfg.DriveAPIURL, url.QueryEscape(hashedPassword))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func deriveEncryptedPassword(password, hexSalt, secret string) (string, error) {
	// decrypt the OpenSSL‐style salt blob to hex salt string
	saltHex, err := decryptTextWithKey(hexSalt, secret)
	if err != nil {
		return "", err
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", err
	}
	// PBKDF2‐SHA1
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha1.New)
	hashHex := hex.EncodeToString(key)

	// re‐encrypt with OpenSSL style AES‑CBC
	return encryptTextWithKey(hashHex, secret)
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
