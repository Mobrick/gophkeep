package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func (env *ClientEnv) HandleEditFile(metadata gophmodel.Metadata, newMetadata gophmodel.SimpleMetadata, filePath []byte) (int, gophmodel.Metadata, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestPath := "/api/editfile"

	editData := gophmodel.EditData{
		StaticID:    metadata.StaticID,
		UserID:      metadata.UserID,
		Name:        metadata.Name,
		Description: newMetadata.Description,
		DataType:    metadata.DataType,
	}
	s := string(filePath)

	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}

	var fullMetadata gophmodel.Metadata

	bodyInfo, err := json.Marshal(editData)
	if err != nil {
		return 0, fullMetadata, err
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	metaPart, err := writer.CreateFormField("metadata")
	if err != nil {
		return 0, fullMetadata, err
	}

	metaPart.Write(bodyInfo)

	part, err := writer.CreateFormFile("file", s)
	if err != nil {
		return 0, fullMetadata, err
	}

	file, err := os.Open(s)
	if err != nil {
		return 0, fullMetadata, err
	}

	defer file.Close()

	_, err = io.Copy(part, file)
	if err != nil {
		return 0, fullMetadata, err
	}

	err = writer.Close()
	if err != nil {
		return 0, fullMetadata, err
	}

	req, err := http.NewRequest("POST", baseURL+requestPath, buf)
	if err != nil {
		return 0, fullMetadata, err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := env.httpClient.Do(req)
	if err != nil {
		return 0, fullMetadata, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return 0, fullMetadata, err
		}

		if err = json.Unmarshal(bytes, &fullMetadata); err != nil {
			return 0, fullMetadata, err
		}

		return response.StatusCode, fullMetadata, nil
	}
	return response.StatusCode, fullMetadata, nil
}
