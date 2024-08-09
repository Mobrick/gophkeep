package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"time"
)

func (env *ClientEnv) EditHandle(metadata gophmodel.Metadata, newMetadata gophmodel.SimpleMetadata, data []byte) (int, gophmodel.Metadata, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()
	requestURL := "http://localhost:8080"
	requestPath := "/api/edit"

	editData := gophmodel.EditData{
		StaticID:    metadata.StaticID,
		UserID:      metadata.UserID,
		Name:        metadata.Name,
		Description: newMetadata.Description,
		DataType:    metadata.DataType,
		Data:        string(data),
	}

	var fullMetadata gophmodel.Metadata

	body, err := json.Marshal(editData)
	if err != nil {
		return 0, fullMetadata, err
	}

	req, err := http.NewRequest("POST", requestURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, fullMetadata, err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return 0, fullMetadata, err
	}
	if response.StatusCode == 200 {
		var buf bytes.Buffer

		_, err = buf.ReadFrom(response.Body)
		if err != nil {
			return 0, fullMetadata, err
		}

		if err = json.Unmarshal(buf.Bytes(), &fullMetadata); err != nil {
			return 0, fullMetadata, err
		}

		return response.StatusCode, fullMetadata, nil
	}
	defer response.Body.Close()
	return response.StatusCode, fullMetadata, nil
}
