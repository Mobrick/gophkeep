package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"gophkeep/internal/logger"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
	"time"
)

func (env *ClientEnv) WriteHandle(metadata gophmodel.SimpleMetadata, data []byte) (int, gophmodel.Metadata, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestURL := "http://localhost:8080"
	requestPath := "/api/keep"

	initialData := gophmodel.InitialData{
		Name:        metadata.Name,
		Description: metadata.Description,
		DataType:    metadata.DataType,
		Data:        string(data),
	}

	var fullMetadata gophmodel.Metadata

	body, err := json.Marshal(initialData)
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
			logger.Log.Info("could not unmarshal metadata")
			return 0, fullMetadata, err
		}

		return response.StatusCode, fullMetadata, nil
	}
	return response.StatusCode, fullMetadata, nil
}
