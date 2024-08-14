package communication

import (
	"net/http"
)

type ClientEnv struct {
	authCookie *http.Cookie
	httpClient *http.Client
}

const (
	TimeoutSeconds = 10
	baseURL        = "http://localhost:8080"
	loginPath      = "/api/user/login"
	registerPath   = "/api/user/register"
	deletePath     = "/api/delete"
	editfilePath   = "/api/editfile"
	editPath       = "/api/edit"
	pingPath       = "/ping"
	readFilePath   = "/api/readfile"
	readPath       = "/api/read"
	syncPath       = "/api/user/sync"
	writeFilePath  = "/api/keepfile"
	writePath      = "/api/keep"
)
