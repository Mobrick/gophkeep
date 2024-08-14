package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (env ClientEnv) ReadFileHandle(metadata gophmodel.Metadata) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestURL := "http://localhost:8080"
	requestPath := "/api/readfile"

	dataInfo := gophmodel.DataToRead{
		StaticID: metadata.StaticID,
		UserID:   metadata.UserID,
		DataType: metadata.DataType,
	}

	body, err := json.Marshal(dataInfo)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("GET", requestURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, "", err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)

	response, err := env.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return response.StatusCode, "", nil
	}

	var buf bytes.Buffer
	var fileData gophmodel.FileData

	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return 0, "", err
	}

	err = json.Unmarshal(buf.Bytes(), &fileData)
	if err != nil {
		return 0, "", err
	}

	filePath := "/tmp/" + fileData.Name

	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return 0, "", err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return 0, "", err
	}

	_, err = file.Write([]byte(fileData.Data))
	if err != nil {
		return 0, "", err
	}

	err = file.Close()
	if err != nil {
		return 0, "", err
	}

	path, err := os.Getwd()
	if err != nil {
		return 0, "", err
	}
	filePath = filepath.Join(path, filePath)
	return response.StatusCode, filePath, err
}
