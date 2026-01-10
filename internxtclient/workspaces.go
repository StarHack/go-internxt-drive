package internxtclient

import (
	"encoding/json"
	"path"
)

type WorkspacesService struct {
	client *Client
}

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
func (w *WorkspacesService) GetWorkspaces() (*WorkspacesResponse, error) {
	endpoint := path.Join(workspacesPath)

	var workspacesResponse WorkspacesResponse

	if resp, err := w.client.Get(APITypeDrive, endpoint, &workspacesResponse, nil); err != nil {
		return nil, w.client.GetError(endpoint, resp, err)
	}
	return &workspacesResponse, nil
}
