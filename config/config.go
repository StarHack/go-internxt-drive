package config

import (
	"net"
	"net/http"
	"time"
)

const (
	DefaultDriveAPIURL      = "https://api.internxt.com/drive"
	DefaultAuthAPIURL       = "https://api.internxt.com/drive/auth"
	DefaultUsersAPIURL      = "https://api.internxt.com/users"
	DefaultAppCryptoSecret  = "6KYQBP847D4ATSFA"
	DefaultAppCryptoSecret2 = "8Q8VMUE3BJZV87GT"
	DefaultAppMagicIV       = "d139cb9a2cd17092e79e1861cf9d7023"
	DefaultAppMagicSalt     = "38dce0391b49efba88dbc8c39ebf868f0267eb110bb0012ab27dc52a528d61b1d1ed9d76f400ff58e3240028442b1eab9bb84e111d9dadd997982dbde9dbd25e"
)

type Config struct {
	Email             string       `json:"email,omitempty"`
	Password          string       `json:"password,omitempty"`
	TFA               string       `json:"tfa,omitempty"`
	Token             string       `json:"token,omitempty"`
	RootFolderID      string       `json:"root_folder_id,omitempty"`
	Bucket            string       `json:"bucket,omitempty"`
	Mnemonic          string       `json:"mnemonic,omitempty"`
	BasicAuthHeader   string       `json:"basic_auth_header,omitempty"`
	DriveAPIURL       string       `json:"drive_api_url,omitempty"`
	AuthAPIURL        string       `json:"auth_api_url,omitempty"`
	UsersAPIURL       string       `json:"users_api_url,omitempty"`
	AppCryptoSecret   string       `json:"app_crypto_secret,omitempty"`
	AppCryptoSecret2  string       `json:"app_crypto_secret2,omitempty"`
	AppMagicIV        string       `json:"app_magic_iv,omitempty"`
	AppMagicSalt      string       `json:"app_magic_salt,omitempty"`
	EncryptedPassword string       `json:"encrypted_password,omitempty"`
	HTTPClient        *http.Client `json:"-"`
}

func NewDefault(email, password string) *Config {
	cfg := &Config{
		Email:    email,
		Password: password,
	}
	cfg.applyDefaults()
	return cfg
}

func NewDefault2FA(email, password, tfa string) *Config {
	cfg := &Config{
		Email:    email,
		Password: password,
		TFA:      tfa,
	}
	cfg.applyDefaults()
	return cfg
}

func NewDefaultToken(token string) *Config {
	cfg := &Config{
		Token: token,
	}
	cfg.applyDefaults()
	return cfg
}

func (c *Config) applyDefaults() {
	if c.DriveAPIURL == "" {
		c.DriveAPIURL = DefaultDriveAPIURL
	}
	if c.AuthAPIURL == "" {
		c.AuthAPIURL = DefaultAuthAPIURL
	}
	if c.UsersAPIURL == "" {
		c.UsersAPIURL = DefaultUsersAPIURL
	}
	if c.AppCryptoSecret == "" {
		c.AppCryptoSecret = DefaultAppCryptoSecret
	}
	if c.AppCryptoSecret2 == "" {
		c.AppCryptoSecret2 = DefaultAppCryptoSecret2
	}
	if c.AppMagicIV == "" {
		c.AppMagicIV = DefaultAppMagicIV
	}
	if c.AppMagicSalt == "" {
		c.AppMagicSalt = DefaultAppMagicSalt
	}
	if c.HTTPClient == nil {
		c.HTTPClient = newHTTPClient()
	}
}

// newHTTPClient: properly configured HTTP client with sensible timeouts
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     50,
			IdleConnTimeout:     90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			ForceAttemptHTTP2:     true,
		},
	}
}
