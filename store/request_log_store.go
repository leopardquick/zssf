package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

var (
	ErrRequestLogNotFound      = errors.New("request log not found")
	ErrRequestLogAlreadyExists = errors.New("request log already exists")
)

type RequestLog struct {
	RequestID          string
	UserID             string
	RequestMethod      string
	RequestPath        string
	RequestQuery       string
	RequestBody        json.RawMessage
	RequestHeaders     json.RawMessage
	ResponseStatusCode int
	ResponseBody       json.RawMessage
	ResponseHeaders    json.RawMessage
	RequestReceipt     string
	CreatedAt          time.Time
}

type RequestLogStore interface {
	Create(ctx context.Context, log RequestLog) error
	GetByRequestID(ctx context.Context, requestID string) (RequestLog, error)
}

type SQLRequestLogStore struct {
	DB *sql.DB
}

func NewSQLRequestLogStore(db *sql.DB) *SQLRequestLogStore {
	return &SQLRequestLogStore{DB: db}
}

func (s *SQLRequestLogStore) Create(ctx context.Context, log RequestLog) error {
	if s == nil || s.DB == nil {
		return errors.New("db is not configured")
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO request_logs (
			user_id,
			request_id,
			request_method,
			request_path,
			request_query,
			request_body,
			request_headers,
			response_status_code,
			response_body,
			response_headers,
			request_receipt
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		log.UserID,
		log.RequestID,
		log.RequestMethod,
		log.RequestPath,
		log.RequestQuery,
		log.RequestBody,
		log.RequestHeaders,
		log.ResponseStatusCode,
		log.ResponseBody,
		log.ResponseHeaders,
		log.RequestReceipt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" {
			return ErrRequestLogAlreadyExists
		}
		return err
	}

	return nil
}

func (s *SQLRequestLogStore) GetByRequestID(ctx context.Context, requestID string) (RequestLog, error) {
	if s == nil || s.DB == nil {
		return RequestLog{}, errors.New("db is not configured")
	}

	row := s.DB.QueryRowContext(ctx, `
		SELECT request_id, user_id, request_method, request_path, request_query, request_body, request_headers,
			response_status_code, response_body, response_headers, request_receipt, created_at
		FROM request_logs
		WHERE request_id = $1
	`, requestID)

	var log RequestLog
	if err := row.Scan(
		&log.RequestID,
		&log.UserID,
		&log.RequestMethod,
		&log.RequestPath,
		&log.RequestQuery,
		&log.RequestBody,
		&log.RequestHeaders,
		&log.ResponseStatusCode,
		&log.ResponseBody,
		&log.ResponseHeaders,
		&log.RequestReceipt,
		&log.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RequestLog{}, ErrRequestLogNotFound
		}
		return RequestLog{}, err
	}

	return log, nil
}
