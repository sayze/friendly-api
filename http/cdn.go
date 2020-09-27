package http

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type Cdn struct {
	uploadUrl string
	imageUrl  string
	apiSecret string
	apiKey    string
}

type uploadResponse struct {
	PublicId     string `json:"public_id"`
	Version      uint   `json:"version"`
	SecureUrl    string `json:"secure_url"`
	ResourceType string `json:"resource_type"`
	Format       string `json:"format"`
	Size         int    `json:"bytes"`
}

func (cdn *Cdn) buildImageUrl(publicID string) string {
	return fmt.Sprintf("%s/%s", cdn.imageUrl, publicID)
}

func (cdn *Cdn) deleteImage(id string) error {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	err := form.WriteField("public_id", id)

	if err != nil {
		return err
	}

	signature, err := sign(strconv.FormatInt(time.Now().Unix(), 10), id, cdn.apiSecret)

	if err != nil {
		return err
	}

	err = form.WriteField("signature", signature)

	if err != nil {
		return err
	}

	_, err = http.Post(cdn.uploadUrl, form.FormDataContentType(), body)

	return err
}

func (cdn *Cdn) uploadImage(img io.Reader, filename, id string) (string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	publicID := id

	err := form.WriteField("timestamp", ts)

	if err != nil {
		return filename, err
	}

	err = form.WriteField("public_id", publicID)

	if err != nil {
		return filename, err
	}

	err = form.WriteField("api_key", cdn.apiKey)

	if err != nil {
		return filename, err
	}

	signature, err := sign(ts, publicID, cdn.apiSecret)

	if err != nil {
		return filename, err
	}

	err = form.WriteField("signature", signature)

	if err != nil {
		return filename, err
	}

	formImage, err := form.CreateFormFile("file", trimExt(filename))

	if err != nil {
		return filename, err
	}

	tmp, err := ioutil.ReadAll(img)

	if err != nil {
		return filename, err
	}

	_, err = formImage.Write(tmp)

	if err != nil {
		return filename, err
	}

	err = form.Close()

	if err != nil {
		return filename, err
	}

	cdnResponse, err := http.Post(cdn.uploadUrl, form.FormDataContentType(), body)

	if err != nil {
		return filename, err
	}

	if cdnResponse.StatusCode == http.StatusOK {
		respBody, err := ioutil.ReadAll(cdnResponse.Body)

		if err != nil {
			return filename, err
		}

		uploadResp := uploadResponse{}

		if err = json.Unmarshal(respBody, &uploadResp); err != nil {
			return filename, err
		}
		return uploadResp.SecureUrl, nil
	} else {
		return filename, errors.New("Request error:" + cdnResponse.Status)
	}
}

func sign(timestamp, publicID, secret string) (string, error) {
	paramStr := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, secret)
	hasher := sha1.New()
	_, err := hasher.Write([]byte(paramStr))

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func trimExt(filename string) string {
	fileExt := filepath.Ext(filename)
	return filename[0 : len(filename)-len(fileExt)]
}
