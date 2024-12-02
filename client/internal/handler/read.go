package handler

import (
	"encoding/json"
	"fmt"
	gophmodel "gophkeep/internal/model"
	"io"
	"net/http"
)

func (env ClientEnv) HandleRead(metadata gophmodel.Metadata) (int, []byte, error) {
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

	response, err := env.makeRequest(http.MethodGet, readPath, body, true)
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
