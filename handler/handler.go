package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/leopardquick/zssf/helper"
	"github.com/leopardquick/zssf/model"
	"github.com/leopardquick/zssf/setup"
	"github.com/leopardquick/zssf/store"
)

type Handler struct {
	Client      *http.Client
	RequestLogs store.RequestLogStore
}

func New(client *http.Client, requestLogs store.RequestLogStore) *Handler {
	if client == nil {
		client = http.DefaultClient
	}

	return &Handler{
		Client:      client,
		RequestLogs: requestLogs,
	}
}

func (h *Handler) AccountBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "unknown"
	}

	requestBodyBytes, _ := io.ReadAll(r.Body)
	requestBodyJSON := normalizeJSON(requestBodyBytes)
	requestHeadersJSON := mustJSON(headerToMap(r.Header))
	requestMethod := r.Method
	requestPath := r.URL.Path
	requestQuery := r.URL.RawQuery
	if !json.Valid(requestBodyBytes) {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		base.RequestMethod = requestMethod
		base.RequestPath = requestPath
		base.RequestQuery = requestQuery
		respondWithLog(h, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request payload"})
		return
	}

	var accountBalanceRequest model.AccountBalanceRequest
	if err := json.Unmarshal(requestBodyBytes, &accountBalanceRequest); err != nil {
		base := buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, helper.GenerateReferenceNumber(), userID)
		base.RequestMethod = requestMethod
		base.RequestPath = requestPath
		base.RequestQuery = requestQuery
		respondWithLog(h, w, r, base, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request payload"})
		return
	}

	if accountBalanceRequest.AccountNumber == "" {
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, accountBalanceRequest.RequestID, ""), http.StatusBadRequest, model.ErrorResponse{Error: "account number is required"})
		return
	}

	if accountBalanceRequest.RequestID == "" {
		generatedID := helper.GenerateReferenceNumber()
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, generatedID, ""), http.StatusBadRequest, model.ErrorResponse{Error: "requestId is required"})
		return
	}

	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}

	if h.RequestLogs == nil {
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, accountBalanceRequest.RequestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "request log store is not configured"})
		return
	}

	requestID := accountBalanceRequest.RequestID
	requestLog, err := h.RequestLogs.GetByRequestID(r.Context(), requestID)
	if err != nil {
		if !errors.Is(err, store.ErrRequestLogNotFound) {
			go helper.InsertActivityLog(model.ActivityLog{
				UserID:     userID,
				LogMessage: "Account balance request failed to read request log error : " + err.Error(),
			},
			)
			respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "failed to process request"})
			return
		}
	} else if requestLog.RequestID != "" {
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusConflict, model.ErrorResponse{Error: "request already used"})
		return
	}

	var accountVerification model.AccountVerificationRequest
	referenceNumber := helper.GenerateReferenceNumber()

	accountVerification.AccountNumber = accountBalanceRequest.AccountNumber
	accountVerification.ReferenceNumber = referenceNumber

	accountVerificationBytes, err := json.Marshal(accountVerification)

	if err != nil {
		fmt.Println("Error marshalling account verification request", err)
		// insert into activity log table in a go routine if their is error fmt.Println(err)
		go helper.InsertActivityLog(model.ActivityLog{
			UserID:     userID,
			LogMessage: "Account balance request failed to marshal account verification request error : " + err.Error(),
		},
		)
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get account balance"})
		return
	}

	req, err := http.NewRequest(http.MethodPost, setup.ACCOUNT_VERIFICATION_URL+"/service1/account-verification", bytes.NewBuffer(accountVerificationBytes))

	if err != nil {
		fmt.Println("Error creating account verification request", err)
		// insert into activity log table in a go routine if their is error fmt.Println(err)
		go helper.InsertActivityLog(model.ActivityLog{
			UserID:     userID,
			LogMessage: "Account balance request failed to create account verification reques error : " + err.Error(),
		},
		)
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get account balance"})
		return
	}

	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"x-request-id":  []string{referenceNumber},
		"Authorization": []string{setup.ACCOUNT_VERIFICATION_KEY},
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error getting account balance", err)
		// insert into activity log table in a go routine if their is error fmt.Println(err)
		go helper.InsertActivityLog(model.ActivityLog{
			UserID:     userID,
			LogMessage: "Account balance request failed to get account balance error : " + err.Error(),
		},
		)
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "request error service unvailable"})
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		var errorResponse model.ErrorResponse

		err = json.NewDecoder(resp.Body).Decode(&errorResponse)

		if err != nil {
			// insert into activity log table in a go routine if their is error fmt.Println(err)
			go helper.InsertActivityLog(model.ActivityLog{
				UserID:     userID,
				LogMessage: "Account balance request failed to decode error response error : " + err.Error(),
			},
			)
			respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "Error gettting account balance"})
			return
		}

		// if status code is  bad request
		go helper.InsertActivityLog(model.ActivityLog{
			UserID:     userID,
			LogMessage: "Account balance request failed- " + errorResponse.Error,
		},
		)

		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: errorResponse.Error})
		return

	}

	var accountVerificationRespond model.AccountVerificationRespond

	err = json.NewDecoder(resp.Body).Decode(&accountVerificationRespond)

	if err != nil {
		// insert into activity log table in a go routine if their is error fmt.Println(err)
		go helper.InsertActivityLog(model.ActivityLog{
			UserID:     userID,
			LogMessage: "Account balance request failed to decode account verification response error : " + err.Error(),
		},
		)
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "Error verifying registration"})
		return
	}

	currency := "$ "

	if accountVerificationRespond.AccountCurrency == "1 TZS" {
		currency = "TZS "
	}

	// requires adding "strconv" and "strings" to imports
	raw := strings.TrimSpace(accountVerificationRespond.AccountBalance)
	raw = strings.ReplaceAll(raw, ",", "")
	raw = strings.TrimPrefix(raw, "$")
	raw = strings.TrimPrefix(raw, "Tshs")
	raw = strings.TrimSpace(raw)

	balance, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		go helper.InsertActivityLog(model.ActivityLog{
			UserID:     userID,
			LogMessage: "Account balance request failed to parse account balance error : " + err.Error(),
		})
		respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusInternalServerError, model.ErrorResponse{Error: "invalid account balance format"})
		return
	}

	accountBalance := model.AccountBalanceResponse{
		AccountBalance: balance,
		Currency:       currency,
		AccountNumber:  accountBalanceRequest.AccountNumber,
	}

	respondWithLog(h, w, r, buildRequestLogBase(r, requestBodyJSON, requestHeadersJSON, requestID, userID), http.StatusOK, accountBalance)

	//send sms to user in a go routine

	//remove the plus sign from the phone number
	// number := user.PhoneNumber[1:]

	// //replace space with plus sign in time now string
	// timeNow := strings.ReplaceAll(time.Now().Format("2006-01-02 15:04:05"), " ", "+")

	// go helper.SendSMS(number, "Salio+lako+la+"+accountBalanceRequest.AccountNumber+"+ni+"+formatCurrency(accountBalance.AccountBalance)+".+"+timeNow+"+Tuma+Pesa+kwa+urahisi+na+PBZ+APP")

	// insert into activity log table in a go routine if their is error fmt.Println(err)

	go helper.InsertActivityLog(model.ActivityLog{
		UserID:     userID,
		LogMessage: "Account balance generated successfully",
	},
	)

}

