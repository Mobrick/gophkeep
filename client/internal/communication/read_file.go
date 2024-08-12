package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"os"
	"time"
)

func (env ClientEnv) ReadFileHandle(metadata gophmodel.Metadata) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
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

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
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

	f, err := os.Create(filePath)
	if err != nil {
		return 0, "", err
	}

	_, err = f.Write([]byte(fileData.Data))
	if err != nil {
		return 0, "", err
	}
	return response.StatusCode, filePath, err
}
