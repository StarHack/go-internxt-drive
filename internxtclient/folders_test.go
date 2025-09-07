package internxtclient_test

import (
	"testing"
	"time"

	"github.com/StarHack/go-internxt-drive/internxtclient"
)

func TestFoldersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var subFolder1UUID string
	var subFolder2UUID string

	foldersSubfolder1 := "go-internxt-drive-subfolder-1"
	foldersSubfolder2 := "go-internxt-drive-subfolder-2"

	t.Run("CreateFolder", func(t *testing.T) {
		folder1 := createFolder(t, foldersSubfolder1, testFolderUUID)
		folder2 := createFolder(t, foldersSubfolder2, testFolderUUID)
		subFolder1UUID = folder1.UUID
		subFolder2UUID = folder2.UUID
	})

	time.Sleep(1 * time.Second)
	t.Logf("Created two folders:\n%s -> %s\n%s -> %s", foldersSubfolder1, subFolder1UUID, foldersSubfolder2, subFolder2UUID)

	t.Run("GetFolderMeta", func(t *testing.T) {
		getFolderMeta(t, subFolder1UUID)
	})

	t.Run("RenameFolder", func(t *testing.T) {
		renameFolder(t, subFolder1UUID)
	})

	t.Run("MoveFolder", func(t *testing.T) {
		moveFolder(t, subFolder2UUID, subFolder1UUID)
	})

	t.Run("GetFolderSize", func(t *testing.T) {
		getFolderSize(t, 0)
	})

	createFile(t, "1", subFolder1UUID)
	createFile(t, "2", subFolder2UUID)

	/*
		We should now have this folder structure
		RootFolder
		└── go-internxt-drive-tests					<- Base testing folder: testFolderUUID
			└── renamed								<- subFolder1
				├── 1
				└── go-internxt-drive-subfolder-2	<- subFolder2
					└── 2
	*/

	time.Sleep(1 * time.Second)

	t.Run("Tree", func(t *testing.T) {
		folder := tree(t, testFolderUUID)

		if folder == nil {
			t.Fatalf("folder is nil")
		}
		if folder.Children == nil {
			t.Fatalf("folder.Children is nil")
		}
		if folder.Files == nil {
			t.Fatalf("folder.Files is nil")
		}
		if len(folder.Files) != 0 {
			t.Fatalf("folder.Files should be 0, but is %d", len(folder.Files))
		}
		if len(folder.Children) != 1 {
			t.Fatalf("folder.Children should be 1, but is %d", len(folder.Children))
		}

		if folder.Children[0].PlainName != "renamed" {
			t.Fatalf("folder.Children[0]'s name should be 'renamed' but is %s", folder.Children[0].PlainName)
		}
		if folder.Children[0].Children[0].PlainName != foldersSubfolder2 {
			t.Fatalf("folder.Children[0].Children[0]'s name should be '%s' but is %s", foldersSubfolder2, folder.Children[0].Children[0].PlainName)
		}
		if folder.Children[0].Files[0].PlainName != "1" {
			t.Fatalf("folder.Children[0].Files[0]'s name should be '1' but is %s", folder.Children[0].Files[0].PlainName)
		}
		if folder.Children[0].Children[0].Files[0].PlainName != "2" {
			t.Fatalf("folder.Children[0].Children[0].Files[0]'s name should be '2' but is %s", folder.Children[0].Children[0].Files[0].PlainName)
		}
	})

	t.Run("ListFiles", func(t *testing.T) {
		files, err := c.Folders.ListFiles(subFolder1UUID, &internxtclient.ListOptions{Limit: 1, Offset: 0})
		if err != nil {
			t.Fatalf("Couldn't ListFiles:  %v", err)
		}
		if files == nil {
			t.Fatalf("files is nil")

		}
		if len(files) != 1 {
			t.Fatalf("Listed wrong amount of files")
		}
		if files[0].PlainName != "1" {
			t.Fatalf("Listed wrong file: %s, expected '1'", files[0].PlainName)
		}
	})

	t.Run("ListFolders", func(t *testing.T) {
		folders, err := c.Folders.ListFolders(subFolder1UUID, &internxtclient.ListOptions{Limit: 1, Offset: 0})
		if err != nil {
			t.Fatalf("Couldn't ListFolders:  %v", err)
		}
		if folders == nil {
			t.Fatalf("files is nil")

		}
		if len(folders) != 1 {
			t.Fatalf("Listed wrong amount of folders")
		}
		if folders[0].PlainName != foldersSubfolder2 {
			t.Fatalf("Listed wrong folder: %s, expected '%s'", folders[0].PlainName, foldersSubfolder2)
		}
	})

	t.Run("DeleteFolder", func(t *testing.T) {
		deleteFolder(t, subFolder1UUID)
	})

}