func buildRequestLogBase(r *http.Request, requestBodyJSON []byte, requestHeadersJSON []byte, requestID string, userID string) store.RequestLog {
	if requestID == "" {
		requestID = helper.GenerateReferenceNumber()
	}

	return store.RequestLog{
		RequestID:      requestID,
		UserID:         userID,
		RequestMethod:  r.Method,
		RequestPath:    r.URL.Path,
		RequestQuery:   r.URL.RawQuery,
		RequestBody:    requestBodyJSON,
		RequestHeaders: requestHeadersJSON,
		RequestReceipt: requestID,
	}
}

func respondWithLog(h *Handler, w http.ResponseWriter, r *http.Request, base store.RequestLog, status int, payload any) {
	responseBody := mustJSON(wrapResponse(status, payload))
	responseHeaders := mustJSON(headerToMap(http.Header{"Content-Type": []string{"application/json"}}))

	base.ResponseStatusCode = status
	base.ResponseBody = responseBody
	base.ResponseHeaders = responseHeaders
	if h != nil && h.RequestLogs != nil {
		logErr := h.RequestLogs.Create(r.Context(), base)
		if logErr != nil {
			go helper.InsertActivityLog(model.ActivityLog{
				UserID:     base.UserID,
				LogMessage: "Account balance request failed to write request log error : " + logErr.Error(),
			},
			)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(responseBody)
}

type responseEnvelope struct {
	StatusCode int    `json:"statusCode"`
	Data       any    `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
}

func wrapResponse(status int, payload any) responseEnvelope {
	if payload == nil {
		return responseEnvelope{StatusCode: status}
	}

	switch v := payload.(type) {
	case responseEnvelope:
		if v.StatusCode == 0 {
			v.StatusCode = status
		}
		return v
	case *responseEnvelope:
		if v == nil {
			return responseEnvelope{StatusCode: status}
		}
		if v.StatusCode == 0 {
			v.StatusCode = status
		}
		return *v
	case model.ErrorResponse:
		return responseEnvelope{StatusCode: status, Error: v.Error}
	case *model.ErrorResponse:
		if v == nil {
			return responseEnvelope{StatusCode: status}
		}
		return responseEnvelope{StatusCode: status, Error: v.Error}
	default:
		return responseEnvelope{StatusCode: status, Data: payload}
	}
}

func mustJSON(payload any) []byte {
	if payload == nil {
		return []byte("null")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		fallback, _ := json.Marshal(model.ErrorResponse{Error: "internal server error"})
		return fallback
	}

	return data
}

func normalizeJSON(body []byte) []byte {
	if len(body) == 0 {
		return []byte("{}")
	}

	if json.Valid(body) {
		return body
	}

	fallback, _ := json.Marshal(map[string]string{"raw": string(body)})
	return fallback
}

func headerToMap(header http.Header) map[string][]string {
	output := make(map[string][]string, len(header))
	for key, values := range header {
		copied := make([]string, len(values))
		copy(copied, values)
		output[key] = copied
	}
	return output
}

func (h *Handler) ResponseWithError(w http.ResponseWriter, status int, message string) {
	h.ResponseWithJSON(w, status, model.ErrorResponse{Error: message})
}

func (h *Handler) ResponseWithJSON(w http.ResponseWriter, status int, payload any) {
	response, err := json.Marshal(wrapResponse(status, payload))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(response)
}
