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
)
