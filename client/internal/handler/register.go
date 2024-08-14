package handler

import (
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
)

func (env *ClientEnv) HandleRegister(registerData gophmodel.SimpleAccountData) (int, error) {
	body, err := json.Marshal(registerData)
	if err != nil {
		return 0, err
	}

	response, err := env.makeRequest(http.MethodPost, registerPath, body, false)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if len(response.Cookies()) != 0 {
		env.authCookie = response.Cookies()[0]
	}
	return response.StatusCode, nil
}
