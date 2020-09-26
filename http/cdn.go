package http

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Cdn struct {
	baseUri   string
	apiSecret string
	apiKey    string
}

func (cdn Cdn) uploadImage(img []byte, id string) error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	form := url.Values{
		"timestamp": {ts},
		"public_id": {id},
		"api_key":   {cdn.apiKey},
		"signature": {cdn.sign(ts, id)},
	}

	_, err := http.PostForm(cdn.baseUri, form)

	return err
}

func (cdn Cdn) sign(timestamp string, publicID string) string {
	paramStr := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, cdn.apiSecret)
	hasher := sha1.New()
	return string(hasher.Sum([]byte(paramStr)))
}
