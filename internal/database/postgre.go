package database

import (
	"context"
	"database/sql"
	"errors"
	"gophkeep/internal/encryption"
	"gophkeep/internal/logger"
	"gophkeep/internal/model"
	"log"
	"time"

	accountsmigrations "gophkeep/internal/database/accounts_migrations"
	cardsmigrations "gophkeep/internal/database/cards_migrations"
	infosmigrations "gophkeep/internal/database/infos_migrations"
	passwordsmigrations "gophkeep/internal/database/passwords_migrations"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

const accountsTableName = "accounts"

type PostgreDB struct {
	DatabaseConnection *sql.DB
}

func (dbData PostgreDB) PingDB() error {
	err := dbData.DatabaseConnection.Ping()
	return err
}

func (dbData PostgreDB) Close() {
	dbData.DatabaseConnection.Close()
}

// Возвращает true если такой логин уже хранится в базе
func (dbData PostgreDB) AddNewAccount(ctx context.Context, accountData model.SimpleAccountData) (bool, string, error) {

	err := dbData.createAccountsTable(ctx)
	if err != nil {
		return false, "", err
	}

	id := uuid.New().String()

	insertStmt := "INSERT INTO " + accountsTableName + " (uuid, username, password)" +
		" VALUES ($1, $2, $3)"

	_, err = dbData.DatabaseConnection.ExecContext(ctx, insertStmt, id, accountData.Login, accountData.Password)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			log.Printf("login %s already in database", accountData.Login)
			return true, "", nil
		}
		log.Printf("Failed to insert a record: " + accountData.Login)
		return false, "", err
	}

	return false, id, nil
}

func (dbData PostgreDB) createAccountsTable(ctx context.Context) error {
	provider, err := goose.NewProvider(database.DialectPostgres, dbData.DatabaseConnection, accountsmigrations.EmbedAccounts)
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}

	for _, r := range results {
		log.Printf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration)
	}

	logger.Log.Debug("Created table with goose embed")
	return nil
}

func (dbData PostgreDB) CheckLogin(ctx context.Context, accountData model.SimpleAccountData) (string, error) {

	checkStmt := "SELECT uuid FROM " + accountsTableName + " WHERE username=$1 AND password=$2"

	var id string

	err := dbData.DatabaseConnection.QueryRowContext(ctx, checkStmt, accountData.Login, accountData.Password).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No such login password pair: " + accountData.Login)
			return "", nil

		}
		log.Printf("Error querying database: " + accountData.Login)
		return "", err
	}

	return id, nil
}

func (dbData PostgreDB) AddLoginAndPasswordData(ctx context.Context, metadata model.Metadata, data string, dataSK string) error {
	err := dbData.createInfoTable(ctx)
	if err != nil {
		return err
	}
	err = dbData.createLoginAndPasswordTable(ctx)
	if err != nil {
		return err
	}
	tx, err := dbData.DatabaseConnection.Begin()
	if err != nil {
		return err
	}

	insertStmt := "INSERT INTO infos (static_id, dynamic_id, name, description, type, account_uuid, created_at, changed_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	_, err = dbData.DatabaseConnection.ExecContext(ctx, insertStmt,
		metadata.StaticID, metadata.DynamicID, metadata.Name, metadata.Description, metadata.DataType, metadata.UserID, metadata.Created, metadata.Changed)

	if err != nil {
		return err
	}

	passwordInsertStmt := "INSERT INTO passwords (id, data, sk) VALUES ($1, $2, $3)"

	_, err = dbData.DatabaseConnection.ExecContext(ctx, passwordInsertStmt, metadata.StaticID, data, dataSK)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (dbData PostgreDB) AddCardData(ctx context.Context, metadata model.Metadata, data string, dataSK string) error {
	err := dbData.createInfoTable(ctx)
	if err != nil {
		return err
	}
	err = dbData.createCardTable(ctx)
	if err != nil {
		return err
	}

	tx, err := dbData.DatabaseConnection.Begin()
	if err != nil {
		return err
	}

	insertStmt := "INSERT INTO infos (static_id, dynamic_id, name, description, type, account_uuid, created_at, changed_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	_, err = dbData.DatabaseConnection.ExecContext(ctx, insertStmt,
		metadata.StaticID, metadata.DynamicID, metadata.Name, metadata.Description, metadata.DataType, metadata.UserID, metadata.Created, metadata.Changed)

	if err != nil {
		return err
	}

	cardInsertStmt := "INSERT INTO cards (id, data, sk) VALUES ($1, $2, $3)"

	_, err = dbData.DatabaseConnection.ExecContext(ctx, cardInsertStmt, metadata.StaticID, data, dataSK)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (dbData PostgreDB) createInfoTable(ctx context.Context) error {
	provider, err := goose.NewProvider(database.DialectPostgres, dbData.DatabaseConnection, infosmigrations.EmbedInfos)
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}

	for _, r := range results {
		log.Printf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration)
	}

	logger.Log.Debug("Created table with goose embed")
	return nil
}

