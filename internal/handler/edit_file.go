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
	"time"

	"github.com/google/uuid"
)

func (env Env) EditFileHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID, ok := auth.CookieIsValid(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	var editData model.EditData
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

	if err = json.Unmarshal([]byte(metadataJson), &editData); err != nil {
		logger.Log.Info("could not unmarshal initial data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	encryptedData, err := encryption.EncryptSimpleData(realSK, string(fileJSON))
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
