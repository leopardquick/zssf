package model

type ApiRequestEnquire struct {
	ControlNo string `json:"control_number"`
	RequestID string `json:"-"`
}

type EnquireRequest struct {
	ControlNo    string `json:"controlNo"`
	RequestId    string `json:"requestId"`
	ChannelCode  string `json:"channelCode"`
	SecurityCode string `json:"securityCode"`
}

type EnquireResponse struct {
	StatusId      string      `json:"statusId"`
	StatusMessage string      `json:"statusMessage"`
	Data          ApiResponse `json:"data"`
}

type ApiResponse struct {
	ControlNo       string `json:"controlNo"`
	BillDescription string `json:"billDescription"`
	RequestId       string `json:"requestId"`
	ApiResponseId   int    `json:"apiResponseId"`
	VDResponseID    string `json:"vdResponseId"`
	ApiResponseDate string `json:"apiResponseDate"`
	PayerName       string `json:"payerName"`
	MobileNo        string `json:"mobileNo"`
	Email           string `json:"email"`
	GatewayCode     string `json:"gatewayCode"`
	GatewayName     string `json:"gatewayName"`
	GatewayRefId    string `json:"gatewayRefId"`
	SpCode          string `json:"spCode"`
	SpName          string `json:"spName"`
	CreditAccount   string `json:"creditAccount"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	MinAmount       string `json:"minAmount"`
	PaymentPlan     string `json:"paymentPlan"`
	PaymentOption   string `json:"paymentOption"`
	BillExpireDate  string `json:"billExpireDate"`
}

type PaymentRequest struct {
	ControlNo      string `json:"controlNo"`
	RequestID      string `json:"requestId"`
	ChannelCode    string `json:"channelCode"`
	SecurityCode   string `json:"securityCode"`
	VDResponseID   string `json:"vdResponseId"`
	PayerName      string `json:"payerName"`
	MobileNo       string `json:"mobileNo"`
	Email          string `json:"email"`
	DebitAccount   string `json:"debitAccount"`
	CreditAccount  string `json:"creditAccount"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	PaymentMethod  string `json:"paymentMethod"`
	PSPReferenceID string `json:"pspReferenceId"`
	CBFlag         string `json:"cbFlag"`
	CLFlag         string `json:"clFlag"`
}

type PaymentRequestApi struct {
	ControlNo      string `json:"controlNo"`
	VDResponseID   string `json:"vdResponseId"`
	PayerName      string `json:"payerName"`
	MobileNo       string `json:"mobileNo"`
	Email          string `json:"email"`
	DebitAccount   string `json:"debitAccount"`
	CreditAccount  string `json:"creditAccount"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	PaymentMethod  string `json:"paymentMethod"`
	PSPReferenceID string `json:"pspReferenceId"`
	CBFlag         string `json:"cbFlag"`
	CLFlag         string `json:"clFlag"`
	Pin            string `json:"pin"`
}

type ControlNumberPaymentResponse struct {
	StatusId      string            `json:"statusId"`
	StatusMessage string            `json:"statusMessage"`
	Data          ControlNumberData `json:"data"`
}

type ControlNumberData struct {
	ControlNo       string `json:"controlNo"`
	RequestId       string `json:"requestId"`
	ApiResponseId   int    `json:"apiResponseId"`
	ApiResponseDate string `json:"apiResponseDate"`
	DebitAccount    string `json:"debitAccount"`
	CreditAccount   string `json:"creditAccount"`
	Amount          int    `json:"amount"`
	Currency        string `json:"currency"`
	GatewayRefId    string `json:"gatewayRefId,omitempty"`
	ReceiptNo       string `json:"receiptNo,omitempty"`
}
