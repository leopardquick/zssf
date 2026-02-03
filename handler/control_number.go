package handler

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"net/http"
	"time"

	"github.com/leopardquick/zssf/helper"
	"github.com/leopardquick/zssf/model"
	"github.com/leopardquick/zssf/setup"
	"github.com/leopardquick/zssf/store"
)

type ControlNumberHandler struct {
	Client      *http.Client
	RequestLogs store.RequestLogStore
	L           errorLogger
	db          *sql.DB
}

func NewControlNumberHandler(client *http.Client, requestLogs store.RequestLogStore) *ControlNumberHandler {
	if client == nil {
		client = http.DefaultClient
	}

	return &ControlNumberHandler{
		Client:      client,
		RequestLogs: requestLogs,
		L:           stdErrorLogger{Logger: log.Default()},
	}
}

type errorLogger interface {
	Error(args ...any)
}

type stdErrorLogger struct {
	Logger *log.Logger
}

func (l stdErrorLogger) Error(args ...any) {
	if l.Logger == nil {
		log.Print(args...)
		return
	}
	l.Logger.Print(args...)
}

type contextKey string

const (
	userKey           contextKey = "user"
	accountKey        contextKey = "account"
	deviceUniqueIDKey contextKey = "deviceUniqueID"
)

func (cn *ControlNumberHandler) Enquire(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "unknown"
	}

	requestBodyBytes, _ := io.ReadAll(r.Body)
	requestBodyJSON := normalizeJSON(requestBodyBytes)
	requestHeadersJSON := mustJSON(headerToMap(r.Header))

	if !json.Valid(requestBodyBytes) {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	var apiRequestEnquire model.ApiRequestEnquire

	err := json.Unmarshal(requestBodyBytes, &apiRequestEnquire)

	if err != nil {
		cn.L.Error("error decoding request body", err)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	apiRequestEnquire.RequestID = "PBZAPP" + fmt.Sprintf("%d", time.Now().UnixNano()) + "CN-" + apiRequestEnquire.ControlNo
	// check if request id is empty

	if apiRequestEnquire.RequestID == "" {
		cn.L.Error("request id is empty")
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "request id is empty"})
		return
	}

	// check if control number is empty

	if apiRequestEnquire.ControlNo == "" {
		cn.L.Error("control number is empty")
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "control number is empty"})
		return

	}

	security_code, err := cn.GenerateSecurityCode(setup.CHANNEL_CODE, apiRequestEnquire.RequestID, setup.SECURITY_CODE)

	if err != nil {
		cn.L.Error("error generating security code", err)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "Operation failed"})
		return
	}

	if cn.RequestLogs == nil {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "request log store is not configured"})
		return
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	defer client.CloseIdleConnections()

	url := setup.BASE_URL + "bill/query"

	enquireRequest := model.EnquireRequest{
		ControlNo:    apiRequestEnquire.ControlNo,
		RequestId:    apiRequestEnquire.RequestID,
		ChannelCode:  setup.CHANNEL_CODE,
		SecurityCode: security_code,
	}

	enquireRequestJson, err := json.Marshal(enquireRequest)

	if err != nil {
		cn.L.Error("error creating request", err)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "Operation failed"})
		return
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(enquireRequestJson))

	if err != nil {
		cn.L.Error("error creating request", err)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "Operation failed"})
		return
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)

	if err != nil {
		cn.L.Error("error sending request", err)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "Operation failed"})
		return
	}

	defer response.Body.Close()

	var enquireResponse model.EnquireResponse

	err = json.NewDecoder(response.Body).Decode(&enquireResponse)

	if err != nil {
		cn.L.Error("error decoding response body", err)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "Operation failed"})
		return
	}

	if enquireResponse.StatusId != "2000" {

		if enquireResponse.StatusMessage == "" {
			base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
			respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "OPERATION FAILED"})
			return
		}
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, enquireResponse)
		return
	}

	// insert into activity log
	base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, apiRequestEnquire.RequestID, userID)
	respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusOK, enquireResponse)
}

