package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"gophkeep/client/internal/gophmodel"
	"net/http"
	"time"
)

func PingServerHandle() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
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

func LoginHandle(loginData gophmodel.SimpleAccountData) (int, error) {
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
	defer response.Body.Close()
	return response.StatusCode, nil
}
