package handler

import "net/http"

func (env Env) PingDBHandle(res http.ResponseWriter, _ *http.Request) {
	err := env.Storage.PingDB()
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
