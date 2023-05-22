package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type B2AuthResp struct {
	AbsoluteMinimumPartSize int    `json:"absoluteMinimumPartSize"`
	AccountId               string `json:"accountId"`
	Allowed                 struct {
		BucketId     string   `json:"bucketId"`
		BucketName   string   `json:"bucketName"`
		Capabilities []string `json:"capabilities"`
		NamePrefix   string   `json:"namePrefix"`
	} `json:"allowed"`
	ApiUrl              string `json:"apiUrl"`
	AuthorizationToken  string `json:"authorizationToken"`
	DownloadUrl         string `json:"downloadUrl"`
	RecommendedPartSize int    `json:"recommendedPartSize"`
	S3ApiUrl            string `json:"s3ApiUrl"`
}

type B2UpUrlResp struct {
	BucketId           string `json:"bucketId"`
	UploadUrl          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

type B2UploadResp struct {
	FileId        string `json:"fileId"`
	FileName      string `json:"fileName"`
	AccountId     string `json:"accountId"`
	BucketId      string `json:"bucketId"`
	ContentLength int    `json:"contentLength"`
	ContentSha1   string `json:"contentSha1"`
	ContentType   string `json:"contentType"`
	FileInfo      struct {
		Author string `json:"author"`
	} `json:"fileInfo"`
	ServerSideEncryption struct {
		Algorithm string `json:"algorithm"`
		Mode      string `json:"mode"`
	} `json:"serverSideEncryption"`
}

type B2LargeFileStartResp struct {
	FileId   string `json:"fileId"`
	FileInfo struct {
		FileName    string `json:"fileName"`
		ContentType string `json:"contentType"`
	}
	FileName string `json:"fileName"`
}

type B2Error struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type B2AuthTokens struct {
	ApiUrl              string
	AuthorizationToken  string
	DownloadUrl         string
	RecommendedPartSize int
}

type FileInfo struct {
	FileName    string
	ContentType string
}

type B2LargeFileTokens struct {
	FileId   string
	FileInfo FileInfo
}

type B2UploadTokens struct {
	Endpoint    string
	AuthToken   string
	DownloadUrl string
}

// Get B2 keys from .env file.
func getKeys(key string) (string, error) {
	envPath := os.Getenv("PWD") + "/.env.benten"
	err := godotenv.Load(envPath)
	if err != nil {
		return "", fmt.Errorf("[GetKeys]: %w", err)
	}

	APP_KEY_ID := os.Getenv("B2_APP_KEY_ID")
	APP_KEY := os.Getenv("B2_APP_KEY")
	BUCKET_ID := os.Getenv("B2_BUCKET_ID")
	BUCKET_NAME := os.Getenv("B2_BUCKET_NAME")

	keys := map[string]string{
		"APP_KEY_ID":  APP_KEY_ID,
		"APP_KEY":     APP_KEY,
		"BUCKET_ID":   BUCKET_ID,
		"BUCKET_NAME": BUCKET_NAME,
	}

	return keys[key], nil
}

// Authorize B2 bucket for upload.
// DOCS: https://www.backblaze.com/b2/docs/b2_authorize_account.html
func authTokens() (B2AuthTokens, error) {
	keyID, err := getKeys("APP_KEY_ID")
	if err != nil {
		return B2AuthTokens{}, fmt.Errorf("[authTokens][GetKeys](APP_KEY_ID): %w", err)
	}

	key, err := getKeys("APP_KEY")
	if err != nil {
		return B2AuthTokens{}, fmt.Errorf("[authTokens][GetKeys](APP_KEY): %w", err)
	}

	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", keyID, key)))
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.backblazeb2.com/b2api/v2/b2_authorize_account", nil)
	if err != nil {
		return B2AuthTokens{}, fmt.Errorf("[authTokens][http.NewRequest]: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return B2AuthTokens{}, fmt.Errorf("[authTokens][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b2Error B2Error
		err := json.NewDecoder(resp.Body).Decode(&b2Error)
		if err != nil {
			return B2AuthTokens{}, fmt.Errorf("[authTokens][json.NewDecoder](b2Error): %w", err)
		}

		msg := b2Error.Message
		if msg == "" {
			msg = fmt.Sprintf("%d - %s", b2Error.Status, b2Error.Code)
		}
		return B2AuthTokens{}, fmt.Errorf("[authTokens][b2Error]: %s", msg)
	}

	var results B2AuthResp
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return B2AuthTokens{}, fmt.Errorf("[authTokens][json.NewDecoder](results): %w", err)
	}

	authTokens := B2AuthTokens{
		ApiUrl:              results.ApiUrl,
		AuthorizationToken:  results.AuthorizationToken,
		DownloadUrl:         results.DownloadUrl,
		RecommendedPartSize: results.RecommendedPartSize,
	}

	return authTokens, nil
}

// Start B2 large file upload.
// DOCS: https://www.backblaze.com/b2/docs/b2_start_large_file.html
func startLargeFile(params FileInfo) (B2LargeFileTokens, error) {
	authData, err := authTokens()
	if err != nil {
		return B2LargeFileTokens{}, fmt.Errorf("[startLargeFile][authTokens]: %w", err)
	}

	bucketID, err := getKeys("BUCKET_ID")
	if err != nil {
		return B2LargeFileTokens{}, fmt.Errorf("[startLargeFile][GetKeys](BUCKET_ID): %w", err)
	}

	payload := map[string]string{
		"bucketId":    bucketID,
		"fileName":    params.FileName,
		"contentType": params.ContentType,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/b2api/v2/b2_start_large_file", authData.ApiUrl), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return B2LargeFileTokens{}, fmt.Errorf("[startLargeFile][http.NewRequest]: %w", err)
	}

	req.Header.Set("Authorization", authData.AuthorizationToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return B2LargeFileTokens{}, fmt.Errorf("[startLargeFile][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b2Error B2Error
		err := json.NewDecoder(resp.Body).Decode(&b2Error)
		if err != nil {
			return B2LargeFileTokens{}, fmt.Errorf("[startLargeFile][json.NewDecoder](b2Error): %w", err)
		}

		msg := b2Error.Message
		if msg == "" {
			msg = fmt.Sprintf("%d - %s", b2Error.Status, b2Error.Code)
		}
		return B2LargeFileTokens{}, fmt.Errorf("[startLargeFile][b2Error]: %s", msg)
	}

	var results B2LargeFileStartResp
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return B2LargeFileTokens{}, fmt.Errorf("[startLargeFile][json.NewDecoder](results): %w", err)
	}

	largeFileTokens := B2LargeFileTokens{
		FileId:   results.FileId,
		FileInfo: FileInfo(results.FileInfo),
	}

	return largeFileTokens, nil
}

// Get B2 endpoint for upload.
// DOCS: https://www.backblaze.com/b2/docs/b2_get_upload_url.html
// DOCS: https://www.backblaze.com/b2/docs/b2_get_upload_part_url.html
func getUploadURL(large bool) (B2UploadTokens, error) {
	endpoint := "b2_get_upload_url"
	if large {
		endpoint = "b2_get_upload_part_url"
	}

	authData, err := authTokens()
	if err != nil {
		return B2UploadTokens{}, fmt.Errorf("[getUploadURL][authTokens]: %w", err)
	}

	bucketID, err := getKeys("BUCKET_ID")
	if err != nil {
		return B2UploadTokens{}, fmt.Errorf("[getUploadURL][GetKeys](BUCKET_ID): %w", err)
	}

	payload := map[string]string{"bucketId": bucketID}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/b2api/v1/%s", authData.ApiUrl, endpoint), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return B2UploadTokens{}, fmt.Errorf("[getUploadURL][http.NewRequest]: %w", err)
	}

	req.Header.Set("Authorization", authData.AuthorizationToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return B2UploadTokens{}, fmt.Errorf("[getUploadURL][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b2Error B2Error
		err := json.NewDecoder(resp.Body).Decode(&b2Error)
		if err != nil {
			return B2UploadTokens{}, fmt.Errorf("[getUploadURL][json.NewDecoder](b2Error): %w", err)
		}

		msg := b2Error.Message
		if msg == "" {
			msg = fmt.Sprintf("%d - %s", b2Error.Status, b2Error.Code)
		}
		return B2UploadTokens{}, fmt.Errorf("[getUploadURL][b2Error]: %s", msg)
	}

	var results B2UpUrlResp
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return B2UploadTokens{}, fmt.Errorf("[getUploadURL][json.NewDecoder](results): %w", err)
	}

	uploadTokens := B2UploadTokens{
		Endpoint:    results.UploadUrl,
		AuthToken:   results.AuthorizationToken,
		DownloadUrl: authData.DownloadUrl,
	}

	return uploadTokens, nil
}

// Upload file to B2 bucket.
// DOCS: https://www.backblaze.com/b2/docs/b2_upload_file.html
func UploadToB2(data []byte, name string, fileType string, large bool) (string, error) {
	authData, err := getUploadURL(large)
	if err != nil {
		return "", fmt.Errorf("[UploadToB2][getUploadURL]: %w", err)
	}

	hasher := sha1.New()
	hasher.Write(data)
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	if fileType == "" {
		fileType = "b2/x-auto"
	}

	req, err := http.NewRequest("POST", authData.Endpoint, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("[UploadToB2][http.NewRequest]: %w", err)
	}

	req.Header.Set("Authorization", authData.AuthToken)
	req.Header.Set("X-Bz-File-Name", name)
	req.Header.Set("Content-Type", fileType)
	req.Header.Set("Content-Length", strconv.Itoa(len(data)))
	req.Header.Set("X-Bz-Content-Sha1", hash)
	req.Header.Set("X-Bz-Info-Author", "rivendell")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("[UploadToB2][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b2Error B2Error
		err := json.NewDecoder(resp.Body).Decode(&b2Error)
		if err != nil {
			return "", fmt.Errorf("[UploadToB2][json.NewDecoder](b2Error): %w", err)
		}

		msg := b2Error.Message
		if msg == "" {
			msg = fmt.Sprintf("%d - %s", b2Error.Status, b2Error.Code)
		}
		return "", fmt.Errorf("[UploadToB2][b2Error]: %s", msg)
	}

	var results B2UploadResp
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return "", fmt.Errorf("[UploadToB2][json.NewDecoder](results): %w", err)
	}

	bucketName, err := getKeys("BUCKET_NAME")
	if err != nil {
		return "", fmt.Errorf("[UploadToB2][GetKeys](BUCKET_NAME): %w", err)
	}

	log.Printf("[UploadToB2]: Uploaded '%s'.\n", results.FileName)

	publicURL := fmt.Sprintf("%s/file/%s/%s", authData.DownloadUrl, bucketName, results.FileName)

	return publicURL, nil
}
