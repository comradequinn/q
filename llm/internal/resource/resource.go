package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type (
	UploadRequest struct {
		Key  string
		URL  string
		File string
	}
	Reference struct {
		URI      string
		MIMEType string
		Label    string
	}
)

var mimeTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".tif":  "image/tiff",
	".tiff": "image/tiff",
	".ico":  "image/x-icon",
	".pdf":  "application/pdf",
}

func Upload(uploadRequest UploadRequest, debugPrintf func(msg string, args ...any)) (Reference, error) {

	contentType := mimeTypes[filepath.Ext(uploadRequest.File)]
	if contentType == "" {
		contentType = "text/plain"
	}

	fileInfo, err := os.Stat(uploadRequest.File)
	if err != nil {
		return Reference{}, fmt.Errorf("invalid filepath. '%v' file does not exist. %w", uploadRequest.File, err)
	}

	url := fmt.Sprintf(uploadRequest.URL, uploadRequest.Key)

	rq, err := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(`{"file":{"display_name":"%v"}}`, fileInfo.Name())))
	if err != nil {
		return Reference{}, fmt.Errorf("unable to create start-upload request. %w", err)
	}

	rq.Header.Set("X-Goog-Upload-Protocol", "resumable")
	rq.Header.Set("X-Goog-Upload-Command", "start")
	rq.Header.Set("X-Goog-Upload-Header-Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	rq.Header.Set("X-Goog-Upload-Header-Content-Type", contentType)
	rq.Header.Set("Content-Type", "application/json")

	debugPrintf("sending start upload request", "type", "start_upload_request", "url", url, "headers", rq.Header)

	rs, err := http.DefaultClient.Do(rq)
	if err != nil {
		return Reference{}, fmt.Errorf("error starting file upload. %w", err)
	}
	defer rs.Body.Close()

	body, _ := io.ReadAll(rs.Body)

	debugPrintf("received start upload response", "type", "start_upload_response", "status", rs.Status, "response", string(body))

	if rs.StatusCode != http.StatusOK {
		return Reference{}, fmt.Errorf("start-upload request failed with status code %v. %v", rs.StatusCode, string(body))
	}

	uploadURL := rs.Header.Get("X-Goog-Upload-Url")
	if uploadURL == "" {
		return Reference{}, fmt.Errorf("upload url not found in start-upload response header of 'x-goog-upload-url'")
	}

	file, err := os.Open(uploadRequest.File)
	if err != nil {
		return Reference{}, fmt.Errorf("unable to open file '%v' for upload. %w", uploadRequest.File, err)
	}
	defer file.Close()

	rq, err = http.NewRequest("POST", uploadURL, file) // Use the file as the request body
	if err != nil {
		return Reference{}, fmt.Errorf("unable to create upload-request. %w", err)
	}

	rq.Header.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	rq.Header.Set("X-Goog-Upload-Offset", "0")
	rq.Header.Set("X-Goog-Upload-Command", "upload, finalize")

	debugPrintf("sending upload request", "type", "upload_request", "url", url, "headers", rq.Header, "bytes", strconv.FormatInt(fileInfo.Size(), 10))

	rs, err = http.DefaultClient.Do(rq)
	if err != nil {
		return Reference{}, fmt.Errorf("error during upload-request. %w", err)
	}
	defer rs.Body.Close()

	body, err = io.ReadAll(rs.Body)

	debugPrintf("received upload response", "type", "upload_response", "status", rs.Status, "response", string(body))

	if rs.StatusCode != http.StatusOK || err != nil {
		return Reference{}, fmt.Errorf("upload-request failed with status code %v. error: %w. body: %v", rs.StatusCode, err, string(body))
	}

	uploadResponse := struct {
		File struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			MimeType    string `json:"mimeType"`
			SizeBytes   string `json:"sizeBytes"`
			CreateTime  string `json:"createTime"`
			UpdateTime  string `json:"updateTime"`
			URI         string `json:"uri"`
		} `json:"file"`
	}{}

	if err = json.Unmarshal(body, &uploadResponse); err != nil {
		return Reference{}, fmt.Errorf("unable to marshal upload-request response. %w", err)
	}

	return Reference{
		URI:      uploadResponse.File.URI,
		MIMEType: uploadResponse.File.MimeType,
		Label:    uploadResponse.File.DisplayName,
	}, nil
}
