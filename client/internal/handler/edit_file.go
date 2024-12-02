package handler

import (
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
)

func (env *ClientEnv) HandleEditFile(metadata gophmodel.Metadata, newMetadata gophmodel.SimpleMetadata, filePath []byte) (int, gophmodel.Metadata, error) {
	editData := gophmodel.EditData{
		StaticID:    metadata.StaticID,
		UserID:      metadata.UserID,
		Name:        metadata.Name,
		Description: newMetadata.Description,
		DataType:    metadata.DataType,
	}

	var fullMetadata gophmodel.Metadata

	bodyInfo, err := json.Marshal(editData)
	if err != nil {
		return 0, fullMetadata, err
	}

	response, err := env.makeWriteFileRequest(editFilePath, string(filePath), bodyInfo)
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
