package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/internxt/rclone/config"
)

type TFAEnableRequest struct {
	Key  string `json:"key"`
	Code string `json:"code"`
}

type TFADisableRequest struct {
	Pass string `json:"pass"`
	Code string `json:"code"`
}

type TFASecretResponse struct {
	Code string `json:"code"`
	QR   string `json:"qr"`
}

// IsTFAEnabled returns true if TFA is already enabled, false if not
func IsTFAEnabled(cfg *config.Config) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.DriveAPIURL+"/auth/tfa", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return false, nil // Not enabled
	case http.StatusConflict:
		return true, nil // Already enabled
	default:
		return false, errors.New("unexpected TFA status code: " + resp.Status)
	}
}

// FetchTFASetup fetches the secret code and QR image for setting up TFA
func FetchTFASetup(cfg *config.Config) (*TFASecretResponse, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.DriveAPIURL+"/auth/tfa", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch TFA setup: " + resp.Status)
	}

	var data TFASecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

// EnableTFA enables two-factor authentication
func EnableTFA(cfg *config.Config, key, code string) error {
	body, _ := json.Marshal(TFAEnableRequest{Key: key, Code: code})

	req, err := http.NewRequest(http.MethodPut, cfg.DriveAPIURL+"/auth/tfa", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to enable TFA: " + resp.Status)
	}
	return nil
}

// DisableTFA disables two-factor authentication
func DisableTFA(cfg *config.Config, code string) error {
	bodyBytes, _ := json.Marshal(TFADisableRequest{
		Pass: cfg.EncryptedPassword,
		Code: code,
	})

	req, err := http.NewRequest(http.MethodDelete, cfg.DriveAPIURL+"/auth/tfa", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(bodyBytes))

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to disable TFA: %s - %s", resp.Status, string(respBody))
	}
	return nil
}
