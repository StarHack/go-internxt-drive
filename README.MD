# go-internxt-drive

> Updated to `client` branch which introduced some breaking changes. Old codebase was in commit `8c3b319`

A pure Go library allowing you to interact with the [internxt Drive API](https://api.internxt.com/drive/). Please refer to `docs/` directory for examples on how to get started.

Currently supports:

- Auth **(all)**
  - Login (`docs/login.md`)
  - Logout
  - 2FA
- File
  - Upload file
  - Download file
  - Delete
  - Rename
  - Move
- Folder
  - Create
  - Delete
  - Rename
  - Move
  - List files
  - List directories
- Fuzzy Search **(all)**
- Trash **(all)**

## Status

Implementation is **WIP** and here's the current status.

### Auth

| Method | Endpoint                              | Description                                   | Implemented |
| ------ | ------------------------------------- | --------------------------------------------- | ----------- |
| POST   | `/drive/auth/login`                   | Get security details to log in                | Yes         |
| POST   | `/drive/auth/login/access`            | Access user account                           | Yes         |
| GET    | `/drive/auth/logout`                  | Log out of the account                        | Yes         |
| GET    | `/drive/auth/tfa`                     | Get two-factor authentication                 | Yes          |
| PUT    | `/drive/auth/tfa`                     | Update two-factor authentication              | Yes          |
| DELETE | `/drive/auth/tfa`                     | Delete two-factor authentication              | Yes          |
| GET    | `/drive/auth/are-credentials-correct` | Check if current user credentials are correct | Yes          |

### File

| Method | Endpoint                           | Description                           | Implemented |
| ------ | ---------------------------------- | ------------------------------------- | ----------- |
| POST   | `/drive/files`                     | Create File                           | No          |
| GET    | `/drive/files`                     | —                                     | No          |
| GET    | `/drive/files/count`               | —                                     | No          |
| GET    | `/drive/files/{uuid}/meta`         | —                                     | Yes          |
| PUT    | `/drive/files/{uuid}/meta`         | Update File data                      | Yes          |
| PUT    | `/drive/files/{uuid}`              | —                                     | No          |
| PATCH  | `/drive/files/{uuid}`              | Move a file to a new dir UUID         | Yes          |
| DELETE | `/drive/files/{uuid}`              | Delete file from storage and database | Yes         |
| GET    | `/drive/files/recents`             | Retrieves a list of recent files with the given limit                                     | Yes          |
| GET    | `/drive/files/meta`                | —                                     | No          |
| POST   | `/drive/files/thumbnail`           | Create Thumbnail                      | No          |
| DELETE | `/drive/files/{bucketId}/{fileId}` | Delete file from storage by fileId    | No          |

### Folder

| Method | Endpoint                                          | Description                     | Implemented |
| ------ | ------------------------------------------------- | ------------------------------- | ----------- |
| POST   | `/drive/folders`                                  | Create Folder                   | Yes         |
| DELETE | `/drive/folders`                                  | —                               | Yes         |
| GET    | `/drive/folders`                                  | —                               | No          |
| GET    | `/drive/folders/count`                            | —                               | No          |
| GET    | `/drive/folders/content/{uuid}/files`             | —                               | Yes         |
| GET    | `/drive/folders/{id}/files`                       | —                               | No          |
| GET    | `/drive/folders/{id}/file`                        | —                               | No          |
| GET    | `/drive/folders/content/{uuid}/folders`           | —                               | Yes         |
| POST   | `/drive/folders/content/{uuid}/folders/existence` | Checks folders existence (POST) | No          |
| POST   | `/drive/folders/content/{uuid}/files/existence`   | Checks files existence          | No          |
| GET    | `/drive/folders/content/{uuid}`                   | Gets folder content             | No          |
| GET    | `/drive/folders/{id}/folders`                     | —                               | No          |
| GET    | `/drive/folders/{uuid}/meta`                      | Retrieves metadata for a folder by UUID                               | Yes          |
| PUT    | `/drive/folders/{uuid}/meta`                      | Renames a folder                | Yes         |
| GET    | `/drive/folders/{uuid}/ancestors`                 | Retrieves the ancestor folders of the given folder UUID                               | Yes          |
| GET    | `/drive/folders/{uuid}/tree`                      | Retrieves a full recursive folder tree starting from the given UUID                               | Yes          | 
| GET    | `/drive/folders/{id}/metadata`                    | Retrieves metadata for a folder by numeric ID                               | Yes          |
| GET    | `/drive/folders/{uuid}/size`                      | —                               | Yes         |
| PATCH  | `/drive/folders/{uuid}`                           | Moves folder to a new parent    | Yes         |
| DELETE | `/drive/folders/{uuid}`                           | Delete Folder                   | Yes         |
| GET    | `/drive/folders/meta`                             | Retrieves folder metadata using a full path string                               | Yes          |

### Share

| Method | Endpoint                                 | Description                        | Implemented |
| ------ | ---------------------------------------- | ---------------------------------- | ----------- |
| GET    | `/drive/storage/share/domains`           | —                                  | No          |
| GET    | `/drive/storage/share/list`              | Get share list                     | No          |
| GET    | `/drive/storage/share/{token}`           | Get share by token                 | No          |
| PUT    | `/drive/storage/share/{token}/view`      | Increment share view by token      | No          |
| PUT    | `/drive/storage/share/{shareId}`         | Update share by id                 | No          |
| DELETE | `/drive/storage/share/{shareId}`         | Delete share by id                 | No          |
| POST   | `/drive/storage/share/file/{fileId}`     | Create share for file              | No          |
| POST   | `/drive/storage/share/folder/{folderId}` | Create share for folder            | No          |
| GET    | `/drive/storage/share/down/files`        | Get all files by token paginated   | No          |
| GET    | `/drive/storage/share/down/folders`      | Get all folders by token paginated | No          |

## Sharing

| Method | Endpoint                                                | Description                          | Implemented |
| ------ | ------------------------------------------------------- | ------------------------------------ | ----------- |
| GET    | `/drive/sharings/{sharingId}/meta`                      | Get sharing metadata                 | No          |
| GET    | `/drive/sharings/public/{sharingId}/item`               | Get sharing item info                | No          |
| PATCH  | `/drive/sharings/{sharingId}/password`                  | Set password for public sharing      | No          |
| DELETE | `/drive/sharings/{sharingId}/password`                  | Remove password from public sharing  | No          |
| GET    | `/drive/sharings/{itemType}/{itemId}/invites`           | List invites for an item             | No          |
| PUT    | `/drive/sharings/{itemType}/{itemId}/type`              | Change sharing type for an item      | No          |
| GET    | `/drive/sharings/{itemType}/{itemId}/type`              | Get sharing type for an item         | No          |
| GET    | `/drive/sharings/{itemType}/{itemId}/info`              | Get info related to item sharing     | No          |
| GET    | `/drive/sharings/invites`                               | Get all invites received by the user | No          |
| POST   | `/drive/sharings/invites/send`                          | Send a sharing invite                | No          |
| GET    | `/drive/sharings/invites/{id}/validate`                 | Validate a sharing invite            | No          |
| POST   | `/drive/sharings/invites/{id}/accept`                   | Accept a sharing invite              | No          |
| DELETE | `/drive/sharings/invites/{id}`                          | Delete a sharing invite              | No          |
| GET    | `/drive/sharings/items/{sharedFolderId}/folders`        | Get folders in a shared folder       | No          |
| GET    | `/drive/sharings/items/{sharedFolderId}/files`          | Get files in a shared folder         | No          |
| GET    | `/drive/sharings/public/items/{sharedFolderId}/files`   | Get files in a public share          | No          |
| GET    | `/drive/sharings/public/items/{sharedFolderId}/folders` | Get folders in a public share        | No          |
| POST   | `/drive/sharings`                                       | Share an item                        | No          |
| DELETE | `/drive/sharings/{itemType}/{itemId}`                   | Stop sharing an item                 | No          |
| GET    | `/drive/sharings/roles`                                 | List sharing roles                   | No          |
| GET    | `/drive/sharings/{sharingId}/role`                      | Get role of a sharing                | No          |
| PUT    | `/drive/sharings/{sharingId}/role`                      | Set role of a sharing                | No          |
| DELETE | `/drive/sharings/{sharingId}/roles/{sharingRoleId}`     | Delete a sharing role                | No          |
| GET    | `/drive/sharings/shared-with-me/folders`                | Folders shared with me               | No          |
| GET    | `/drive/sharings/shared-by-me/folders`                  | Folders I’ve shared                  | No          |
| GET    | `/drive/sharings/folders`                               | All folder sharings                  | No          |
| GET    | `/drive/sharings/files`                                 | All file sharings                    | No          |
| GET    | `/drive/sharings/shared-with/{itemType}/{itemId}`       | Users with access to an item         | No          |
| GET    | `/drive/sharings/shared-with/{folderId}`                | Users with access to a folder        | No          |
| DELETE | `/drive/sharings/{itemType}/{itemId}/users/{userId}`    | Remove a user from a shared item     | No          |
| GET    | `/drive/sharings/public/{id}/folder/size`               | Get size of a public shared folder   | No          |

### Workspaces

| Method | Endpoint                                                                         | Description                       | Implemented |
| ------ | -------------------------------------------------------------------------------- | --------------------------------- | ----------- |
| GET    | `/drive/workspaces`                                                              | Get available workspaces          | No          |
| GET    | `/drive/workspaces/pending-setup`                                                | Get workspaces pending setup      | No          |
| GET    | `/drive/workspaces/invitations`                                                  | Get pending workspace invitations | No          |
| POST   | `/drive/workspaces/invitations/accept`                                           | Accept a workspace invitation     | No          |
| GET    | `/drive/workspaces/invitations/{inviteId}/validate`                              | Validate a workspace invitation   | No          |
| DELETE | `/drive/workspaces/invitations/{inviteId}`                                       | Decline a workspace invitation    | No          |
| GET    | `/drive/workspaces/teams/{teamId}/members`                                       | List team members                 | No          |
| PATCH  | `/drive/workspaces/teams/{teamId}`                                               | Edit team data                    | No          |
| DELETE | `/drive/workspaces/{workspaceId}/teams/{teamId}`                                 | Delete a team                     | No          |
| POST   | `/drive/workspaces/teams/{teamId}/user/{userUuid}`                               | Add user to team                  | No          |
| DELETE | `/drive/workspaces/teams/{teamId}/user/{userUuid}`                               | Remove user from team             | No          |
| GET    | `/drive/workspaces/{workspaceId}/files`                                          | List workspace files              | No          |
| POST   | `/drive/workspaces/{workspaceId}/files`                                          | Create workspace file             | No          |
| GET    | `/drive/workspaces/{workspaceId}/folders`                                        | List workspace folders            | No          |
| POST   | `/drive/workspaces/{workspaceId}/folders`                                        | Create workspace folder           | No          |
| PATCH  | `/drive/workspaces/{workspaceId}/teams/{teamId}/manager`                         | Change team manager               | No          |
| GET    | `/drive/workspaces/{workspaceId}/invitations`                                    | List workspace invitations        | No          |
| PATCH  | `/drive/workspaces/{workspaceId}/setup`                                          | Setup an initialized workspace    | No          |
| POST   | `/drive/workspaces/{workspaceId}/avatar`                                         | Upload workspace avatar           | No          |
| DELETE | `/drive/workspaces/{workspaceId}/avatar`                                         | Delete workspace avatar           | No          |
| GET    | `/drive/workspaces/{workspaceId}/credentials`                                    | Get workspace credentials         | No          |
| GET    | `/drive/workspaces/{workspaceId}/usage`                                          | Get workspace usage               | No          |
| PATCH  | `/drive/workspaces/{workspaceId}/members/{memberId}/usage`                       | Change member assigned space      | No          |
| GET    | `/drive/workspaces/{workspaceId}/members`                                        | List workspace members            | No          |
| POST   | `/drive/workspaces/{workspaceId}/members/invite`                                 | Invite user to workspace          | No          |
| POST   | `/drive/workspaces/{workspaceId}/teams`                                          | Create workspace team             | No          |
| GET    | `/drive/workspaces/{workspaceId}/teams`                                          | List workspace teams              | No          |
| GET    | `/drive/workspaces/{workspaceId}/usage/member`                                   | Get user usage in drive           | No          |
| POST   | `/drive/workspaces/{workspaceId}/shared`                                         | Share item to workspace           | No          |
| GET    | `/drive/workspaces/{workspaceId}/teams/{teamId}/shared/files`                    | List shared files in team         | No          |
| GET    | `/drive/workspaces/{workspaceId}/shared/files`                                   | List shared files                 | No          |
| GET    | `/drive/workspaces/{workspaceId}/teams/{teamId}/shared/folders`                  | List shared folders in team       | No          |
| GET    | `/drive/workspaces/{workspaceId}/shared/folders`                                 | List shared folders               | No          |
| GET    | `/drive/workspaces/{workspaceId}/teams/{teamId}/shared/{sharedFolderId}/folders` | List folders in a shared folder   | No          |
| GET    | `/drive/workspaces/{workspaceId}/shared/{sharedFolderId}/folders`                | List folders in a shared folder   | No          |
| GET    | `/drive/workspaces/{workspaceId}/teams/{teamId}/shared/{sharedFolderId}/files`   | List files in a shared folder     | No          |
| GET    | `/drive/workspaces/{workspaceId}/shared/{sharedFolderId}/files`                  | List files in a shared folder     | No          |
| GET    | `/drive/workspaces/{workspaceId}/shared/{itemType}/{itemId}/shared-with`         | List shares for an item           | No          |
| GET    | `/drive/workspaces/{workspaceId}/trash`                                          | Get workspace trash               | No          |
| DELETE | `/drive/workspaces/{workspaceId}/trash`                                          | Empty workspace trash             | No          |
| GET    | `/drive/workspaces/{workspaceId}/folders/{folderUuid}/folders`                   | List subfolders                   | No          |
| GET    | `/drive/workspaces/{workspaceId}/folders/{folderUuid}/files`                     | List files in a folder            | No          |
| PATCH  | `/drive/workspaces/{workspaceId}/teams/{teamId}/members/{memberId}/role`         | Change member’s role              | No          |
| PATCH  | `/drive/workspaces/{workspaceId}`                                                | Edit workspace details            | No          |
| GET    | `/drive/workspaces/{workspaceId}`                                                | Get workspace details             | No          |
| DELETE | `/drive/workspaces/{workspaceId}/members/leave`                                  | Leave workspace                   | No          |
| GET    | `/drive/workspaces/{workspaceId}/members/{memberId}`                             | Get member details                | No          |
| DELETE | `/drive/workspaces/{workspaceId}/members/{memberId}`                             | Remove member                     | No          |
| PATCH  | `/drive/workspaces/{workspaceId}/members/{memberId}/deactivate`                  | Deactivate a member               | No          |
| PATCH  | `/drive/workspaces/{workspaceId}/members/{memberId}/activate`                    | Activate a member                 | No          |

### User

| Method | Endpoint                                                   | Description                         | Implemented |
| ------ | ---------------------------------------------------------- | ----------------------------------- | ----------- |
| POST   | `/drive/users`                                             | Create a user                       | No          |
| GET    | `/drive/users/user/{email}`                                | Get user data & subscription status | No          |
| POST   | `/drive/users/pre-created-users/register`                  | Register a pre-created user         | No          |
| POST   | `/drive/users/pre-create`                                  | Pre-create a user                   | No          |
| GET    | `/drive/users/c/{uuid}`                                    | Get user credentials                | No          |
| GET    | `/drive/users/refresh`                                     | Refresh session token               | No          |
| PATCH  | `/drive/users/password`                                    | Change user password                | No          |
| POST   | `/drive/users/recover-account`                             | Request account recovery            | No          |
| PUT    | `/drive/users/recover-account`                             | Recover account                     | No          |
| POST   | `/drive/users/unblock-account`                             | Request account unblock             | No          |
| PUT    | `/drive/users/unblock-account`                             | Reset login error counter           | No          |
| GET    | `/drive/users/public-key/{email}`                          | Get public key by email             | No          |
| POST   | `/drive/users/attempt-change-email`                        | Initiate email change               | No          |
| POST   | `/drive/users/attempt-change-email/{id}/accept`            | Accept email change                 | No          |
| GET    | `/drive/users/attempt-change-email/{id}/verify-expiration` | Verify email-change link expiration | No          |
| GET    | `/drive/users/meet-token/beta`                             | Get Meet token (beta)               | No          |
| GET    | `/drive/users/meet-token/anon`                             | Get anonymous Meet token            | No          |
| POST   | `/drive/users/notification-token`                          | Add a notification token            | No          |
| POST   | `/drive/users/email-verification/send`                     | Send account verification email     | No          |
| POST   | `/drive/users/email-verification`                          | Verify user email                   | No          |
| PATCH  | `/drive/users/profile`                                     | Update user profile                 | No          |
| PUT    | `/drive/users/avatar`                                      | Upload user avatar                  | No          |
| DELETE | `/drive/users/avatar`                                      | Delete user avatar                  | No          |
| POST   | `/drive/users/deactivation/send`                           | Send deactivation email             | No          |
| POST   | `/drive/users/deactivation/confirm`                        | Confirm user deactivation           | No          |
| GET    | `/drive/users/usage`                                       | Get used storage space              | Yes         |
| GET    | `/drive/users/limit`                                       | Get maximum allowed storage         | Yes         |

### Fuzzy

| Method | Endpoint                | Description                  | Implemented |
| ------ | ----------------------- | ---------------------------- | ----------- |
| GET    | `/drive/fuzzy/{search}` | Search items by partial name | Yes         |

### Backup

| Method | Endpoint                                | Description                             | Implemented |
| ------ | --------------------------------------- | --------------------------------------- | ----------- |
| POST   | `/drive/backup/activate`                | Activate user backup                    | No          |
| POST   | `/drive/backup/deviceAsFolder`          | Create device as folder                 | No          |
| GET    | `/drive/backup/deviceAsFolder`          | List device-as-folders                  | No          |
| GET    | `/drive/backup/deviceAsFolder/{uuid}`   | Get device-as-folder by UUID            | No          |
| PATCH  | `/drive/backup/deviceAsFolder/{uuid}`   | Update device-as-folder                 | No          |
| GET    | `/drive/backup/deviceAsFolderById/{id}` | Get device-as-folder by ID (deprecated) | No          |
| GET    | `/drive/backup/devices`                 | List user devices                       | No          |
| DELETE | `/drive/backup/devices/{deviceId}`      | Delete a user device                    | No          |
| GET    | `/drive/backup/{mac}`                   | Get backups for a device                | No          |
| DELETE | `/drive/backup/{backupId}`              | Delete a backup                         | No          |

### Trash

| Method | Endpoint                                 | Description                              | Implemented |
| ------ | ---------------------------------------- | ---------------------------------------- | ----------- |
| GET    | `/drive/storage/trash/paginated`         | Get paginated trash content              | Yes         |
| POST   | `/drive/storage/trash/add`               | Add files/folders to trash               | Yes         |
| DELETE | `/drive/storage/trash/all`               | Delete all items from user’s trash       | Yes         |
| DELETE | `/drive/storage/trash/all/request`       | (alternate) Delete all trash via request | Yes         |
| DELETE | `/drive/storage/trash`                   | Delete specified trash items             | Yes         |
| DELETE | `/drive/storage/trash/file/{fileId}`     | Delete single file from trash            | Yes         |
| DELETE | `/drive/storage/trash/folder/{folderId}` | Delete single folder from trash          | Yes         |

### Sends

| Method | Endpoint                | Description         | Implemented |
| ------ | ----------------------- | ------------------- | ----------- |
| POST   | `/drive/links`          | Create a send link  | No          |
| GET    | `/drive/links/{linkId}` | Get send link by ID | No          |

### Device

| Method | Endpoint                    | Description           | Implemented |
| ------ | --------------------------- | --------------------- | ----------- |
| POST   | `/drive/device/context`     | Get device context    | No          |
| POST   | `/drive/device/geolocation` | Get geolocation by IP | No          |

### Gateway

| Method | Endpoint                                                        | Description                         | Implemented |
| ------ | --------------------------------------------------------------- | ----------------------------------- | ----------- |
| POST   | `/drive/gateway/workspaces`                                     | Initiate a workspace                | No          |
| DELETE | `/drive/gateway/workspaces`                                     | Destroy a workspace                 | No          |
| PUT    | `/drive/gateway/workspaces/storage`                             | Update a workspace                  | No          |
| POST   | `/drive/gateway/workspaces/{workspaceId}/storage/upgrade-check` | Pre-check workspace storage upgrade | No          |
| GET    | `/drive/gateway/users`                                          | Get gateway user details            | No          |
| GET    | `/drive/gateway/users/storage/stackability`                     | Check if user can expand storage    | No          |
| PATCH  | `/drive/gateway/users/{uuid}`                                   | Update user plan and storage        | No          |
