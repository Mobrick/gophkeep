package handler

import (
	"bytes"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"net/http"
)

func (env Env) EditHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID, ok := auth.CookieIsValid(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	var editData model.EditData
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &editData); err != nil {
		logger.Log.Info("could not unmarshal initial data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	
	if (editData.UserID != userID) {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = env.Storage.Edit(ctx, editData)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}
