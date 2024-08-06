package handler

import (
	"bytes"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"net/http"
)

func (env Env) DeleteHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID, ok := auth.CookieIsValid(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	var deleteData model.DataToDelete
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &deleteData); err != nil {
		logger.Log.Info("could not unmarshal data to delete")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if (deleteData.UserID != userID) {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = env.Storage.Delete(ctx, deleteData)
	if err != nil {
		logger.Log.Debug("could not delete")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
