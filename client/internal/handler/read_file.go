package handler

import (
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (env ClientEnv) HandleReadFile(metadata gophmodel.Metadata) (int, string, error) {
	dataInfo := gophmodel.DataToRead{
		StaticID: metadata.StaticID,
		UserID:   metadata.UserID,
		DataType: metadata.DataType,
	}

	body, err := json.Marshal(dataInfo)
	if err != nil {
		return 0, "", err
	}

	response, err := env.makeRequest(http.MethodGet, readPath, body, true)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return response.StatusCode, "", nil
	}

	filePath, err := makeFile(response.Body)
	if err != nil {
		return 0, "", err
	}
	return response.StatusCode, filePath, err
}

func makeFile(respBody io.ReadCloser) (string, error) {
	var fileData gophmodel.FileData

	bytes, err := io.ReadAll(respBody)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(bytes, &fileData)
	if err != nil {
		return "", err
	}

	filePath := "/tmp/" + fileData.Name

	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}

	_, err = file.Write([]byte(fileData.Data))
	if err != nil {
		return "", err
	}

	err = file.Close()
	if err != nil {
		return "", err
	}

	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	filePath = filepath.Join(path, filePath)
	return filePath, nil
}
