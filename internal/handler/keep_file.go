package handler

import (
	"encoding/json"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"net/http"
)

func (env Env) KeepFileHandle(res http.ResponseWriter, req *http.Request) {
	_ = req.Context()
	/* _, ok := auth.CookieIsValid(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	} */

	var initialData model.InitialData

	req.ParseMultipartForm(2097152)

	metadataJson := req.FormValue("metadata")
	file, header, err := req.FormFile("file")
	if err != nil {
		logger.Log.Info("could not take file")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	if err = json.Unmarshal([]byte(metadataJson), &initialData); err != nil {
		logger.Log.Info("could not unmarshal initial data")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	testData := model.TestFileData{
		Name:        header.Filename,
		Description: initialData.Description,
	}

	resp, err := json.Marshal(testData)
	if err != nil {
		logger.Log.Debug("could not marshal response")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(resp))
}
