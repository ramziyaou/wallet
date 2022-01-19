package repository

import "wallet/domain"

type DBInterface interface {
	GetLastAccountNo() (string, error)
	InsertWallet(account, IIN string) error
	GetAmount(string) (string, error)
	GetWallets(string) ([]domain.Wallet, error)
	GetWalletList(string) ([]string, error)
	UpdateTransferStatus(from, to string) error
	GetTransactions(account string) ([]domain.Transaction, error)
	ConfirmIIN(IIN, account string) (bool, error)
	Close()
	TopUp(string, int) error
	Transfer(from, to, amt string) error
	SetZeroStatus(from, to string) error
}
