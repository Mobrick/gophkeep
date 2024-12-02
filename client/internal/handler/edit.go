package handler

import (
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
)

func (env *ClientEnv) HandleEdit(metadata gophmodel.Metadata, newMetadata gophmodel.SimpleMetadata, data []byte) (int, gophmodel.Metadata, error) {
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
	response, err := env.makeRequest(http.MethodPost, editPath, body, true)
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
	}
	return response.StatusCode, fullMetadata, nil
}
