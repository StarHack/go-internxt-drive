package workspaces

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/internxt/rclone/config"
)

const workspacesPath = "/workspaces"

// WorkspaceUser holds per-user settings within a workspace
type WorkspaceUser struct {
	ID           string          `json:"id"`
	MemberID     string          `json:"memberId"`
	Key          string          `json:"key"`
	WorkspaceID  string          `json:"workspaceId"`
	RootFolderID string          `json:"rootFolderId"`
	SpaceLimit   int64           `json:"spaceLimit"`
	DriveUsage   int64           `json:"driveUsage"`
	BackupsUsage int64           `json:"backupsUsage"`
	Deactivated  bool            `json:"deactivated"`
	Member       json.RawMessage `json:"member"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
}

// Workspace holds metadata about a workspace
type Workspace struct {
	ID              string `json:"id"`
	OwnerID         string `json:"ownerId"`
	Address         string `json:"address"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Avatar          string `json:"avatar"`
	DefaultTeamID   string `json:"defaultTeamId"`
	WorkspaceUserID string `json:"workspaceUserId"`
	SetupCompleted  bool   `json:"setupCompleted"`
	NumberOfSeats   int    `json:"numberOfSeats"`
	PhoneNumber     string `json:"phoneNumber"`
	RootFolderID    string `json:"rootFolderId"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// AvailableWorkspace ties a user to a workspace
type AvailableWorkspace struct {
	WorkspaceUser WorkspaceUser `json:"workspaceUser"`
	Workspace     Workspace     `json:"workspace"`
}

// WorkspacesResponse is the response for GET /drive/workspaces
type WorkspacesResponse struct {
	Available []AvailableWorkspace `json:"availableWorkspaces"`
	Pending   []json.RawMessage    `json:"pendingWorkspaces"`
}

// GetWorkspaces calls GET {DriveAPIURL}/workspaces and returns the parsed response
func GetWorkspaces(cfg *config.Config) (*WorkspacesResponse, error) {
	endpoint := cfg.DriveAPIURL + workspacesPath
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wr WorkspacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		// on decode error, print raw body for inspection
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		fmt.Println("raw response:", buf.String())
		return nil, err
	}
	return &wr, nil
}
