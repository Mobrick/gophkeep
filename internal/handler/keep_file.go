package handler

import (
	"bytes"
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/encryption"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func (env Env) KeepFileHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID, ok := auth.CookieIsValid(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	var initialData model.InitialData

	req.ParseMultipartForm(2097152)

	metadataJson := req.FormValue("metadata")
	file, header, err := req.FormFile("file")
	if err != nil {
		logger.Log.Info("could not take file")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, file); err != nil {
		logger.Log.Info("could not read file")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	fileData := model.FileData{
		Name: header.Filename,
		Size: header.Size,
		Data: buf.String(),
	}

	fileJSON, err := json.Marshal(fileData)
	if err != nil {
		logger.Log.Debug("could not marshal response")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// создаем и шифруем этот ключ ключом шифрования
	encryptedSK, realSK, err := encryption.GenerateSK(uuid.NewString())
	if err != nil {
		logger.Log.Info("could not create key")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal([]byte(metadataJson), &initialData); err != nil {
		logger.Log.Info("could not unmarshal initial data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	metadata, err := StorageData(ctx, initialData, userID, env, realSK, encryptedSK, string(fileJSON))
	if err != nil {
		logger.Log.Info("could not keep file data")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

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
