package myerrors

import "errors"

var (
	ErrGotInvalidAcc     = errors.New("invalid length of account retrieved from DB")
	ErrIINMismatch       = errors.New("IINs don't match")
	ErrInvalidAmt        = errors.New("invalid amount")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrTokenExpired      = errors.New("Token is expired")
	ErrUpdateRows        = errors.New("failed to update both rows")
)