func createFolder(t *testing.T, folderName, parentFolderUUID string) *internxtclient.Folder {
	folder, err := c.Folders.CreateFolder(internxtclient.CreateFolderRequest{PlainName: folderName, ParentFolderUUID: parentFolderUUID})
	if err != nil {
		t.Fatalf("Couldn't create folder:  %v", err)
	}
	if folder == nil {
		t.Fatalf("Couldn't create folder: folder is nil")
	}
	if folder.UUID == "" {
		t.Fatalf("Couldn't create folder: folder has no UUID")
	}
	return folder
}

/*
func createFolder(t *testing.T) (string, string) {
	folder1, err := c.Folders.CreateFolder(internxtclient.CreateFolderRequest{PlainName: TESTSUBFOLDER1, ParentFolderUUID: testFolderUUID})
	if err != nil {
		t.Fatalf("Couldn't create folder:  %v", err)
	}
	if folder1 == nil {
		t.Fatalf("Couldn't create folder: folder is nil")
	}
	if folder1.UUID == "" {
		t.Fatalf("Couldn't create folder: folder has no UUID")
	}

	folder2, err := c.Folders.CreateFolder(internxtclient.CreateFolderRequest{PlainName: TESTSUBFOLDER2, ParentFolderUUID: testFolderUUID})
	if err != nil {
		t.Fatalf("Couldn't create folder:  %v", err)
	}
	if folder2 == nil {
		t.Fatalf("Couldn't create folder: folder is nil")
	}
	if folder2.UUID == "" {
		t.Fatalf("Couldn't create folder: folder has no UUID")
	}

	return folder1.UUID, folder2.UUID
}
*/

func getFolderSize(t *testing.T, shouldBe int64) {
	s, err := c.Folders.GetFolderSize(testFolderUUID)
	if err != nil {
		t.Fatalf("Couldn't get folder size: %v", err)
	}
	if s != shouldBe {
		t.Fatalf("Folder size should be %d, but got %d", shouldBe, s)
	}
}

func renameFolder(t *testing.T, folderUUID string) {
	newName := "renamed"
	err := c.Folders.RenameFolder(folderUUID, newName)
	if err != nil {
		t.Fatalf("Couldn't rename folder: %v", err)
	}
	time.Sleep(1 * time.Second)
	f := getFolderMeta(t, folderUUID)
	if f == nil {
		t.Fatalf("Couldn't get meta of renamed folder")
	}
	if f.PlainName != newName {
		t.Fatalf("Renamed folder does not have the correct new name. Should be %s but is %s", newName, f.PlainName)
	}
}

func getFolderMeta(t *testing.T, folderUUID string) *internxtclient.Folder {
	folder, err := c.Folders.GetFolderMeta(folderUUID)
	if err != nil {
		t.Fatalf("Couldn't get folder meta: %v", err)
	}
	if folder == nil {
		t.Fatalf("Folder is nil")
	}
	if folder.UUID == "" {
		t.Fatalf("Folder has no UUID")
	}

	return folder
}

func moveFolder(t *testing.T, folderUUID, destUUID string) {
	err := c.Folders.MoveFolder(folderUUID, destUUID)
	if err != nil {
		t.Fatalf("Couldn't move folder: %v", err)
	}

	time.Sleep(1 * time.Second)

	f := getFolderMeta(t, folderUUID)
	if f == nil {
		t.Fatalf("Couldn't get meta of renamed folder")
	}
	if f.ParentUUID != destUUID {
		t.Fatalf("Moved folder does not have the correct new parent. Should be %s but is %s", destUUID, f.ParentUUID)
	}
}

func tree(t *testing.T, folderUUID string) *internxtclient.Folder {
	folder, err := c.Folders.Tree(folderUUID)
	if err != nil {
		t.Fatalf("Couldn't delete folder:  %v", err)
	}
	if folder == nil {
		t.Fatalf("Folder is nil")
	}
	if folder.UUID == "" {
		t.Fatalf("Folder has no UUID")
	}

	return folder
}

func deleteFolder(t *testing.T, folderUUID string) {
	err := c.Folders.DeleteFolder(folderUUID)
	if err != nil {
		t.Fatalf("Couldn't delete folder:  %v", err)
	}
}
