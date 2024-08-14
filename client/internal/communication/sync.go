package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"gophkeep/internal/logger"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"time"
)

func (env ClientEnv) SyncHandle() (int, []gophmodel.Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestURL := "http://localhost:8080"
	requestPath := "/api/user/sync"

	req, err := http.NewRequest("GET", requestURL+requestPath, nil)
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
		var buf bytes.Buffer
		var metadata []gophmodel.Metadata

		_, err = buf.ReadFrom(response.Body)
		if err != nil {
			return 0, nil, err
		}

		if err = json.Unmarshal(buf.Bytes(), &metadata); err != nil {
			logger.Log.Info("could not unmarshal metadata")
			return 0, nil, err
		}

		return response.StatusCode, metadata, nil
	}

	return response.StatusCode, nil, nil
}
