package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
	"time"
)

func (env ClientEnv) ReadHandle(metadata gophmodel.Metadata) (int, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSeconds)
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
		err = fmt.Errorf("error: %s with data: %s %s %s", err, dataInfo.StaticID, dataInfo.UserID, dataInfo.DataType)
		return 0, nil, err
	}

	req, err := http.NewRequest("GET", requestURL+requestPath, bytes.NewBuffer(body))
	if err != nil {
		return 0, nil, err
	}
	req = req.WithContext(ctx)
	req.AddCookie(env.authCookie)

	response, err := env.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return response.StatusCode, nil, nil
	}

	var readData gophmodel.ReadResponse

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}

	if err = json.Unmarshal(bytes, &readData); err != nil {
		return 0, nil, err
	}

	return response.StatusCode, []byte(readData.Data), nil
}
