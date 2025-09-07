package internxtclient_test

import (
	"strings"
	"testing"
	"time"

	client "github.com/StarHack/go-internxt-drive/internxtclient"
)

func TestTrashIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	trashFile1CreateResponse := createFile(t, "trash_file1", testFolderUUID)
	trashFile2CreateResponse := createFile(t, "trash_file2", testFolderUUID)

	trashFolder1 := createFolder(t, "trash_folder1", testFolderUUID)
	trashFolder2 := createFolder(t, "trash_folder2", testFolderUUID)

	time.Sleep(1 * time.Second)

	trashFile1 := getFileMeta(t, trashFile1CreateResponse.UUID)
	trashFile2 := getFileMeta(t, trashFile2CreateResponse.UUID)

	time.Sleep(1 * time.Second)

	t.Run("AddToTrash", func(t *testing.T) {
		fileTrashRefs := c.Trash.FilesToTrashRefs([]client.File{*trashFile1, *trashFile2})
		folderTrashRefs := c.Trash.FoldersToTrashRefs([]client.Folder{*trashFolder1, *trashFolder2})
		trashRefs := append(fileTrashRefs, folderTrashRefs...)
		addToTrash(t, trashRefs)
		time.Sleep(1 * time.Second)

		trashedTrashFolder1 := getFolderMeta(t, trashFolder1.UUID)
		trashedTrashFolder2 := getFolderMeta(t, trashFolder2.UUID)
		trashedTrashFile1 := getFileMeta(t, trashFile1.UUID)
		trashedTrashFile2 := getFileMeta(t, trashFile2.UUID)

		if trashedTrashFolder1.Status != "TRASHED" {
			t.Fatalf("AddToTrash for trashedTrashFolder failed")
		}
		if trashedTrashFolder2.Status != "TRASHED" {
			t.Fatalf("AddToTrash for trashedTrashFolder failed")
		}

		if trashedTrashFile1.Status != "TRASHED" {
			t.Fatalf("AddToTrash for trashedTrashFile failed")
		}
		if trashedTrashFile2.Status != "TRASHED" {
			t.Fatalf("AddToTrash for trashedTrashFile failed")
		}
	})

	time.Sleep(1 * time.Second)

	t.Run("DeleteTrashFile", func(t *testing.T) {
		deleteTrashFile(t, trashFile1.FileID)

		time.Sleep(1 * time.Second)

		ensureFileIsDeleted(t, trashFile1.UUID)
	})

	t.Run("DeleteTrashFolder", func(t *testing.T) {
		deleteTrashFolder(t, trashFolder1.ID)

		time.Sleep(1 * time.Second)

		ensureFolderIsDeleted(t, trashFolder1.UUID)
	})

	t.Run("DeleteSpecidiedTrashItems", func(t *testing.T) {
		fileTrashRefs := c.Trash.FilesToTrashRefs([]client.File{*trashFile2})
		folderTrashRefs := c.Trash.FoldersToTrashRefs([]client.Folder{*trashFolder2})
		trashRefs := append(fileTrashRefs, folderTrashRefs...)
		deleteSpecidiedTrashItems(t, trashRefs)

		time.Sleep(1 * time.Second)

		ensureFileIsDeleted(t, trashFile2.UUID)
		ensureFolderIsDeleted(t, trashFolder2.UUID)
	})

}

func ensureFileIsDeleted(t *testing.T, fileUUID string) {
	deletedFileMeta := getFileMeta(t, fileUUID)
	if deletedFileMeta == nil {
		t.Fatalf("DeleteSpecidiedTrashItems failed: couldn't get deleted file meta")
	}

	if deletedFileMeta.Status != "DELETED" {
		t.Fatalf("DeleteSpecidiedTrashItems failed: file status isn't DELETED")
	}
}

func ensureFolderIsDeleted(t *testing.T, folderUUID string) {
	deletedTrashFolder, err := c.Folders.GetFolderMeta(folderUUID)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			t.Fatalf("DeleteTrashFolder failed: %v", err)
		}
	}
	if deletedTrashFolder != nil {
		t.Fatalf("DeleteTrashFolder failed: got back a folder when we should have got nil")
	}
}

func deleteSpecidiedTrashItems(t *testing.T, trashRefs []client.TrashRef) {
	err := c.Trash.DeleteSpecifiedTrashItems(trashRefs)
	if err != nil {
		t.Fatalf("DeleteSpecidiedTrashItems failed: %v", err)
	}
}

func deleteTrashFile(t *testing.T, fileID string) {
	err := c.Trash.DeleteTrashFile(fileID)
	if err != nil {
		t.Fatalf("DeleteTrashFile failed: %v", err)
	}
}

func deleteTrashFolder(t *testing.T, folderID int64) {
	err := c.Trash.DeleteTrashFolder(folderID)
	if err != nil {
		t.Fatalf("DeleteTrashFolder failed: %v", err)
	}
}

func deleteSpecifiedTrashItems(t *testing.T, trashRefs []client.TrashRef) {
	err := c.Trash.DeleteSpecifiedTrashItems(trashRefs)
	if err != nil {
		t.Fatalf("DeleteSpecifiedTrashItems failed: %v", err)
	}
}

func addToTrash(t *testing.T, trashRefs []client.TrashRef) {
	err := c.Trash.AddToTrash(trashRefs)
	if err != nil {
		t.Fatalf("AddToTrash failed: %v", err)
	}
}
