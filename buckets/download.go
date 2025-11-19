package buckets

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/internxt/rclone/config"
)

const bucketAPIBase = "https://api.internxt.com"

// ShardInfo mirrors the per‑shard info returned by /files/{fileID}/info
type ShardInfo struct {
	Index int    `json:"index"`
	Hash  string `json:"hash"`
	URL   string `json:"url"`
}

// BucketFileInfo is the metadata returned by GET /buckets/{bucketID}/files/{fileID}/info
type BucketFileInfo struct {
	Bucket   string      `json:"bucket"`
	Index    string      `json:"index"`
	Size     int64       `json:"size"`
	Version  int         `json:"version"`
	Created  string      `json:"created"`
	Renewal  string      `json:"renewal"`
	Mimetype string      `json:"mimetype"`
	Filename string      `json:"filename"`
	ID       string      `json:"id"`
	Shards   []ShardInfo `json:"shards"`
}

// GetBucketFileInfo calls the correct /info endpoint and parses its JSON.
func GetBucketFileInfo(cfg *config.Config, bucketID, fileID string) (*BucketFileInfo, error) {
	url := fmt.Sprintf("%s/buckets/%s/files/%s/info", bucketAPIBase, bucketID, fileID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", cfg.BasicAuthHeader)
	req.Header.Set("internxt-version", "1.0")
	req.Header.Set("internxt-client", "internxt-go-sdk")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("file info fetch failed: %d %s", resp.StatusCode, string(body))
	}

	var info BucketFileInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// DownloadFile downloads and decrypts the first shard of the given file.
func DownloadFile(cfg *config.Config, fileID, destPath string) error {
	// 1) fetch file info from the bucket API
	info, err := GetBucketFileInfo(cfg, cfg.Bucket, fileID)
	if err != nil {
		return err
	}
	if len(info.Shards) == 0 {
		return fmt.Errorf("no shards found for file %s", fileID)
	}
	shard := info.Shards[0]

	// 2) derive fileKey+iv using the stored index (hex of random index)
	key, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, info.Index)
	if err != nil {
		return err
	}

	// 3) GET the encrypted shard directly from its presigned URL
	req, err := http.NewRequest("GET", shard.URL, nil)
	if err != nil {
		return err
	}
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shard download failed: %d %s", resp.StatusCode, string(body))
	}

	// 4) wrap in AES‑CTR decryptor
	decReader, err := DecryptReader(resp.Body, key, iv)
	if err != nil {
		return err
	}

	// 5) write plaintext to file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, decReader); err != nil {
		return err
	}
	return nil
}

// DownloadFileStream returns a ReadCloser that streams the decrypted contents
// of the file with the given UUID. The caller must close the returned ReadCloser.
// It takes an optional range header in the format of either "bytes=100-199" or "bytes=100-".
func DownloadFileStream(cfg *config.Config, fileUUID string, optionalRange ...string) (io.ReadCloser, error) {
	rangeValue := ""
	if len(optionalRange) > 0 {
		rangeValue = optionalRange[0]
	}

	// 1) Fetch file info (including shards and index)
	info, err := GetBucketFileInfo(cfg, cfg.Bucket, fileUUID)
	if err != nil {
		return nil, err
	}
	if len(info.Shards) == 0 {
		return nil, fmt.Errorf("no shards found for file %s", fileUUID)
	}
	shard := info.Shards[0]

	// 2) Derive fileKey and IV from the stored index
	key, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, info.Index)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file key: %w", err)
	}

	// 3) Calculate the IV for the requested range
	if rangeValue != "" {
		startByte, endByte, err := getStartByteAndEndByte(rangeValue)
		if err != nil {
			return nil, fmt.Errorf("invalid range: %w", err)
		}

		// Ensure AES block alignment for correct decryption
		// Find the nearest block and call this function again with the adjusted range, then discard the unwanted bytes before returning
		if offset := startByte % 16; offset != 0 {
			alignedStart := startByte - offset
			var adjustedRange string
			if endByte == -1 {
				adjustedRange = fmt.Sprintf("bytes=%d-", alignedStart)
			} else {
				adjustedRange = fmt.Sprintf("bytes=%d-%d", alignedStart, endByte)
			}

			stream, err := DownloadFileStream(cfg, fileUUID, adjustedRange)
			if err != nil {
				return nil, err
			}

			// Discard unwanted bytes and return the requested range exactly
			if _, err := io.CopyN(io.Discard, stream, int64(offset)); err != nil {
				stream.Close()
				return nil, fmt.Errorf("failed to discard offset bytes: %w", err)
			}
			return stream, nil
		}

		adjustIV(iv, startByte/16)
	}

	// 4) Download the encrypted shard, include the Range header if any
	req, err := http.NewRequest("GET", shard.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if rangeValue != "" {
		req.Header.Set("Range", rangeValue)
	}

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("shard download failed: %d %s", resp.StatusCode, string(body))
	}

	// 5) Wrap in AES‑CTR decryptor
	decReader, err := DecryptReader(resp.Body, key, iv)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	// 6) Return a ReadCloser that closes the HTTP body when closed
	return struct {
		io.Reader
		io.Closer
	}{Reader: decReader, Closer: resp.Body}, nil
}

// This will return the startByte and endByte of a range header in these formats: "bytes=100-199" or "bytes=100-"
// In the case of the "bytes=100-" the returned endByte will be -1.
// Formats like "bytes=-200" and "bytes=0-99,200-299" are not supported.
func getStartByteAndEndByte(rangeHeader string) (int, int, error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, fmt.Errorf("invalid Range header format")
	}

	rangePart := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangePart, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid Range header format")
	}

	startByte, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start byte in Range header: %w", err)
	}

	// Handle optional endByte
	if parts[1] == "" {
		return startByte, -1, nil
	}

	endByte, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end byte in Range header: %w", err)
	}

	return startByte, endByte, nil
}

// adjustIV increments the IV based on the given block index.
func adjustIV(iv []byte, blockIndex int) {
	for i := 0; i < blockIndex; i++ {
		for j := len(iv) - 1; j >= 0; j-- {
			iv[j]++
			if iv[j] != 0 {
				break
			}
		}
	}
}
