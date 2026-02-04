package store

import (
	"context"
	"database/sql"
	"errors"
)

type AccountStore interface {
	ExistsByAccountNumber(ctx context.Context, accountNumber string) (bool, error)
}

type SQLAccountStore struct {
	DB *sql.DB
}

func NewSQLAccountStore(db *sql.DB) *SQLAccountStore {
	return &SQLAccountStore{DB: db}
}

func (s *SQLAccountStore) ExistsByAccountNumber(ctx context.Context, accountNumber string) (bool, error) {
	if s == nil || s.DB == nil {
		return false, errors.New("db is not configured")
	}

	row := s.DB.QueryRowContext(ctx, `SELECT 1 FROM accounts WHERE account_number = $1 LIMIT 1`, accountNumber)
	var found int
	if err := row.Scan(&found); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
