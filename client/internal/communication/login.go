package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"time"
)

func (env *ClientEnv) LoginHandle(loginData gophmodel.SimpleAccountData) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	requestURL := "http://localhost:8080"
	requestPath := "/api/user/login"

	body, err := json.Marshal(loginData)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", requestURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if len(response.Cookies()) != 0 {
		env.authCookie = response.Cookies()[0]
	}

	defer response.Body.Close()
	return response.StatusCode, nil
}
