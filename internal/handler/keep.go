package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/encryption"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (env Env) KeepHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID, ok := auth.CookieIsValid(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	var initialData model.InitialData
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &initialData); err != nil {
		logger.Log.Info("could not unmarshal initial data")
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

	var metadata model.Metadata

	switch initialData.DataType {
	case "passwords":
		metadata, err = loginAndPasswordKeep(ctx, initialData, userID, env, realSK, encryptedSK)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case "cards":
		metadata, err = cardKeep(ctx, initialData, userID, env, realSK, encryptedSK)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case "binaries":
	case "texts":
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

func loginAndPasswordKeep(ctx context.Context, initialData model.InitialData, userID string, env Env, realSK string, encryptedSK string) (model.Metadata, error) {
	metadata := model.Metadata{
		Name:        initialData.Name,
		Description: initialData.Description,
		DataType:    initialData.DataType,
		Created:     time.Now(),
		Changed:     time.Now(),
		StaticID:    uuid.New().String(),
		DynamicID:   uuid.New().String(),
		UserID:      userID,
	}

	encryptedData, err := encryption.EncryptSimpleData(realSK, initialData.Data)
	if err != nil {
		return metadata, err
	}

	err = env.Storage.AddLoginAndPasswordData(ctx, metadata, encryptedData, encryptedSK)
	if err != nil {
		return metadata, err
	}
	return metadata, err
}

func cardKeep(ctx context.Context, initialData model.InitialData, userID string, env Env, realSK string, encryptedSK string) (model.Metadata, error) {
	metadata := model.Metadata{
		Name:        initialData.Name,
		Description: initialData.Description,
		DataType:    initialData.DataType,
		Created:     time.Now(),
		Changed:     time.Now(),
		StaticID:    uuid.New().String(),
		DynamicID:   uuid.New().String(),
		UserID:      userID,
	}

	encryptedData, err := encryption.EncryptSimpleData(realSK, initialData.Data)
	if err != nil {
		return metadata, err
	}

	err = env.Storage.AddCardData(ctx, metadata, encryptedData, encryptedSK)
	if err != nil {
		return metadata, err
	}
	return metadata, err
}
