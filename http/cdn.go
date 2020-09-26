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
	"strconv"
	"time"
)

type Cdn struct {
	url       string
	apiSecret string
	apiKey    string
}

type uploadResponse struct {
	PublicId     string `json:"public_id"`
	Version      uint   `json:"version"`
	ResourceType string `json:"resource_type"`
	Format       string `json:"format"`
	Size         int    `json:"bytes"`
}

func (cdn Cdn) uploadImage(img io.Reader, filename string) (string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	publicID := "cat-" + time.Now().Format("20060102150405")

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

	err = form.WriteField("signature", sign(ts, publicID, cdn.apiSecret))

	if err != nil {
		return filename, err
	}

	formImage, err := form.CreateFormFile("file", filename)

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

	cdnResponse, err := http.Post(cdn.url+"/image/upload", form.FormDataContentType(), body)

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

		return uploadResp.PublicId, nil
	} else {
		return filename, errors.New("Request error:" + cdnResponse.Status)
	}
}

func sign(timestamp, publicID, secret string) string {
	paramStr := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, secret)
	hasher := sha1.New()
	return string(hasher.Sum([]byte(paramStr)))
}
