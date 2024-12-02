package handler

import (
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
)

func (env *ClientEnv) HandleDelete(metadata gophmodel.Metadata) (int, error) {
	deleteData := gophmodel.DataToDelete{
		StaticID: metadata.StaticID,
		UserID:   metadata.UserID,
		DataType: metadata.DataType,
	}

	body, err := json.Marshal(deleteData)
	if err != nil {
		return 0, err
	}

	response, err := env.makeRequest(http.MethodPost, deletePath, body, true)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	return response.StatusCode, nil
}
