package model

import "time"

type AccountVerificationRequest struct {
	AccountNumber   string `json:"accountNumber"`
	ReferenceNumber string `json:"referenceNumber"`
}

type AccountVerificationRespond struct {
	CustomerNumber  string  `json:"customerNumber"`
	NationalNumber  string  `json:"nationalNumber"`
	AccountCurrency string  `json:"accountCurrency"`
	AccountBalance  float64 `json:"accountBalance"`
	CustomerName    string  `json:"customerName,omitempty"`
	MobileNumber    string  `json:"mobileNumber,omitempty"`
}

type AccountBalanceRequest struct {
	AccountNumber string `json:"accountNumber"`
	RequestID     string `json:"requestId"`
}

type AccountBalanceResponse struct {
	AccountBalance float64 `json:"accountBalance"`
	Currency       string  `json:"currency"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ActivityLog struct {
	UserID     string `json:"userId"`
	LogMessage string `json:"logMessage"`
}

type User struct {
	ID             string `json:"id,omitempty"`
	Password       string `json:"password,omitempty"`
	CustomerNumber string `json:"customerNumber"`
	NationalID     string `json:"nationalId"`
	PhoneNumber    string `json:"phoneNumber"`
}

type Account struct {
	ID            int    `json:"id"`
	AccountNumber string `json:"accountNumber"`
}

type RequestLog struct {
	RequestID          string                 `json:"requestId"`
	UserID             string                 `json:"userId"`
	RequestMethod      string                 `json:"requestMethod"`
	RequestPath        string                 `json:"requestPath"`
	RequestQuery       string                 `json:"requestQuery"`
	RequestBody        interface{}            `json:"requestBody"`
	ResponseStatusCode int                    `json:"responseStatusCode"`
	ResponseBody       interface{}            `json:"responseBody"`
	ResponseHeaders    map[string]interface{} `json:"responseHeaders"`
}

type TransactionModel struct {
	TransactionID              int       `json:"transactionId"`
	UserID                     string    `json:"userId"`
	AccountID                  int       `json:"accountId"`
	TransactionType            string    `json:"transactionType"`
	Amount                     float64   `json:"amount"`
	TransactionDate            time.Time `json:"transactionDate"`
	Description                string    `json:"description"`
	TransactionTo              string    `json:"transactionTo"`
	TransactionReferenceNumber string    `json:"transactionReferenceNumber"`
}
