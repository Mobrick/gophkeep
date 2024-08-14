package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"time"
)

func (env *ClientEnv) HandleLogin(loginData gophmodel.SimpleAccountData) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	requestPath := "/api/user/login"

	body, err := json.Marshal(loginData)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", baseURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	response, err := env.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if len(response.Cookies()) != 0 {
		env.authCookie = response.Cookies()[0]
	}
	return response.StatusCode, nil
}
