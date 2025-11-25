package domain

import "errors"

var (
	ErrAssetExists        = errors.New("wallet: asset already registered")
	ErrAssetNotFound      = errors.New("wallet: asset not found")
	ErrAccountExists      = errors.New("wallet: wallet account already exists")
	ErrAccountNotFound    = errors.New("wallet: wallet account not found")
	ErrBlockNotFound      = errors.New("wallet: block not found")
	ErrDepositNotFound    = errors.New("wallet: deposit record not found")
	ErrWithdrawalNotFound = errors.New("wallet: withdrawal record not found")
	ErrInsufficientFunds  = errors.New("wallet: insufficient funds")
	ErrWithdrawalRejected = errors.New("wallet: withdrawal rejected by risk control")
	ErrWithdrawalPending  = errors.New("wallet: withdrawal already in pending state")
	ErrSignerUnavailable  = errors.New("wallet: signer unavailable")
)
