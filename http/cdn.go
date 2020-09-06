package http

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Cdn implements external storage provider
type Cdn struct {
	Url string
	ApiKey string
	CloudName string
}


// UploadImage performs an upload to storage provider
func (cdn *Cdn) UploadImage(fileData []byte) error {
	timestamp := time.Now().Unix()
	formBody := url.Values{}
	formBody.Add("timestamp", strconv.FormatInt(timestamp, 10))
	formBody.Add("api_key", cdn.ApiKey)
	formBody.Add("file", fileData)

	hash := sha1.New()
	hash.Write([]byte("timestamp=" + formBody.Get("timestamp") + cdn.ApiKey))
	formBody.Add("signature", string(hash.Sum(nil)))
	_, err := http.PostForm(fmt.Sprintf("%s/%s/%s/upload", cdn.Url, cdn.CloudName, "image"), formBody)



	if err != nil {
		return  err
	}

	return nil
}

func NewCdn(url string, apiKey string, cloudName string) *Cdn {
	return &Cdn{
		Url:       url,
		ApiKey:    apiKey,
		CloudName: cloudName,
	}
}
