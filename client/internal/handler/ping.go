package handler

import (
	"net/http"
)

func (env *ClientEnv) HandlePingServer() (int, error) {
	env.httpClient = &http.Client{}
	response, err := env.makeRequest(http.MethodGet, pingPath, nil, false)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	return response.StatusCode, nil
}
