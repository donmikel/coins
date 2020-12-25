# Coins Payment Service

This is the system that stores payment and methods to send money between accounts. Final Payment API  you can see the [OpenAPI Specification](https://github.com/donmikel/coins/blob/main/api/coins-openapi.yaml),
this will help to understand the ways of using.

### Installing

To install from source, clone the Payment git repository:

```shell script
git clone git@github.com:donmikel/coins.git
cd ./coins
```

Run `docker-compose up` and Compose will start the entire app:

```shell script
cd ./deployments
docker-compose up -d
```

### Test and linters

Run all tests from root of project :

```shell script
make test
```

linters

before [install](https://golangci-lint.run/usage/install/#local-installation) golangci-lint 
```shell script
golangci-lint run -c .golangci.yml
```

### Examples:

Get all accounts 

```shell script
curl --request GET \
  --url http://localhost:8080/api/v1/accounts
```

Get all payments

```shell script
curl --request GET \
  --url 'http://localhost:8080/api/v1/payments'
```

Send payment

```shell script
curl --request POST \
  --url 'http://localhost:8080/api/v1/payments' \
  --header 'content-type: application/json' \
  --data '{"from_account":"bob123", "to_account":"alice456", "amount":"100", "direction": 1}'
```

You see something like

```json
[
  {
    "id":1,
    "from_account":"bob123",
    "amount":"100",
    "to_account":"alice456",
    "direction":1,
    "dt":"2020-12-25T22:41:58.401358Z"
  }
]
```

# Data structure

Basic type that uses in payment service:
 - Payment (`payment.Payment`)
 - Account (`account.Account`)

## Payment

Payment struct contains of uniq ID of payment transaction, source and destination account, amount of sending money, direction and operation time.

```go
type Payment struct {
	ID          uint64          `json:"id" db:"id"`
	FromAccount string          `json:"from_account" db:"from_account"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	ToAccount   string          `json:"to_account" db:"to_account"`
	Direction   Direction       `json:"direction" db:"direction"`
	Dt          *time.Time      `json:"dt" db:"dt"`
}
```

## Account

Account struct contains of uniq ID, balance and currency.
Send payment only allowed between accounts with the same currency.
```go
type Account struct {
	ID       string          `json:"id" db:"id"`
	Balance  decimal.Decimal `json:"balance" db:"balance"`
	Currency string          `json:"currency" db:"currency"`
}
```
