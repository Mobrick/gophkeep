package handler

import (
	"bytes"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"net/http"
)

func (env Env) ReadHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID := ctx.Value(auth.KeyUserID).(string)

	var readData model.DataToRead
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &readData); err != nil {
		logger.Log.Info("could not unmarshal initial data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if readData.UserID != userID {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := env.Storage.Read(ctx, readData)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	readResponse := model.ReadResponse{
		StaticID: readData.StaticID,
		Data:     data,
	}

	resp, err := json.Marshal(readResponse)
	if err != nil {
		logger.Log.Debug("could not marshal response")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(resp))
}
