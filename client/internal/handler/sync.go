package handler

import (
	"encoding/json"
	"gophkeep/internal/logger"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
)

func (env ClientEnv) HandleSync() (int, []gophmodel.Metadata, error) {
	response, err := env.makeRequest(http.MethodGet, syncPath, nil, true)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNoContent {
		return response.StatusCode, nil, nil
	}

	if response.StatusCode == http.StatusOK {
		var metadata []gophmodel.Metadata
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return 0, nil, err
		}

		if err = json.Unmarshal(bytes, &metadata); err != nil {
			logger.Log.Info("could not unmarshal metadata")
			return 0, nil, err
		}

		return response.StatusCode, metadata, nil
	}

	return response.StatusCode, nil, nil
}
