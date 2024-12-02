-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS files(
    id    TEXT PRIMARY KEY,
	data  VARCHAR NOT NULL,
    sk    TEXT NOT NULL
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cards;
-- +goose StatementEnd
