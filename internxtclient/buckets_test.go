package internxtclient_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/StarHack/go-internxt-drive/internxtclient"
)

func TestBucketsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var bucketsFile *internxtclient.CreateMetaResponse

	t.Run("CreateFile", func(t *testing.T) {
		bucketsFile = createFile(t, "buckets_file", testFolderUUID)
	})

	time.Sleep(1 * time.Second)
	deleteFile(t, bucketsFile.UUID)
	time.Sleep(1 * time.Second)
}

func createFile(t *testing.T, filename, destFolderUUID string) *internxtclient.CreateMetaResponse {
	createMetaResponse, err := c.Buckets.UploadFileStream(destFolderUUID, filename, bytes.NewReader(testBytes), int64(len(testBytes)), time.Now())
	if err != nil {
		t.Fatalf("couldn't upload filestream: %v", err)
	}
	if createMetaResponse == nil {
		t.Fatalf("createMetaResponse is nil")
	}
	if createMetaResponse.Bucket != c.UserData.AccessData.User.Bucket {
		t.Fatalf("createMetaResponse.Bucket is not the same as user's bucket. User's bucket is %s, but got %s", createMetaResponse.Bucket, c.UserData.AccessData.User.Bucket)
	}

	return createMetaResponse
}
