package communication

import (
	"context"
	"net/http"
	"time"
)

func (env ClientEnv) HandlePingServer() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestPath := "/ping"

	req, err := http.NewRequest("GET", baseURL+requestPath, nil)
	if err != nil {
		return 0, err
	}

	req = req.WithContext(ctx)

	env.httpClient = &http.Client{}
	response, err := env.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	return response.StatusCode, nil
}
