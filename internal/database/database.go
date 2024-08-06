package database

import (
	"context"
	"database/sql"
	"gophkeep/internal/model"
	"log"
)

type Storage interface {
	PingDB() error
	AddNewAccount(context.Context, model.SimpleAccountData) (bool, string, error)
	CheckLogin(context.Context, model.SimpleAccountData) (string, error)
	AddLoginAndPasswordData(context.Context, model.Metadata, string, string) error
	AddCardData(context.Context, model.Metadata, string, string) error
	GetSimpleMetadataByUserID(context.Context, string) ([]model.SimpleMetadata, error)
	Delete(context.Context, model.DataToDelete) error
	Edit(context.Context, model.EditData) error
	Read(context.Context, model.DataToRead) (string, error)
	// TODO добавление произвольных данных
	Close()
}

func NewDB(connectionString string) Storage {
	dbData := PostgreDB{
		DatabaseConnection: NewDBConnection(connectionString),
	}

	return dbData
}

func NewDBConnection(connectionString string) *sql.DB {
	// Закрывается в основном потоке
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}
