package handler

import (
	"encoding/json"
	"gophkeep/internal/logger"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
)

func (env *ClientEnv) HandleWrite(metadata gophmodel.SimpleMetadata, data []byte) (int, gophmodel.Metadata, error) {
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

	response, err := env.makeRequest(http.MethodPost, writePath, body, true)
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
	}
	return response.StatusCode, fullMetadata, nil
}
