package fuzzy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/internxt/rclone/config"
)

type SearchResult struct {
	ID         string  `json:"id"`
	ItemID     string  `json:"itemId"`
	ItemType   string  `json:"itemType"` // "file" or "folder"
	Name       string  `json:"name"`
	Rank       int     `json:"rank"`
	Similarity float64 `json:"similarity"`
}

type SearchResponse struct {
	Data []SearchResult `json:"data"`
}

// FuzzySearch performs a fuzzy search with a given term and offset.
func FuzzySearch(cfg *config.Config, term string, offset int) (*SearchResponse, error) {
	encodedTerm := url.PathEscape(term)
	endpoint := fmt.Sprintf("%s/fuzzy/%s?offset=%d", cfg.DriveAPIURL, encodedTerm, offset)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
