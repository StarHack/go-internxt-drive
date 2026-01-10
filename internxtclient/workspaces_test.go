package internxtclient_test

import (
	"testing"

	client "github.com/StarHack/go-internxt-drive/internxtclient"
)

func TestWorkspacesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("GetWorkspaces", func(t *testing.T) {
		getWorkspaces(t)
	})

}

func getWorkspaces(t *testing.T) *client.WorkspacesResponse {
	workspacesResponse, err := c.Workspaces.GetWorkspaces()
	if err != nil {
		t.Fatalf("Error getting workspaces: %v", err)
	}
	if workspacesResponse == nil {
		t.Fatalf("Workspaces is nil")
	}

	return workspacesResponse
}
