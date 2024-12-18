package database

import (
	"context"
	"database/sql"
	"gophkeep/internal/model"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage interface {
	PingDB() error
	AddNewAccount(context.Context, model.SimpleAccountData) (bool, string, error)
	CheckLogin(context.Context, model.SimpleAccountData) (string, error)
	AddData(context.Context, model.Metadata, string, string, string) error
	GetMetadataByUserID(context.Context, string) ([]model.Metadata, error)
	Delete(context.Context, model.DataToDelete) error
	Edit(context.Context, model.EditData, string, string) error
	Read(context.Context, model.DataToRead) (string, error)
	// TODO добавление произвольных данных
	Close()
}

func NewDB(ctx context.Context, connectionString string) Storage {
	dbData := PostgreDB{
		DatabaseConnection: NewDBConnection(ctx, connectionString),
	}

	err := dbData.CreateAccountsTable(ctx)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = dbData.CreateInfoTable(ctx)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = dbData.CreateCardTable(ctx)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	err = dbData.CreateLoginAndPasswordTable(ctx)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = dbData.CreateFileTable(ctx)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return dbData
}

func NewDBConnection(ctx context.Context, connectionString string) *sql.DB {
	// Закрывается в основном потоке
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}
