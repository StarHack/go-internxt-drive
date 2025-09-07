package internxtclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/appscode/go-querystring/query"
)

const (
	DefaultDriveAPIURL  = "https://api.internxt.com/drive"
	DefaultAuthAPIURL   = "https://api.internxt.com/drive/auth"
	DefaultUsersAPIURL  = "https://api.internxt.com/users"
	DefaultBucketAPIURL = "https://api.internxt.com/buckets"
	DefaultBaseAPIURL   = "https://api.internxt.com"

	DefaultAppCryptoSecret  = "6KYQBP847D4ATSFA"
	DefaultAppCryptoSecret2 = "8Q8VMUE3BJZV87GT"
	DefaultAppMagicIV       = "d139cb9a2cd17092e79e1861cf9d7023"
	DefaultAppMagicSalt     = "38dce0391b49efba88dbc8c39ebf868f0267eb110bb0012ab27dc52a528d61b1d1ed9d76f400ff58e3240028442b1eab9bb84e111d9dadd997982dbde9dbd25e"
)

type APIType int

const (
	APITypeDrive APIType = iota
	APITypeAuth
	APITypeUsers
	APITypeBucket
	APITypeBase
)

// Getter for convenience
func (c *Client) URL(api APIType) string {
	return c.Config.APIURLs[api]
}

type Client struct {
	Config     Config
	HTTPClient *http.Client
	UserData   *UserData

	Folders    *FoldersService
	Files      *FilesService
	Auth       *AuthService
	Users      *UsersService
	Fuzzy      *FuzzyService
	Buckets    *BucketsService
	Workspaces *WorkspacesService
	Trash      *TrashService
}

type Config struct {
	APIURLs map[APIType]string `json:"api_urls,omitempty"`

	AppCryptoSecret   string `json:"app_crypto_secret,omitempty"`
	AppCryptoSecret2  string `json:"app_crypto_secret2,omitempty"`
	AppMagicIV        string `json:"app_magic_iv,omitempty"`
	AppMagicSalt      string `json:"app_magic_salt,omitempty"`
	EncryptedPassword string `json:"encrypted_password,omitempty"`
	PasswordHash      string `json:"password_hash,omitempty"`
}

type UserData struct {
	LoginData       *LoginResponse  `json:"login,omitempty"`
	AccessData      *AccessResponse `json:"access,omitempty"`
	BasicAuthHeader string          `json:"basic_auth_header,omitempty"`
}

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func defaultAPIURLs() map[APIType]string {
	return map[APIType]string{
		APITypeDrive:  DefaultDriveAPIURL,
		APITypeAuth:   DefaultAuthAPIURL,
		APITypeUsers:  DefaultUsersAPIURL,
		APITypeBucket: DefaultBucketAPIURL,
		APITypeBase:   DefaultBaseAPIURL,
	}
}

