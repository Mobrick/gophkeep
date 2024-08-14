package communication

import (
	"context"
	"encoding/json"
	"gophkeep/internal/logger"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
	"time"
)

func (env ClientEnv) HandleSync() (int, []gophmodel.Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestPath := "/api/user/sync"

	req, err := http.NewRequest("GET", baseURL+requestPath, nil)
	if err != nil {
		return 0, nil, err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)

	response, err := env.httpClient.Do(req)
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
