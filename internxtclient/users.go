package internxtclient

import (
	"fmt"
	"path"
	"time"
)

type UsersService struct {
	client *Client
}

type User struct {
	Email          string    `json:"email"`
	UserID         string    `json:"userId"`
	Mnemonic       string    `json:"mnemonic"`
	RootFolderID   int       `json:"root_folder_id"`
	RootFolderUUID string    `json:"rootFolderId"`
	Name           string    `json:"name"`
	Lastname       string    `json:"lastname"`
	UUID           string    `json:"uuid"`
	Credit         int       `json:"credit"`
	CreatedAt      time.Time `json:"createdAt"`
	PrivateKey     string    `json:"privateKey"`
	PublicKey      string    `json:"publicKey"`
	RevocateKey    string    `json:"revocateKey"`
	Keys           struct {
		Ecc struct {
			PrivateKey string `json:"privateKey"`
			PublicKey  string `json:"publicKey"`
		} `json:"ecc"`
		Kyber struct {
			PrivateKey string `json:"privateKey"`
			PublicKey  string `json:"publicKey"`
		} `json:"kyber"`
	} `json:"keys"`
	Bucket                string `json:"bucket"`
	RegisterCompleted     bool   `json:"registerCompleted"`
	Teams                 bool   `json:"teams"`
	Username              string `json:"username"`
	BridgeUser            string `json:"bridgeUser"`
	SharedWorkspace       bool   `json:"sharedWorkspace"`
	AppSumoDetails        any    `json:"appSumoDetails"`
	HasReferralsProgram   bool   `json:"hasReferralsProgram"`
	BackupsBucket         any    `json:"backupsBucket"`
	Avatar                any    `json:"avatar"`
	EmailVerified         bool   `json:"emailVerified"`
	LastPasswordChangedAt any    `json:"lastPasswordChangedAt"`
}

type GetUserCredentialsResponse struct {
	User struct {
		ID         int    `json:"id"`
		UserID     string `json:"userId"`
		Name       string `json:"name"`
		Lastname   string `json:"lastname"`
		Email      string `json:"email"`
		Username   string `json:"username"`
		BridgeUser string `json:"bridgeUser"`
		Password   struct {
			Type string `json:"type"`
			Data []byte `json:"data"`
		} `json:"password"`
		Mnemonic struct {
			Type string `json:"type"`
			Data []byte `json:"data"`
		} `json:"mnemonic"`
		RootFolderID int `json:"rootFolderId"`
		HKey         struct {
			Type string `json:"type"`
			Data []byte `json:"data"`
		} `json:"hKey"`
		Secret2FA             any       `json:"secret_2FA"`
		ErrorLoginCount       int       `json:"errorLoginCount"`
		IsEmailActivitySended bool      `json:"isEmailActivitySended"`
		ReferralCode          string    `json:"referralCode"`
		Referrer              any       `json:"referrer"`
		SyncDate              any       `json:"syncDate"`
		UUID                  string    `json:"uuid"`
		LastResend            any       `json:"lastResend"`
		Credit                int       `json:"credit"`
		WelcomePack           bool      `json:"welcomePack"`
		RegisterCompleted     bool      `json:"registerCompleted"`
		BackupsBucket         any       `json:"backupsBucket"`
		SharedWorkspace       bool      `json:"sharedWorkspace"`
		Avatar                any       `json:"avatar"`
		LastPasswordChangedAt any       `json:"lastPasswordChangedAt"`
		TierID                string    `json:"tierId"`
		EmailVerified         bool      `json:"emailVerified"`
		UpdatedAt             time.Time `json:"updatedAt"`
		CreatedAt             time.Time `json:"createdAt"`
	} `json:"user"`
	OldToken string `json:"oldToken"`
	NewToken string `json:"newToken"`
}

type LimitResponse struct {
	MaxSpaceBytes int64 `json:"maxSpaceBytes"`
}

type UsageResponse struct {
	Drive int64 `json:"drive"`
}

const userPath = "users"

// GetUserCredentials gets the user's data by user uuid
func (u *UsersService) GetUserCredentials() (*GetUserCredentialsResponse, error) {
	if !u.client.hasUserDataAccessDataUser() {
		return nil, fmt.Errorf("can't get user credentials, missing user data")
	}

	endpoint := path.Join(userPath, "c", u.client.UserData.AccessData.User.UUID)

	var userCredentials GetUserCredentialsResponse

	resp, err := u.client.Get(APITypeDrive, endpoint, &userCredentials, nil)
	if err != nil {
		return nil, u.client.GetError(endpoint, resp, err)
	}

	return &userCredentials, nil
}

// GetLimit calls {DRIVE_API_URL}/users/limit and returns the maximum available storage of the account.
func (u *UsersService) GetLimit() (*LimitResponse, error) {
	endpoint := path.Join("users", "limit")

	var limit LimitResponse

	if resp, err := u.client.Get(APITypeDrive, endpoint, &limit, nil); err != nil {
		return nil, u.client.GetError(endpoint, resp, err)
	}

	return &limit, nil
}

// GetUsage calls GET {DRIVE_API_URL}/users/usage and returns the account's current usage in bytes.
func (u *UsersService) GetUsage() (*UsageResponse, error) {
	endpoint := path.Join("users", "usage")

	var usage UsageResponse

	if resp, err := u.client.Get(APITypeDrive, endpoint, &usage, nil); err != nil {
		return nil, u.client.GetError(endpoint, resp, err)
	}

	return &usage, nil
}
