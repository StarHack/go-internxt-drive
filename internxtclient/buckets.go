package internxtclient

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

type BucketsService struct {
	client *Client
}

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

// UploadPartSpec defines each part’s index and size for the start call
type UploadPartSpec struct {
	Index int   `json:"index"`
	Size  int64 `json:"size"`
}

type startUploadReq struct {
	Uploads []UploadPartSpec `json:"uploads"`
}

type UploadPart struct {
	Index int    `json:"index"`
	UUID  string `json:"uuid"`
	URL   string `json:"url"`
}

type StartUploadResp struct {
	Uploads []UploadPart `json:"uploads"`
}

type CreateMetaRequest struct {
	Name             string    `json:"name"`
	Bucket           string    `json:"bucket"`
	FileID           string    `json:"fileId"`
	EncryptVersion   string    `json:"encryptVersion"`
	FolderUuid       string    `json:"folderUuid"`
	Size             int64     `json:"size"`
	PlainName        string    `json:"plainName"`
	Type             string    `json:"type"`
	CreationTime     time.Time `json:"creationTime"`
	Date             time.Time `json:"date"`
	ModificationTime time.Time `json:"modificationTime"`
}

type CreateMetaResponse struct {
	UUID           string      `json:"uuid"`
	Name           string      `json:"name"`
	Bucket         string      `json:"bucket"`
	FileID         string      `json:"fileId"`
	EncryptVersion string      `json:"encryptVersion"`
	FolderUuid     string      `json:"folderUuid"`
	Size           json.Number `json:"size"`
	PlainName      string      `json:"plainName"`
	Type           string      `json:"type"`
	Created        string      `json:"created"`
}

type Shard struct {
	Hash string `json:"hash"`
	UUID string `json:"uuid"`
}

type FinishUploadResp struct {
	Bucket   string `json:"bucket"`
	Index    string `json:"index"`
	ID       string `json:"id"`
	Version  int    `json:"version"`
	Created  string `json:"created"`
	Renewal  string `json:"renewal"`
	Mimetype string `json:"mimetype"`
	Filename string `json:"filename"`
}

// GetBucketFileInfo calls the correct /info endpoint and parses its JSON.
func (b *BucketsService) GetBucketFileInfo(bucketID, fileID string) (*BucketFileInfo, error) {
	if !b.client.hasUserData() {
		return nil, fmt.Errorf("can't get bucket file info, client has no user data")
	}

	endpoint := path.Join(bucketID, "files", fileID, "info")

	var info BucketFileInfo

	headers := http.Header{}
	headers.Set("Authorization", b.client.UserData.BasicAuthHeader)

	if resp, err := b.client.Get(APITypeBucket, endpoint, &info, &headers); err != nil {
		return nil, b.client.GetError(endpoint, resp, err)
	}

	return &info, nil
}

