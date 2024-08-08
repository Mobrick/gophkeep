package communication

import (
	"net/http"
)

type ClientEnv struct {
	authCookie *http.Cookie
}
