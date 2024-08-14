package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"time"
)

func (env *ClientEnv) HandleDelete(metadata gophmodel.Metadata) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestPath := "/api/delete"

	deleteData := gophmodel.DataToDelete{
		StaticID: metadata.StaticID,
		UserID:   metadata.UserID,
		DataType: metadata.DataType,
	}

	body, err := json.Marshal(deleteData)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", baseURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)
	req.Header.Set("Content-Type", "application/json")

	response, err := env.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	return response.StatusCode, nil
}
