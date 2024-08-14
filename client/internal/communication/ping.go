package communication

import (
	"context"
	"net/http"
	"time"
)

func (env ClientEnv) PingServerHandle() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestURL := "http://localhost:8080"
	requestPath := "/ping"

	req, err := http.NewRequest("GET", requestURL+requestPath, nil)
	if err != nil {
		return 0, err
	}

	req = req.WithContext(ctx)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	return response.StatusCode, nil
}