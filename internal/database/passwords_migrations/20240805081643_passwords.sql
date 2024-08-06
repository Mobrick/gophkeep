-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS passwords(
    id    TEXT PRIMARY KEY,
	data  TEXT NOT NULL,
    sk    TEXT NOT NULL,
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS passwords;
-- +goose StatementEnd