func (dbData PostgreDB) createLoginAndPasswordTable(ctx context.Context) error {
	provider, err := goose.NewProvider(database.DialectPostgres, dbData.DatabaseConnection, passwordsmigrations.EmbedPasswords, goose.WithAllowOutofOrder(true))
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}

	for _, r := range results {
		log.Printf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration)
	}

	logger.Log.Debug("Created table with goose embed")
	return nil
}

func (dbData PostgreDB) createCardTable(ctx context.Context) error {
	provider, err := goose.NewProvider(database.DialectPostgres, dbData.DatabaseConnection, cardsmigrations.EmbedCards)
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}

	for _, r := range results {
		log.Printf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration)
	}

	logger.Log.Debug("Created table with goose embed")
	return nil
}

func (dbData PostgreDB) GetMetadataByUserID(ctx context.Context, userID string) ([]model.Metadata, error) {
	metadata := make([]model.Metadata, 0)
	exists, err := dbData.tableExists(ctx, "infos")
	if err != nil {
		return nil, err
	}
	if !exists {
		return metadata, nil
	}

	stmt := "SELECT static_id, dynamic_id, name, description, type, created_at, changed_at FROM infos WHERE account_uuid = $1"
	rows, err := dbData.DatabaseConnection.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var name, description, dataType, static_id, dynamic_id string
		var created_at, changed_at time.Time
		err := rows.Scan(&static_id, &dynamic_id, &name, &description, &dataType, &created_at, &changed_at)
		if err != nil {
			return nil, err
		}

		metadata = append(metadata, model.Metadata{
			StaticID:    static_id,
			DynamicID:   dynamic_id,
			Name:        name,
			Description: description,
			DataType:    dataType,
			UserID:      userID,
			Created:     created_at,
			Changed:     changed_at,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()
	return metadata, nil
}

func (dbData PostgreDB) tableExists(ctx context.Context, tableName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_name = $1
	);`

	err := dbData.DatabaseConnection.QueryRowContext(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (dbData PostgreDB) Delete(ctx context.Context, deleteData model.DataToDelete) error {
	decryptedData, err := dataAccess(ctx, dbData, deleteData.StaticID, deleteData.DataType)
	if err != nil {
		return err
	}

	if len(decryptedData) == 0 {
		return errors.New("data is not accessible")
	}

	deleteFromInfosStmt := "DELETE FROM infos WHERE static_id = $1 AND account_uuid = $2"
	_, err = dbData.DatabaseConnection.ExecContext(ctx, deleteFromInfosStmt, deleteData.StaticID, deleteData.UserID)

	if err != nil {
		return err
	}

	dataType := deleteData.DataType

	deleteFromDataStmt := "DELETE FROM " + dataType + " WHERE static_id = $1"
	_, err = dbData.DatabaseConnection.ExecContext(ctx, deleteFromDataStmt, deleteData.StaticID)
	if err != nil {
		return err
	}

	return nil
}

func (dbData PostgreDB) Read(ctx context.Context, readData model.DataToRead) (string, error) {
	decryptedData, err := dataAccess(ctx, dbData, readData.StaticID, readData.DataType)
	if err != nil {
		return "", err
	}

	if len(decryptedData) == 0 {
		return "", errors.New("data is not accessible")
	}

	return decryptedData, nil
}

func (dbData PostgreDB) Edit(ctx context.Context, editData model.EditData) error {
	decryptedData, err := dataAccess(ctx, dbData, editData.StaticID, editData.DataType)
	if err != nil {
		return err
	}

	if len(decryptedData) == 0 {
		return errors.New("data is not accessible")
	}

	stmt := "UPDATE infos SET dynamic_id = $1, description = $2, changed_at = $3 WHERE static_id = $4 AND account_uuid = $5"

	dynamicID := uuid.New().String()
	_, err = dbData.DatabaseConnection.ExecContext(ctx, stmt, dynamicID, editData.Description, time.Now(), editData.StaticID, editData.UserID)

	if err != nil {
		return err
	}

	dataType := editData.DataType

	secondStmt := "UPDATE " + dataType + " SET data = $1 WHERE id = $2"
	_, err = dbData.DatabaseConnection.ExecContext(ctx, secondStmt, editData.Data, editData.StaticID)
	if err != nil {
		return err
	}

	return nil
}

func dataAccess(ctx context.Context, dbData PostgreDB, id string, dataType string) (string, error) {
	stmt := "SELECT data, sk FROM " + dataType + " WHERE id = $1"
	var data, sk string
	err := dbData.DatabaseConnection.QueryRowContext(ctx, stmt, id).Scan(&data, &sk)
	if err != nil {
		log.Printf("Failed to find correlated data")
		return "", err
	}

	decryptedData, err := encryption.DecryptData(sk, data)
	if err != nil {
		log.Printf("Failed to check if data is valid")
		return "", err
	}

	return decryptedData, nil
}
