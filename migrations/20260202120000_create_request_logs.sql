-- +goose Up
CREATE TABLE request_logs (
	id SERIAL PRIMARY KEY,
	user_id VARCHAR(255) NOT NULL  REFERENCES users(user_id) ON DELETE CASCADE,
	request_id VARCHAR(255) NOT NULL UNIQUE,
	request_method VARCHAR(10) NOT NULL,
	request_path VARCHAR(255) NOT NULL,
	request_query VARCHAR(255) NOT NULL,
	request_body JSONB NOT NULL,
	request_headers JSONB NOT NULL,
	response_status_code INTEGER NOT NULL,
	response_body JSONB NOT NULL,
	response_headers JSONB NOT NULL,
	request_receipt VARCHAR(500) NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS request_logs;