func (cn *ControlNumberHandler) PaymentPost(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userKey).(string)
	if userID == "" {
		userID = r.Header.Get("X-User-Id")
		if userID == "" {
			userID = "known"
		}
	}

	requestBodyBytes, _ := io.ReadAll(r.Body)
	requestBodyJSON := normalizeJSON(requestBodyBytes)
	requestHeadersJSON := mustJSON(headerToMap(r.Header))

	if !json.Valid(requestBodyBytes) {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	// control number payment request
	var apiPaymentRequest model.PaymentRequestApi

	err := json.Unmarshal(requestBodyBytes, &apiPaymentRequest)

	if err != nil {
		cn.L.Error("error decoding request body", err)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	// decode request body

	// check if control number is empty
	if apiPaymentRequest.ControlNo == "" {
		cn.L.Error("control number is empty")
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "control number is empty"})
		return
	}

	// set payment Method to MA for mobile app
	apiPaymentRequest.PaymentMethod = "MA"
	// set CBFlag to 1 for service to charge customer directly
	apiPaymentRequest.CBFlag = "1"
	// set CLFlag to 1 for service to charge customer directly
	apiPaymentRequest.CLFlag = "1"

	// check if user id is empty
	if userID == "" || userID == "unknown" {
		cn.L.Error("user id is empty")
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "user id is empty"})
		return
	}

	// create a new db helper
	helper := helper.NewDBHelper(nil, cn.db)

	// insert into activity log

	go helper.InsertActivityLog(
		model.ActivityLog{
			UserID:     userID,
			LogMessage: "Payment post for control number " + apiPaymentRequest.ControlNo,
		},
	)

	requestId := "PBZAPP" + fmt.Sprintf("%d", time.Now().UnixNano())

	// check if request id is empty

	security_code, err := cn.GenerateSecurityCode(setup.CHANNEL_CODE, requestId, setup.SECURITY_CODE)

	if err != nil {
		cn.L.Error("error generating security code", err)
		go helper.InsertActivityLog(
			model.ActivityLog{
				UserID:     userID,
				LogMessage: "Error generating security code-" + err.Error(),
			},
		)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "error generating security code"})
		return
	}

	payment := model.PaymentRequest{
		ControlNo:      apiPaymentRequest.ControlNo,
		RequestID:      requestId,
		ChannelCode:    setup.CHANNEL_CODE,
		SecurityCode:   security_code,
		VDResponseID:   apiPaymentRequest.VDResponseID,
		PayerName:      apiPaymentRequest.PayerName,
		MobileNo:       apiPaymentRequest.MobileNo,
		Email:          apiPaymentRequest.Email,
		DebitAccount:   apiPaymentRequest.DebitAccount,
		CreditAccount:  apiPaymentRequest.CreditAccount,
		Amount:         apiPaymentRequest.Amount,
		Currency:       apiPaymentRequest.Currency,
		PaymentMethod:  apiPaymentRequest.PaymentMethod,
		PSPReferenceID: apiPaymentRequest.PSPReferenceID,
		CBFlag:         apiPaymentRequest.CBFlag,
		CLFlag:         apiPaymentRequest.CLFlag,
	}

	// check if payer name is empty

	if payment.PayerName == "" {
		payment.PayerName = "Not Provided"
	}

	// check if mobile number is empty

	if payment.MobileNo == "" {
		payment.MobileNo = "Not Provided"
	}

	client := &http.Client{
		Timeout: 40 * time.Second,
	}

	defer client.CloseIdleConnections()

	url := setup.BASE_URL + "payment/post"

	paymentRequestJson, err := json.Marshal(payment)

	if err != nil {

		go helper.InsertActivityLog(
			model.ActivityLog{
				UserID:     userID,
				LogMessage: "Error creating request-" + err.Error(),
			},
		)
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "failed"})
		return
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(paymentRequestJson))

	if err != nil {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "failed "})
		return
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)

	if err != nil {

		// insert into request logs
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "failed"})

		return
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "Operation failed"})

		return
	}

	var paymentResponse model.ControlNumberPaymentResponse

	errs := json.Unmarshal(responseBody, &paymentResponse)

	if errs != nil {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "Operation failed"})
		return
	}

	// insert into request logs

	if paymentResponse.StatusId != "2000" {

		if paymentResponse.StatusMessage == "" {
			base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
			respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusInternalServerError, model.ErrorResponse{Error: "OPERATION FAILED"})
			return
		}
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
		respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: paymentResponse.StatusMessage})
		return
	}
	base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestId, userID)
	respondWithLog(&Handler{RequestLogs: cn.RequestLogs}, w, r, base, http.StatusOK, paymentResponse)

}

func (cn *ControlNumberHandler) GenerateSecurityCode(channelCode, requestID, channelPassword string) (string, error) {
	// Concatenate ChannelCode, RequestID, and Base64(ChannelPassword)
	inputString := channelCode + requestID + base64.StdEncoding.EncodeToString([]byte(channelPassword))

	// Calculate SHA256 hash
	hasher := sha256.New()
	hasher.Write([]byte(inputString))
	hash := hasher.Sum(nil)

	// Encode the hash in Base64
	//hashBase64 := base64.StdEncoding.EncodeToString(hash)
	hashBase64String := base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(hash)))

	// fmt.Printf("channelcode: %v \n", channelCode)
	// fmt.Printf("requestid: %v  \n", requestID)
	// fmt.Printf("channelpassword: %v \n", channelPassword)
	// fmt.Printf("hashedpassword: %v \n", base64.StdEncoding.EncodeToString([]byte(channelPassword)))
	// fmt.Printf("beforesha256: %v \n", inputString)
	// fmt.Printf("sha256: %v \n", hex.EncodeToString(hash))
	// fmt.Printf("sha256base64: %v \n", hashBase64)
	// fmt.Printf("sha256base64string: %v \n", hashBase64String)

	return hashBase64String, nil
}

func ResponseWithError(w http.ResponseWriter, code int, message string) {
	ResponseWithJSON(w, code, model.ErrorResponse{Error: message})
}

func ResponseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(wrapResponse(code, payload))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
