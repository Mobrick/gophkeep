package handler

import (
	"context"
	"gophkeep/internal/config"
	"gophkeep/internal/database"
	"gophkeep/internal/encryption"
	"gophkeep/internal/model"
	"time"

	"github.com/google/uuid"
)

type Env struct {
	ConfigStruct *config.Config
	Storage      database.Storage
}

func StorageData(ctx context.Context, initialData model.InitialData, userID string, env Env, realSK string, encryptedSK string, data string) (model.Metadata, error) {
	metadata := model.Metadata{
		Name:        initialData.Name,
		Description: initialData.Description,
		DataType:    initialData.DataType,
		Created:     time.Now(),
		Changed:     time.Now(),
		StaticID:    uuid.New().String(),
		DynamicID:   uuid.New().String(),
		UserID:      userID,
	}

	encryptedData, err := encryption.EncryptSimpleData(realSK, data)
	if err != nil {
		return metadata, err
	}

	err = env.Storage.AddData(ctx, metadata, encryptedData, encryptedSK, initialData.DataType)
	if err != nil {
		return metadata, err
	}
	return metadata, err
}