// Downloads a file by its UUID and places it at destination
func (b *BucketsService) DownloadFile(fileUUID, destination string) error {
	readCloser, err := b.DownloadFileStream(fileUUID)
	if err != nil {
		return b.client.GetError("", nil, err)
	}

	defer readCloser.Close()

	outFile, err := os.Create(destination)
	if err != nil {
		return b.client.GetError("", nil, err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, readCloser)
	if err != nil {
		return b.client.GetError("", nil, err)
	}

	return nil
}

// DownloadFileStream returns a ReadCloser that streams the decrypted contents
// of the file with the given UUID. The caller must close the returned ReadCloser.
// It takes an optional range header in the format of either "bytes=100-199" or "bytes=100-".
func (b *BucketsService) DownloadFileStream(fileUUID string, optionalRange ...string) (io.ReadCloser, error) {
	if !b.client.hasUserDataAccessDataUser() {
		return nil, fmt.Errorf("no user data available when downloading file %s", fileUUID)
	}

	rangeValue := ""
	if len(optionalRange) > 0 {
		rangeValue = optionalRange[0]
	}

	// 1) Fetch file info (including shards and index)
	info, err := b.GetBucketFileInfo(b.client.UserData.AccessData.User.Bucket, fileUUID)
	if err != nil {
		return nil, err
	}
	if len(info.Shards) == 0 {
		return nil, fmt.Errorf("no shards found for file %s", fileUUID)
	}
	shard := info.Shards[0]

	// 2) Derive fileKey and IV from the stored index
	key, iv, err := GenerateFileKey(b.client.UserData.AccessData.User.Mnemonic, b.client.UserData.AccessData.User.Bucket, info.Index)
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

			stream, err := b.DownloadFileStream(fileUUID, adjustedRange)
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

	resp, err := b.client.HTTPClient.Do(req)
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

// UploadFileStream uploads data from the provided io.Reader into Internxt,
// encrypting it on the fly and creating the metadata file in the target folder.
// It returns the CreateMetaResponse of the created file entry.
func (b *BucketsService) UploadFileStream(targetFolderUUID, fileName string, in io.Reader, plainSize int64, modTime time.Time) (*CreateMetaResponse, error) {
	if !b.client.hasUserDataAccessDataUser() {
		return nil, fmt.Errorf("no user data available when uploading file %s", fileName)
	}

	var ph [32]byte
	if _, err := rand.Read(ph[:]); err != nil {
		return nil, fmt.Errorf("cannot generate random index: %w", err)
	}
	plainIndex := hex.EncodeToString(ph[:])

	fileKey, iv, err := GenerateFileKey(b.client.UserData.AccessData.User.Mnemonic, b.client.UserData.AccessData.User.Bucket, plainIndex)
	if err != nil {
		return nil, fmt.Errorf("can't generate file key: %w", err)
	}

	encReader, err := EncryptReader(in, fileKey, iv)
	if err != nil {
		return nil, fmt.Errorf("can't create EncryptReader: %w", err)
	}

	sha256Hasher := sha256.New()
	sha1Hasher := sha1.New()
	r := io.TeeReader(encReader, sha256Hasher)
	r = io.TeeReader(r, sha1Hasher)

	specs := []UploadPartSpec{{Index: 0, Size: plainSize}}
	startResp, err := b.StartUpload(b.client.UserData.AccessData.User.Bucket, specs)
	if err != nil {
		return nil, err
	}

	if len(startResp.Uploads) == 0 {
		return nil, fmt.Errorf("startResp.Uploads is empty")
	}

	part := startResp.Uploads[0]

	if err := b.Transfer(part, r, plainSize); err != nil {
		return nil, err
	}

	encIndex := hex.EncodeToString(ph[:])
	partHash := hex.EncodeToString(sha1Hasher.Sum(nil))
	finishResp, err := b.FinishUpload(encIndex, []Shard{{Hash: partHash, UUID: part.UUID}})
	if err != nil {
		return nil, err
	}

	base := filepath.Base(fileName)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	ext := strings.TrimPrefix(filepath.Ext(base), ".")
	meta, err := b.CreateMetaFile(name, finishResp.ID, "03-aes", targetFolderUUID, name, ext, plainSize, modTime)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (b *BucketsService) Transfer(part UploadPart, r io.Reader, size int64) error {
	req, err := http.NewRequest("PUT", part.URL, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size
	resp, err := b.client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("transfer failed: status %d, %s", resp.StatusCode, string(body))
	}
	return nil
}

// StartUpload reserves all parts at once
func (b *BucketsService) StartUpload(bucketID string, parts []UploadPartSpec) (*StartUploadResp, error) {
	if !b.client.hasUserData() {
		return nil, fmt.Errorf("can't start upload without user data")
	}

	endpoint := path.Join("v2", "buckets", bucketID, "files", "start")
	reqBody := startUploadReq{Uploads: parts}

	headers := http.Header{}
	headers.Set("Authorization", b.client.UserData.BasicAuthHeader)

	var result StartUploadResp

	if resp, err := b.client.doRequestWithQuery(APITypeBase, http.MethodPost, endpoint, map[string]string{"multiparts": "1"}, &reqBody, &result, &headers); err != nil {
		return nil, b.client.GetError(endpoint, resp, err)
	}

	return &result, nil
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

// NewAES256CTRCipher returns a cipher.Stream that performs AES‑256‑CTR encryption
// with the given 32‑byte key and 16‑byte IV, exactly like Node.js’s
// createCipheriv('aes-256-ctr', key, iv).
func NewAES256CTRCipher(key, iv []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewCTR(block, iv), nil
}

// EncryptReader wraps the provided src reader in a StreamReader that
// encrypts all data through AES‑256‑CTR (no padding):
//
//	source -> cipher -> …
func EncryptReader(src io.Reader, key, iv []byte) (io.Reader, error) {
	stream, err := NewAES256CTRCipher(key, iv)
	if err != nil {
		return nil, err
	}
	return cipher.StreamReader{S: stream, R: src}, nil
}

// DecryptReader wraps the provided src reader in a StreamReader that
// decrypts data encrypted with AES‑256‑CTR (no padding):
//
//	encryptedSrc -> source -> …
func DecryptReader(src io.Reader, key, iv []byte) (io.Reader, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, iv)
	return cipher.StreamReader{S: stream, R: src}, nil
}

// GetFileDeterministicKey returns SHA512(key||data)
func GetFileDeterministicKey(key, data []byte) []byte {
	h := sha512.New()
	h.Write(key)
	h.Write(data)
	return h.Sum(nil)
}

// GenerateFileBucketKey derives a bucket-level key from mnemonic and bucketID
func GenerateFileBucketKey(mnemonic, bucketID string) ([]byte, error) {
	seed := bip39.NewSeed(mnemonic, "")
	bucketBytes, err := hex.DecodeString(bucketID)
	if err != nil {
		return nil, err
	}

	return GetFileDeterministicKey(seed, bucketBytes), nil
}

// GenerateBucketKey generates a 64-character hexadecimal bucket key from a mnemonic and bucket ID.
func GenerateBucketKey(mnem string, bucketID []byte) (string, error) {
	seed := bip39.NewSeed(mnem, "")
	deterministicKey, err := GetDeterministicKey(seed, bucketID)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(deterministicKey)[:64], nil
}

func GetDeterministicKey(key []byte, data []byte) ([]byte, error) {
	hasher := sha512.New()
	hasher.Write(key)
	hasher.Write(data)
	return hasher.Sum(nil), nil
}

// GenerateFileKey derives the per-file key and IV from mnemonic, bucketID, and plaintext index
func GenerateFileKey(mnemonic, bucketID, indexHex string) (key, iv []byte, err error) {
	bucketKey, err := GenerateFileBucketKey(mnemonic, bucketID)
	if err != nil {
		return nil, nil, err
	}
	indexBytes, err := hex.DecodeString(indexHex)
	if err != nil {
		return nil, nil, err
	}
	detKey := GetFileDeterministicKey(bucketKey[:32], indexBytes)
	key = detKey[:32]

	iv = indexBytes[0:16]

	return key, iv, nil
}

// Calculates the hash of a file
func CalculateFileHash(reader io.Reader) (string, error) {
	sha256Hasher := sha256.New()

	buf := make([]byte, 4096) // 4KB buffer size
	_, err := io.CopyBuffer(sha256Hasher, reader, buf)
	if err != nil {
		return "", fmt.Errorf("error reading data: %v", err)
	}

	sha256Result := sha256Hasher.Sum(nil)

	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Result)
	ripemd160Result := ripemd160Hasher.Sum(nil)

	return hex.EncodeToString(ripemd160Result), nil
}

func (b *BucketsService) CreateMetaFile(name, fileID, encryptVersion, folderUuid, plainName, fileType string, size int64, modTime time.Time) (*CreateMetaResponse, error) {
	endpoint := path.Join("files")
	reqBody := CreateMetaRequest{
		Name:             name,
		Bucket:           b.client.UserData.AccessData.User.Bucket,
		FileID:           fileID,
		EncryptVersion:   encryptVersion,
		FolderUuid:       folderUuid,
		Size:             size,
		PlainName:        plainName,
		Type:             fileType,
		ModificationTime: modTime,
	}

	var result CreateMetaResponse

	if resp, err := b.client.Post(APITypeDrive, endpoint, &reqBody, &result, nil); err != nil {
		return nil, b.client.GetError(endpoint, resp, err)
	}

	return &result, nil
}

func (b *BucketsService) FinishUpload(index string, shards []Shard) (*FinishUploadResp, error) {
	endpoint := path.Join("v2", "buckets", b.client.UserData.AccessData.User.Bucket, "files", "finish")
	payload := map[string]interface{}{
		"index":  index,
		"shards": shards,
	}

	headers := http.Header{}

	headers.Set("Authorization", b.client.UserData.BasicAuthHeader)
	var result FinishUploadResp

	if resp, err := b.client.Post(APITypeBase, endpoint, &payload, &result, &headers); err != nil {
		return nil, b.client.GetError(endpoint, resp, err)
	}

	return &result, nil
}
