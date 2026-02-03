package setup

import "os"

const (
	ACCOUNT_VERIFICATION_URL = "http://172.20.1.13:2073/api/v1"
	ACCOUNT_VERIFICATION_KEY = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJRCI6IjE4OWU1NzkzLWFhYjYtNGFhNy1iM2RmLWUxYzMxZTg3MTdmNCJ9.eQJQ76SQ1fJ4iJ7mr8D6m4ursO6Glbel3U1GYBNKWXs"
)

const (
	BASE_URL      = "http://172.20.1.66:8087/realbus_api/ax/"
	CHANNEL_CODE  = "CS107"
	SECURITY_CODE = "aJS#5KS$5zBAkY9"
	TIPS_URL      = "http://172.20.1.26:8087/tips_api/ib/"
	TIPS_CHANNEL  = "MA250"
	TIPS_PASSWORD = "Ma#TY90@WeYHg.24"
	// TIPS_URL_QR      = "http://172.20.1.63:8087/tips_api/ib/"
	// TIPS_CHANNEL_QR  = "IB450"
	// TIPS_PASSWORD_QR = "IB.pASS@Pbz.2023"

	TIPS_URL_QR      = "http://172.20.1.26:8087/tips_api/ib/"
	TIPS_CHANNEL_QR  = "MA250"
	TIPS_PASSWORD_QR = "Ma#TY90@WeYHg.24"
)

func DatabaseDriver() string {
	if driver := os.Getenv("DB_DRIVER"); driver != "" {
		return driver
	}

	return "postgres"
}

func DatabaseDSN() string {
	return "user=pbz-airpay  password=pbz@Admin-air123  dbname=pbz-airpay  host=172.20.1.69 port=7020 sslmode=disable"
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
