package internxtclient

import "testing"

func TestShouldRetryParentFolderMissing404(t *testing.T) {
	body := []byte(`{"message":"Parent folder does not exist","error":"Not Found","statusCode":404}`)
	if !shouldRetryResponse(404, body) {
		t.Fatal("expected parent-folder 404 to be retryable")
	}
}

func TestShouldNotRetryOrdinary404(t *testing.T) {
	body := []byte(`{"message":"Folder not found","error":"Not Found","statusCode":404}`)
	if shouldRetryResponse(404, body) {
		t.Fatal("expected ordinary 404 to be non-retryable")
	}
}
