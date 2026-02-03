package helper

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/leopardquick/zssf/model"
)

func GenerateReferenceNumber() string {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}

	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func InsertActivityLog(entry model.ActivityLog) {
	log.Printf("activity_log user=%s msg=%s", entry.UserID, entry.LogMessage)
}

type DBHelper struct {
	logger *log.Logger
	db     *sql.DB
}

func NewDBHelper(logger *log.Logger, db *sql.DB) *DBHelper {
	if logger == nil {
		logger = log.Default()
	}

	return &DBHelper{logger: logger, db: db}
}

func (h *DBHelper) InsertActivityLog(entry model.ActivityLog) {
	InsertActivityLog(entry)
}

func (h *DBHelper) GetAccountsByUserID(userID string) ([]model.Account, error) {
	return nil, errors.New("GetAccountsByUserID is not implemented")
}

func (h *DBHelper) GetUserByID(userID string) (model.User, error) {
	return model.User{}, errors.New("GetUserByID is not implemented")
}

func (h *DBHelper) GetLoginAttempts(userID string) (int, error) {
	return 0, errors.New("GetLoginAttempts is not implemented")
}

func (h *DBHelper) ResetLoginAttempts(userID string) error {
	return errors.New("ResetLoginAttempts is not implemented")
}

func (h *DBHelper) VerifyUser(accountNumber string) (model.AccountVerificationRespond, error) {
	return model.AccountVerificationRespond{}, errors.New("VerifyUser is not implemented")
}

func (h *DBHelper) ContainsAccountNumber(accountNumber string, accounts []model.Account) bool {
	return ContainsAccountNumber(accountNumber, accounts)
}

func (h *DBHelper) UpdateUserStatus(userID, status string) error {
	return UpdateUserStatus(userID, status)
}

func (h *DBHelper) DecryptPassword(pin, encryptedPassword string) error {
	return DecryptPassword(pin, encryptedPassword)
}

func (h *DBHelper) InsertRequestLog(entry model.RequestLog) {
	InsertRequestLog(entry)
}

func (h *DBHelper) InsertTransactionModel(entry model.TransactionModel) error {
	return InsertTransactionModel(entry)
}

func ContainsAccountNumber(accountNumber string, accounts []model.Account) bool {
	for _, account := range accounts {
		if account.AccountNumber == accountNumber {
			return true
		}
	}
	return false
}

func UpdateUserStatus(userID, status string) error {
	return errors.New("UpdateUserStatus is not implemented")
}

func DecryptPassword(pin, encryptedPassword string) error {
	return errors.New("DecryptPassword is not implemented")
}

func InsertRequestLog(entry model.RequestLog) {
	log.Printf("request_log request_id=%s status=%d", entry.RequestID, entry.ResponseStatusCode)
}

func InsertTransactionModel(entry model.TransactionModel) error {
	return errors.New("InsertTransactionModel is not implemented")
}
