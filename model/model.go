package model

import "time"

type AccountVerificationRequest struct {
	AccountNumber   string `json:"account_number"`
	ReferenceNumber string `json:"reference_number"`
}

type AccountVerificationRespond struct {
	ReferenceNumber    string `json:"reference_number"`
	FullAccountNumber  string `json:"full_account_number"`
	CustomerNumber     string `json:"customer_number"`
	CustomerName       string `json:"customer_name"`
	Gender             string `json:"gender"`
	AccountType        string `json:"account_type"`
	AccountStatus      string `json:"account_status"`
	AccountBalance     string `json:"account_balance"`
	AccountRestriction string `json:"account_restriction"`
	AccountCurrency    string `json:"account_currency"`
	AvailabalBalance   string `json:"available_balance"`
	MobileNumber       string `json:"mobile_number"`
	AccountEmail       string `json:"account_email"`
	TotalBlockedFund   string `json:"total_blocked_fund"`
	NationalNumber     string `json:"national_number"`
	Message            string `json:"massager"`
}

type AccountBalanceRequest struct {
	AccountNumber string `json:"accountNumber"`
	RequestID     string `json:"requestId"`
}

type AccountBalanceResponse struct {
	AccountBalance float64 `json:"accountBalance"`
	Currency       string  `json:"currency"`
	AccountNumber  string  `json:"accountNumber"`
	AccountName    string  `json:"accountName"`
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
