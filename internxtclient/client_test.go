package internxtclient_test

import (
	"testing"

	client "github.com/StarHack/go-internxt-drive/internxtclient"
)

func TestClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("NewWithCredentials", func(t *testing.T) {
		newWithCredentials(t)
	})
}

func newWithCredentials(t *testing.T) {
	c, err := client.NewWithCredentials(testEmail, testPassword)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if c.UserData == nil {
		t.Fatal("Expected UserData, got nil")
	}

	if c.UserData.AccessData == nil {
		t.Fatal("Expected UserData.AccessData, got nil")
	}

	if c.UserData.LoginData == nil {
		t.Fatal("Expected UserData.LoginData, got nil")
	}

	if c.UserData.AccessData.NewToken != "" {
		t.Logf("Login successful")
	} else {
		t.Fatal("Expected token, got empty string")
	}
}
