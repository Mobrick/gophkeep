package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TokenExp = time.Hour * 3
const SecretKey = "tratatata"
type userIDKey int

const (
    KeyUserID userIDKey = iota
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func CookieIsValid(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("auth_token")
	// проверяем есть ли кука
	if err != nil {
		return "", false
	}

	// в случае если кука есть проверяем что она проходит проверку подлинности
	token := cookie.Value
	id, ok := GetUserID(token)
	return id, ok
}

func GetUserID(tokenString string) (string, bool) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return "", false
	}

	if !token.Valid {
		log.Printf("Token is not valid")
		return "", false
	}

	log.Printf("Token is valid")
	return claims.UserID, true
}

func CreateNewCookie(id string) (http.Cookie, error) {
	tokenString, err := buildJWTString(id)
	if err != nil {
		return http.Cookie{}, err
	}
	// создание новой куки для юзера если такой куки не существует или она не проходит проверку подлинности
	cookie := http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   3600 * 3,
		HttpOnly: true,
		Secure:   false,
	}

	return cookie, nil
}

func buildJWTString(newID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// собственное утверждение

		UserID: newID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// CookieMiddleware создает куки если её не было, и добавляет к запросу и к ответу.
func CookieMiddleware(h http.Handler) http.Handler {
	cookieFn := func(w http.ResponseWriter, r *http.Request) {
		skipPaths := []string{"/ping", "/api/user/login", "/api/user/register"}

		if !slices.Contains(skipPaths, r.URL.Path) {
			userID, ok := CookieIsValid(r)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), KeyUserID, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		} else {
			h.ServeHTTP(w, r)
		}
	}
	return http.HandlerFunc(cookieFn)
}
