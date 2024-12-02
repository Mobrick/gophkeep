package handler

import (
	"bytes"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"net/http"
)

func (env Env) AuthHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var loginData model.SimpleAccountData
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &loginData); err != nil {
		logger.Log.Info("could not unmarshal registration data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := env.Storage.CheckLogin(ctx, loginData)
	if err != nil {
		logger.Log.Info("could not check login data")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(id) == 0 {
		res.WriteHeader(http.StatusUnauthorized)
	}

	cookie, err := auth.CreateNewCookie(id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(res, &cookie)

	res.WriteHeader(http.StatusOK)
}
