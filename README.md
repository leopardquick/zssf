# ZSSF Service

HTTP service for account balance and control number operations with request logging.

## Requirements

- Go 1.20+ (any recent Go version should work)
- PostgreSQL

## Configuration

Set the following environment variables:

- `DATABASE_URL` (required): PostgreSQL connection string.
- `DB_DRIVER` (optional): SQL driver (default: `postgres`).
- `BASE_URL` (optional): Base URL for the control number API (default: `https://example.com/`).
- `CHANNEL_CODE` (required for control number endpoints).
- `SECURITY_CODE` (required for control number endpoints).

See [setup/setup.go](setup/setup.go) for defaults.

## Database

Request logs are stored in a `request_logs` table. Apply the migration in:

- [migrations/20260202120000_create_request_logs.sql](migrations/20260202120000_create_request_logs.sql)

## Running the service

```
export DATABASE_URL="postgres://user:pass@localhost:5432/zssf?sslmode=disable"
export CHANNEL_CODE="..."
export SECURITY_CODE="..."
export BASE_URL="https://example.com/"

# run
 go run ./...
```

Server listens on `:8080`.

## Endpoints

### Health check

- `GET /healthz`

Response:

```
ok
```

### Root

- `GET /`

Response:

```
hello from chi
```

### Account balance

- `POST /account-balance`
- Header: `X-User-Id` (optional; defaults to `unknown`)

Request body:

```
{
  "accountNumber": "1234567890",
  "requestId": "abc123"
}
```

Success response:

```
{
  "statusCode": 200,
  "data": {
    "accountBalance": 1000.5,
    "currency": "Tshs "
  }
}
```

Error response:

```
{
  "statusCode": 400,
  "error": "message"
}
```

### Control number enquire

- `POST /control-number/enquire`
- Header: `X-User-Id` (optional; defaults to `unknown`)

Request body:

```
{
  "control_number": "123456789012",
  "account_number": "001234567890"
}
```

Success response example:

```
{
  "statusCode": 200,
  "data": {
    "statusId": "2000",
    "statusMessage": "OK",
    "data": {
      "controlNo": "123456789012",
      "billDescription": "Bill description",
      "requestId": "...",
      "apiResponseId": 1,
      "vdResponseId": "...",
      "apiResponseDate": "...",
      "payerName": "John Doe",
      "mobileNo": "255700000000",
      "email": "john@example.com",
      "gatewayCode": "...",
      "gatewayName": "...",
      "gatewayRefId": "...",
      "spCode": "...",
      "spName": "...",
      "creditAccount": "...",
      "amount": "1000",
      "currency": "TZS",
      "minAmount": "1000",
      "paymentPlan": "...",
      "paymentOption": "...",
      "billExpireDate": "..."
    }
  }
}
```

### Control number payment

- `POST /control-number/payment`
- Header: `X-User-Id` (required or provided by auth middleware)

Request body:

```
{
  "controlNo": "123456789012",
  "vdResponseId": "...",
  "payerName": "John Doe",
  "mobileNo": "255700000000",
  "email": "john@example.com",
  "debitAccount": "001234567890",
  "creditAccount": "009876543210",
  "amount": "1000",
  "currency": "TZS",
  "paymentMethod": "MA",
  "pspReferenceId": "...",
  "cbFlag": "1",
  "clFlag": "1",
  "pin": "..."
}
```

Success response example:

```
{
  "statusCode": 200,
  "data": {
    "statusId": "2000",
    "statusMessage": "OK",
    "data": {
      "controlNo": "123456789012",
      "requestId": "...",
      "apiResponseId": 1,
      "apiResponseDate": "...",
      "debitAccount": "001234567890",
      "creditAccount": "009876543210",
      "amount": 1000,
      "currency": "TZS",
      "gatewayRefId": "...",
      "receiptNo": "..."
    }
  }
}
```

## Request logging

Each request/response is persisted to `request_logs` via the request log store. Errors are written to the activity log helper in a goroutine.
