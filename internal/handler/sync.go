package handler

import (
	"encoding/json"
	"gophkeep/internal/auth"
	"gophkeep/internal/logger"
	"net/http"
)

func (env Env) SyncHandle(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID := ctx.Value(auth.KeyUserID).(string)

	metadata, err := env.Storage.GetMetadataByUserID(ctx, userID)
	if err != nil {
		logger.Log.Debug("could not get urls by user id")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(metadata) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
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
