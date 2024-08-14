package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
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
	editFilePath   = "/api/editfile"
	editPath       = "/api/edit"
	pingPath       = "/ping"
	readFilePath   = "/api/readfile"
	readPath       = "/api/read"
	syncPath       = "/api/user/sync"
	writeFilePath  = "/api/keepfile"
	writePath      = "/api/keep"
)

func (env ClientEnv) makeRequest(httpMethod string, requestPath string, body []byte, addAuthCookie bool) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()
	req, err := http.NewRequest(httpMethod, baseURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if addAuthCookie {
		req.AddCookie(env.authCookie)
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := env.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return response, err
}

func (env ClientEnv) makeWriteFileRequest(requestPath string, filepath string, bodyInfo []byte) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
	defer cancel()

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	metaPart, err := writer.CreateFormField("metadata")
	if err != nil {
		return nil, err
	}

	metaPart.Write(bodyInfo)

	s := fixFilePath(filepath)

	part, err := writer.CreateFormFile("file", s)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(s)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+requestPath, buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)

	response, err := env.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return response, err
}

func fixFilePath(filepath string) string {
	s := filepath

	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}
