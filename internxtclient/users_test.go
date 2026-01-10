package internxtclient_test

import (
	"testing"

	client "github.com/StarHack/go-internxt-drive/internxtclient"
)

func TestUsersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("GetLimit", func(t *testing.T) {
		getLimit(t)
	})

	t.Run("GetUsage", func(t *testing.T) {
		getUsage(t)
	})

	t.Run("GetUserCredentials", func(t *testing.T) {
		getUserCredentials(t)
	})
}

func getLimit(t *testing.T) *client.LimitResponse {
	limit, err := c.Users.GetLimit()
	if err != nil {
		t.Fatalf("Error getting limit: %v", err)
	}
	if limit == nil {
		t.Fatalf("Limit is nil")
	}

	return limit
}

func getUsage(t *testing.T) *client.UsageResponse {
	usage, err := c.Users.GetUsage()
	if err != nil {
		t.Fatalf("Error getting usage: %v", err)
	}
	if usage == nil {
		t.Fatalf("Usage is nil")
	}

	return usage
}

func getUserCredentials(t *testing.T) *client.GetUserCredentialsResponse {
	credentials, err := c.Users.GetUserCredentials()
	if err != nil {
		t.Fatalf("Error getting credentials: %v", err)
	}
	if credentials == nil {
		t.Fatalf("Credentials is nil")
	}
	if credentials.NewToken == "" {
		t.Fatalf("NewToken is empty")
	}
	return credentials
}
