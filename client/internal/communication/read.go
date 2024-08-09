package communication

import (
	"bytes"
	"context"
	"encoding/json"
	gophmodel "gophkeep/internal/model"
	"net/http"
	"time"
)

func (env ClientEnv) ReadHandle(metadata gophmodel.Metadata) (int, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	requestURL := "http://localhost:8080"
	requestPath := "/api/read"

	dataInfo := gophmodel.DataToRead{
		StaticID: metadata.StaticID,
		UserID:   metadata.UserID,
		DataType: metadata.DataType,
	}

	body, err := json.Marshal(dataInfo)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest("GET", requestURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, nil, err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	if response.StatusCode != 200 {
		return response.StatusCode, nil, nil
	}

	var buf bytes.Buffer

	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return 0, nil, err
	}

	data := buf.Bytes()

	defer response.Body.Close()

	return response.StatusCode, data, nil
}
