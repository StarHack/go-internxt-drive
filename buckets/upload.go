package buckets

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/internxt/rclone/config"
)

func UploadFile(cfg *config.Config, filePath, targetFolderUUID string, modTime time.Time) (*CreateMetaResponse, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	plainSize := int64(len(raw))
	var ph [32]byte
	if _, err := rand.Read(ph[:]); err != nil {
		return nil, fmt.Errorf("cannot generate random index: %w", err)
	}

	plainIndex := hex.EncodeToString(ph[:])
	fileKey, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, plainIndex)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	encReader, err := EncryptReader(f, fileKey, iv)
	if err != nil {
		return nil, err
	}
	sha256Hasher := sha256.New()
	sha1Hasher := sha1.New()
	r := io.TeeReader(encReader, sha256Hasher)
	r = io.TeeReader(r, sha1Hasher)
	specs := []UploadPartSpec{{Index: 0, Size: plainSize}}
	startResp, err := StartUpload(cfg, cfg.Bucket, specs)
	if err != nil {
		return nil, err
	}
	part := startResp.Uploads[0]
	if err := Transfer(cfg, part, r, plainSize); err != nil {
		return nil, err
	}
	encIndex := hex.EncodeToString(ph[:])
	partHash := hex.EncodeToString(sha1Hasher.Sum(nil))

	finishResp, err := FinishUpload(cfg, cfg.Bucket, encIndex, []Shard{{Hash: partHash, UUID: part.UUID}})
	if err != nil {
		return nil, err
	}
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	ext := strings.TrimPrefix(filepath.Ext(base), ".")
	meta, err := CreateMetaFile(cfg, name, cfg.Bucket, finishResp.ID, "03-aes", targetFolderUUID, name, ext, plainSize, modTime)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// UploadFileStream uploads data from the provided io.Reader into Internxt,
// encrypting it on the fly and creating the metadata file in the target folder.
// It returns the CreateMetaResponse of the created file entry.
func UploadFileStream(cfg *config.Config, targetFolderUUID, fileName string, in io.Reader, plainSize int64, modTime time.Time) (*CreateMetaResponse, error) {
	var ph [32]byte
	if _, err := rand.Read(ph[:]); err != nil {
		return nil, fmt.Errorf("cannot generate random index: %w", err)
	}
	plainIndex := hex.EncodeToString(ph[:])

	fileKey, iv, err := GenerateFileKey(cfg.Mnemonic, cfg.Bucket, plainIndex)
	if err != nil {
		return nil, err
	}

	encReader, err := EncryptReader(in, fileKey, iv)
	if err != nil {
		return nil, err
	}

	sha256Hasher := sha256.New()
	sha1Hasher := sha1.New()
	r := io.TeeReader(encReader, sha256Hasher)
	r = io.TeeReader(r, sha1Hasher)

	specs := []UploadPartSpec{{Index: 0, Size: plainSize}}
	startResp, err := StartUpload(cfg, cfg.Bucket, specs)
	if err != nil {
		return nil, err
	}

	if len(startResp.Uploads) == 0 {
		return nil, fmt.Errorf("startResp.Uploads is empty")
	}

	part := startResp.Uploads[0]

	if err := Transfer(cfg, part, r, plainSize); err != nil {
		return nil, err
	}

	encIndex := hex.EncodeToString(ph[:])
	partHash := hex.EncodeToString(sha1Hasher.Sum(nil))
	finishResp, err := FinishUpload(cfg, cfg.Bucket, encIndex, []Shard{{Hash: partHash, UUID: part.UUID}})
	if err != nil {
		return nil, err
	}

	base := filepath.Base(fileName)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	ext := strings.TrimPrefix(filepath.Ext(base), ".")
	meta, err := CreateMetaFile(cfg, name, cfg.Bucket, finishResp.ID, "03-aes", targetFolderUUID, name, ext, plainSize, modTime)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
