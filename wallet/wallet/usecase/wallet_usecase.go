package usecase

import (
	"fmt"
	"log"
	"strconv"
	"wallet/domain"
	"wallet/myerrors"
	"wallet/wallet/repository"
)

type AddWalletUsecase interface {
	MakeWallet(IIN string) (string, error)
}

type addWalletUsecaseImpl struct {
	dbConn repository.DBInterface
}

// Creates new wallet and returns account number
func (uc *addWalletUsecaseImpl) MakeWallet(IIN string) (string, error) {
	// Get last account number from DB
	lastAccountNo, err := uc.dbConn.GetLastAccountNo()
	if err != nil {
		log.Printf("Error when making new wallet: %v", err)
		return "", fmt.Errorf("addWalletUsecaseImpl error:%w", err)
	}
	if lastAccountNo == "" {
		lastAccountNo = "KZT00000000000"
	}
	// Generate new account number
	newAccountNo, ok := generateAccountNo(lastAccountNo[3:])
	if !ok {
		return "", fmt.Errorf("Limit exceeded")
	}
	// Insert newly generated wallet
	if err = uc.dbConn.InsertWallet(newAccountNo, IIN); err != nil {
		log.Printf("Error when inserting new wallet: %v", err)
		return "", fmt.Errorf("addWalletUsecaseImpl error:%w", err)
	}
	return newAccountNo, nil
}

// NewAddWalletUsecase returns new addWalletUsecase
func NewAddWalletUsecase(db repository.DBInterface) AddWalletUsecase {
	return &addWalletUsecaseImpl{
		dbConn: db,
	}
}

type GetWalletsUsecase interface {
	GetWallets(IIN string) ([]domain.Wallet, error)
}
type getWalletsUsecaseImpl struct {
	dbConn repository.DBInterface
}

func (uc *getWalletsUsecaseImpl) GetWallets(IIN string) ([]domain.Wallet, error) {
	wallets, err := uc.dbConn.GetWallets(IIN)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func NewGetWalletsUsecase(db repository.DBInterface) GetWalletsUsecase {
	return &getWalletsUsecaseImpl{
		dbConn: db,
	}
}

type TopUpUsecase interface {
	TopUp(IIN, account, initAmt string) (string, error)
	// GetAmount(string) (string, error)
}

type topUpUsecaseImpl struct {
	dbConn repository.DBInterface
}

// TopUp implements account topup logic
func (uc *topUpUsecaseImpl) TopUp(IIN, account, amt string) (string, error) {
	ok, err := uc.dbConn.ConfirmIIN(IIN, account)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", myerrors.ErrIINMismatch
	}
	amtInt, err := strconv.Atoi(amt)
	if err != nil || amtInt < 0 {
		return "", myerrors.ErrInvalidAmt
	}

	if err = uc.dbConn.TopUp(account, amtInt); err != nil {
		return "", err
	}
	amount, err := uc.dbConn.GetAmount(account)
	if err != nil {
		return "", err
	}
	return amount, nil
}

// NewTopUpUsecase returns new TopUpUsecase
func NewTopUpUsecase(db repository.DBInterface) TopUpUsecase {
	return &topUpUsecaseImpl{
		dbConn: db,
	}
}

type WalletListUsecase interface {
	GetWalletList(string) ([]string, error)
}

type walletListUsecaseImpl struct {
	dbConn repository.DBInterface
}

// GetWalletList gets all user wallets as string slice
func (uc *walletListUsecaseImpl) GetWalletList(IIN string) ([]string, error) {
	walletList, err := uc.dbConn.GetWalletList(IIN)
	if err != nil {
		return nil, err
	}
	return walletList, nil
}

// NewWalletListUsecase returns new WalletListUSecase
func NewWalletListUsecase(db repository.DBInterface) WalletListUsecase {
	return &walletListUsecaseImpl{
		dbConn: db,
	}
}
