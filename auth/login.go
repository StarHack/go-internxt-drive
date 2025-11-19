// auth/login.go
package auth

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/internxt/rclone/config"
)

type LoginResponse struct {
	HasKeys      bool   `json:"hasKeys"`
	SKey         string `json:"sKey"`
	TFA          bool   `json:"tfa"`
	HasKyberKeys bool   `json:"hasKyberKeys"`
	HasECCKeys   bool   `json:"hasEccKeys"`
}

// Login calls {DRIVE_API_URL}/auth/login with {"email":…}
func Login(cfg *config.Config) (*LoginResponse, error) {
	payload := map[string]string{"email": cfg.Email}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, cfg.DriveAPIURL+"/auth/login", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lr LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, err
	}
	return &lr, nil
}
