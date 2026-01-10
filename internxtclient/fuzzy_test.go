package internxtclient_test

import (
	"testing"
	"time"

	client "github.com/StarHack/go-internxt-drive/internxtclient"
)

func TestFuzzyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createMetaResponse := createFile(t, "fuzzy_file", testFolderUUID)

	time.Sleep(1 * time.Second)

	t.Run("FuzzySearch", func(t *testing.T) {
		searchResponse := fuzzySearch(t, "fuzzy_file")

		found := false
		for _, result := range searchResponse.Data {
			if result.Name == "fuzzy_file" {
				if result.ItemID == createMetaResponse.UUID {
					found = true
					break
				}
			}
		}
		if !found {
			t.Fatalf("FuzzySearch did not find the file")
		}
	})
}

func fuzzySearch(t *testing.T, term string) *client.SearchResponse {
	searchResponse, err := c.Fuzzy.FuzzySearch(term, 0)
	if err != nil {
		t.Fatalf("FuzzySearch failed: %v", err)
	}
	if searchResponse == nil {
		t.Fatalf("FuzzySearch returned nil")
	}
	if len(searchResponse.Data) == 0 {
		t.Fatalf("FuzzySearch returned no results")
	}

	return searchResponse
}
