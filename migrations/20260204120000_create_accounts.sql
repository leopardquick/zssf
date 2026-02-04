-- +goose Up
CREATE TABLE accounts (
	id SERIAL PRIMARY KEY,
	account_number VARCHAR(255) NOT NULL UNIQUE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS accounts;
