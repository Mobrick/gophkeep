package handler

import (
	"bytes"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"log"
	"net/http"

	"go.uber.org/zap"
)

func (env Env) RegisterHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var registrationData model.SimpleAccountData
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &registrationData); err != nil {
		logger.Log.Info("could not unmarshal registration data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	storage := env.Storage
	loginAlreadyInUse, id, err := storage.AddNewAccount(ctx, registrationData)
	if err != nil {
		logger.Log.Info("could not complete user registration", zap.String("Attempted login", string(registrationData.Login)))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if loginAlreadyInUse {
		logger.Log.Info("login already in use", zap.String("Attempted login", string(registrationData.Login)))
		res.WriteHeader(http.StatusConflict)
		return
	}

	cookie, err := auth.CreateNewCookie(id)
	if err != nil {
		log.Printf("could not create cookie: " + err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(res, &cookie)

	res.WriteHeader(http.StatusOK)
}
