-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS infos(
    static_id TEXT PRIMARY KEY,
    dynamic_id TEXT NOT NULL,
    name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT 'empty',
    type TEXT NOT NULL,
    account_uuid TEXT NOT NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS infos;
-- +goose StatementEnd
