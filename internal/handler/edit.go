package handler

import (
	"bytes"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/encryption"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"net/http"
	"time"

	"github.com/google/uuid"
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

	// создаем и шифруем этот ключ ключом шифрования
	encryptedSK, realSK, err := encryption.GenerateSK(uuid.NewString())
	if err != nil {
		logger.Log.Info("could not create key")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &editData); err != nil {
		logger.Log.Info("could not unmarshal initial data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	encryptedData, err := encryption.EncryptSimpleData(realSK, editData.Data)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if editData.UserID != userID {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = env.Storage.Edit(ctx, editData, encryptedData, encryptedSK)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	metadata := model.Metadata{
		StaticID:    editData.StaticID,
		UserID:      editData.UserID,
		Changed:     time.Now(),
		Name:        editData.Name,
		Description: editData.Description,
		DataType:    editData.DataType,
	}

	resp, err := json.Marshal(metadata)
	if err != nil {
		logger.Log.Debug("could not marshal response")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(resp))
}
