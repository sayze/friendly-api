// Package cloudinary implements domain.ImageStore against the Cloudinary
// upload API.
package cloudinary

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sayze/friendly-api/internal/friend/domain"
)

// Config holds the Cloudinary account settings needed to sign and send
// upload/delete requests.
type Config struct {
	UploadURL string
	APIKey    string
	APISecret string
}

type imageStore struct {
	cfg    Config
	client *http.Client
	now    func() time.Time
}

// NewImageStore builds a domain.ImageStore backed by Cloudinary.
func NewImageStore(cfg Config) domain.ImageStore {
	return &imageStore{cfg: cfg, client: http.DefaultClient, now: time.Now}
}

type uploadResponse struct {
	SecureURL string `json:"secure_url"`
}

func (s *imageStore) Upload(ctx context.Context, img io.Reader, filename, id string) (string, error) {
	ts := strconv.FormatInt(s.now().Unix(), 10)

	signature, err := sign(ts, id, s.cfg.APISecret)
	if err != nil {
		return "", err
	}

	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	fields := map[string]string{
		"timestamp": ts,
		"public_id": id,
		"api_key":   s.cfg.APIKey,
		"signature": signature,
	}
	for k, v := range fields {
		if err := form.WriteField(k, v); err != nil {
			return "", err
		}
	}

	fw, err := form.CreateFormFile("file", trimExt(filename))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(fw, img); err != nil {
		return "", err
	}
	if err := form.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.UploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudinary upload failed: %s: %s", resp.Status, respBody)
	}

	var uploaded uploadResponse
	if err := json.Unmarshal(respBody, &uploaded); err != nil {
		return "", err
	}

	return uploaded.SecureURL, nil
}

func (s *imageStore) Delete(ctx context.Context, id string) error {
	ts := strconv.FormatInt(s.now().Unix(), 10)

	signature, err := sign(ts, id, s.cfg.APISecret)
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	fields := map[string]string{
		"public_id": id,
		"signature": signature,
		"timestamp": ts,
		"api_key":   s.cfg.APIKey,
	}
	for k, v := range fields {
		if err := form.WriteField(k, v); err != nil {
			return err
		}
	}
	if err := form.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.UploadURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudinary delete failed: %s: %s", resp.Status, respBody)
	}

	return nil
}

// sign computes the Cloudinary request signature for a public_id/timestamp
// pair, per https://cloudinary.com/documentation/authentication_signatures.
func sign(timestamp, publicID, secret string) (string, error) {
	paramStr := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, secret)
	hasher := sha1.New()
	if _, err := hasher.Write([]byte(paramStr)); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// trimExt strips the file extension from filename, since Cloudinary derives
// the format from the uploaded content rather than the public_id.
func trimExt(filename string) string {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)]
}
