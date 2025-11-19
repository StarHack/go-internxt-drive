package auth

import (
	"fmt"
	"io"
	"net/http"

	"github.com/internxt/rclone/config"
)

// Logout calls GET {DRIVE_API_URL}/logout to end the session.
// It uses cfg.Token as the bearer token.
func Logout(cfg *config.Config) error {
	url := cfg.DriveAPIURL + "/auth/logout"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}
