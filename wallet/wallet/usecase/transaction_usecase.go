package usecase

import (
	"wallet/domain"
	"wallet/myerrors"
	"wallet/wallet/repository"
)

type GetTransactionsUsecase interface {
	GetTransactions(IIN, account string, isAdmin bool) ([]domain.Transaction, error)
}

type getTransactionsUsecaseImpl struct {
	dbConn repository.DBInterface
}

// GetTransactions gets all transactions
func (uc *getTransactionsUsecaseImpl) GetTransactions(IIN, account string, isAdmin bool) ([]domain.Transaction, error) {
	if !isAdmin {
		ok, err := uc.dbConn.ConfirmIIN(IIN, account)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, myerrors.ErrIINMismatch
		}
	}
	return uc.dbConn.GetTransactions(account)
}

// NewGetTransactionsUsecase returns new TransactionsUsecase
func NewGetTransactionsUsecase(db repository.DBInterface) GetTransactionsUsecase {
	return &getTransactionsUsecaseImpl{
		dbConn: db,
	}
}

type TransferUsecase interface {
	MakeTransfer(from, to, amt, IIN string) error
}

type transferUsecaseImpl struct {
	dbConn repository.DBInterface
}

// MakeTransfer implements transfer logic
func (uc *transferUsecaseImpl) MakeTransfer(from, to, amt, IIN string) error {
	ok, err := uc.dbConn.ConfirmIIN(IIN, from)
	if err != nil {
		return err
	}
	if !ok {
		return myerrors.ErrIINMismatch
	}

	if err := uc.dbConn.UpdateTransferStatus(from, to); err != nil {
		return err
	}

	if err := uc.dbConn.Transfer(from, to, amt); err != nil {
		if err != nil {
			if err := uc.dbConn.SetZeroStatus(from, to); err != nil {
				return err
			}
		}
		return err
	}

	if err := uc.dbConn.SetZeroStatus(from, to); err != nil {
		return err
	}
	return nil
}

// NewTransferUsecase returns new TransferUsecase
func NewTransferUsecase(db repository.DBInterface) TransferUsecase {
	return &transferUsecaseImpl{
		dbConn: db,
	}
}
