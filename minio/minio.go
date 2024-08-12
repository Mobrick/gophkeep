package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client interface {
	InitMinio() error
	CreateOne(file helpers.FileDataType) (string, error)
	GetOne(objectID string) (string, error)
	DeleteOne(objectID string) error
}

type minioClient struct {
	mc *minio.Client // Клиент Minio
}

func NewMinioClient() Client {
	return &minioClient{} // Возвращает новый экземпляр minioClient с указанным именем бакета
}

func (m *minioClient) InitMinio() error {
	ctx := context.Background()

	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("1", "key", ""),
		Secure: false,
	}) 
	if err != nil {
		return err
	}

	m.mc = client

	exists, err := m.mc.BucketExists(ctx, "files")
	if err != nil {
		return err
	}
	if !exists {
		err := m.mc.MakeBucket(ctx, "files", minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