func (c *Config) applyDefaults() {
	if c.APIURLs == nil {
		c.APIURLs = defaultAPIURLs()
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
}

func NewWithDefaults() *Client {
	c := Client{}
	c.Folders = &FoldersService{client: &c}
	c.Files = &FilesService{client: &c}
	c.Auth = &AuthService{client: &c}
	c.Users = &UsersService{client: &c}
	c.Fuzzy = &FuzzyService{client: &c}
	c.Buckets = &BucketsService{client: &c}
	c.Workspaces = &WorkspacesService{client: &c}
	c.Trash = &TrashService{client: &c}

	c.Config.applyDefaults()

	c.UserData = &UserData{
		AccessData: &AccessResponse{},
		LoginData:  &LoginResponse{},
	}

	c.HTTPClient = &http.Client{
		Timeout: time.Second * 30,
	}

	return &c
}

func NewWithCredentials(email, password string) (*Client, error) {
	c := NewWithDefaults()

	LoginResponse, err := c.Auth.Login(email)
	if err != nil {
		return nil, err
	}

	c.UserData.LoginData = LoginResponse

	c.UserData.AccessData.User = &User{Email: email}
	AccessResponse, err := c.Auth.AccessLogin(LoginResponse, password)
	if err != nil {
		return nil, err
	}

	c.UserData.AccessData = AccessResponse

	return c, nil
}

func (c *Client) hasUserData() bool {
	return c.UserData != nil
}

func (c *Client) hasUserDataAccessData() bool {
	if c.hasUserData() {
		return c.UserData.AccessData != nil
	}
	return false
}

func (c *Client) hasUserDataAccessDataUser() bool {
	if c.hasUserDataAccessData() {
		return c.UserData.AccessData.User != nil
	}
	return false
}

func (c *Client) hasUserDataLoginData() bool {
	if c.hasUserData() {
		return c.UserData.LoginData != nil
	}
	return false
}

// doRequest handles sending the request and decoding the response into result.
func (c *Client) doRequest(apiType APIType, method, path string, body any, result any, headers *http.Header) (*Response, error) {
	finalUrl, err := url.JoinPath(c.URL(apiType), path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	finalUrlUnescaped, err := url.PathUnescape(finalUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot unescape URL: %w", err)
	}

	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode body: %w", err)
		}
		buf = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, finalUrlUnescaped, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// If we didn't receive headers, assume json
	if headers == nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("internxt-client", "go-internxt-drive")
	} else {
		req.Header = *headers
	}

	if c.hasUserDataAccessData() && c.UserData.AccessData.NewToken != "" {
		// Check if it was already set
		if req.Header.Get("Authorization") == "" {
			req.Header.Set("Authorization", "Bearer "+c.UserData.AccessData.NewToken)
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read response body: %w", err)
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       b,
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	if result != nil {
		if err := json.Unmarshal(b, result); err != nil {
			return response, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return response, nil
}

func (c *Client) doRequestWithQuery(apiType APIType, method, endpoint string, query map[string]string, body, out any, headers *http.Header) (*Response, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return c.doRequest(apiType, method, u.String(), body, out, headers)
}

func (c *Client) doRequestWithStruct(apiType APIType, method, endpoint string, opts any, body, out any, headers *http.Header) (*Response, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if opts != nil {
		v, err := query.Values(opts)
		if err != nil {
			return nil, err
		}
		u.RawQuery = v.Encode()
	}
	return c.doRequest(apiType, method, u.String(), body, out, headers)
}

// Get sends an HTTP GET request with optional headers to the given APIType and path, unmarshaling the response into result.
func (c *Client) Get(apiType APIType, path string, result any, headers *http.Header) (*Response, error) {
	return c.doRequest(apiType, http.MethodGet, path, nil, result, headers)
}

// Get sends an HTTP PATCH request with optional headers to the given APIType and path, unmarshaling the response into result.
func (c *Client) Patch(apiType APIType, path string, body any, result any, headers *http.Header) (*Response, error) {
	return c.doRequest(apiType, http.MethodPatch, path, body, result, headers)
}

// Get sends an HTTP POST request with optional headers to the given APIType and path, unmarshaling the response into result.
func (c *Client) Post(apiType APIType, path string, body any, result any, headers *http.Header) (*Response, error) {
	return c.doRequest(apiType, http.MethodPost, path, body, result, headers)
}

// Get sends an HTTP PUT request with optional headers to the given APIType and path, unmarshaling the response into result.
func (c *Client) Put(apiType APIType, path string, body any, result any, headers *http.Header) (*Response, error) {
	return c.doRequest(apiType, http.MethodPut, path, body, result, headers)
}

// Get sends an HTTP DELETE request with optional headers to the given APIType and path, unmarshaling the response into result.
func (c *Client) Delete(apiType APIType, path string, body, result any, headers *http.Header) (*Response, error) {
	return c.doRequest(apiType, http.MethodDelete, path, body, result, headers)
}

func (c *Client) GetError(endpoint string, resp *Response, err error) error {
	pc, _, _, ok := runtime.Caller(1)
	fnName := "unknown"
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			parts := strings.Split(fn.Name(), ".")
			fnName = parts[len(parts)-1] // just the method name
		}
	}

	if resp != nil {
		const maxLen = 200
		bodyStr := string(resp.Body)
		if len(bodyStr) > maxLen {
			bodyStr = bodyStr[:maxLen] + "...(truncated)"
		}
		return fmt.Errorf("%s failed when getting %s: %d %s -> %w",
			fnName, endpoint, resp.StatusCode, bodyStr, err)
	} else {
		return fmt.Errorf("%s failed when getting %s -> %w",
			fnName, endpoint, err)
	}
}
