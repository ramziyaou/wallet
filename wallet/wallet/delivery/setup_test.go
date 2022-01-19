package delivery

import (
	"fmt"
	"log"
	"time"
	"wallet/domain"
	"wallet/myerrors"
	"wallet/wallet/repository"
	"wallet/wallet/usecase"

	"github.com/buaazp/fasthttprouter"
	"github.com/golang-jwt/jwt"
	"github.com/valyala/fasthttp"
)

const (
	ACCESS_SECRET = "testingaccess"
)

func getRoutes() fasthttp.RequestHandler {
	r := fasthttprouter.New()

	dbConn, err := NewMySQLDBInterface("")
	if err != nil {
		log.Fatalf("Db interface create error: %v", err)
	}
	//defer dbConn.Close()
	walletListUsecase := usecase.NewWalletListUsecase(dbConn)
	transferUsecase := usecase.NewTransferUsecase(dbConn)
	topUpUsecase := usecase.NewTopUpUsecase(dbConn)
	addWalletUsecase := usecase.NewAddWalletUsecase(dbConn)
	getWalletsUsecase := usecase.NewGetWalletsUsecase(dbConn)
	getTransactionsUsecase := usecase.NewGetTransactionsUsecase(dbConn)

	NewGetInfoHandler(r, getWalletsUsecase)
	NewGetTransactionsHandler(r, getTransactionsUsecase)
	NewAddWalletHandler(r, addWalletUsecase)
	NewTopUpHandler(r, topUpUsecase)
	NewTransferHandler(r, transferUsecase)
	NewWalletListHandler(r, walletListUsecase)
	return r.Handler
}

func GenerateTestToken() (string, error) {
	accessTokenExp := time.Now().Add(20 * time.Second).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin":    false,
		"exp":      accessTokenExp,
		"iat":      1640024899,
		"iin":      "910815450350",
		"username": "sth",
		"userts":   "2021-12-31 19:36:36",
	})

	accessTokenString, err := accessToken.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return "", err
	}
	log.Println(accessTokenString)
	return accessTokenString, nil
}

type testDB struct{}

func (m *testDB) GetLastAccountNo() (string, error) {
	return "KZT0000000000", nil
}

func (m *testDB) InsertWallet(account, IIN string) error {
	return nil
}

func (m *testDB) GetWallets(IIN string) ([]domain.Wallet, error) {
	if IIN == "wrong" {
		return nil, fmt.Errorf("Wrong iin")
	}
	return nil, nil
}

func (m *testDB) TopUp(account string, amt int) error {
	if account == "wrong" {
		return fmt.Errorf("Wrong acc")
	}
	return nil
}

func (m *testDB) GetAmount(account string) (string, error) {
	return "", nil
}

func (m *testDB) SetZeroStatus(from, to string) error {
	if from == "wrongg" {
		return fmt.Errorf("Wrong acc")
	}
	return nil
}

func (m *testDB) ConfirmIIN(IIN, account string) (bool, error) {
	if account == "abc" {
		return false, nil
	}
	return true, nil
}

func (m *testDB) GetTransactions(account string) ([]domain.Transaction, error) {
	if account == "wrong" {
		return nil, fmt.Errorf("Wrong acc")
	}
	return nil, nil
}

func (m *testDB) Transfer(from, to, amt string) error {
	if to == "wrong" {
		return fmt.Errorf("Wrong acc")
	}
	if amt == "insufficient" {
		return myerrors.ErrInsufficientFunds
	}
	return nil
}

func (m *testDB) UpdateTransferStatus(from, to string) error {
	if from == "wrong" {
		return fmt.Errorf("Wrong acc")
	}
	return nil
}

func (m *testDB) GetWalletList(IIN string) ([]string, error) {
	return nil, nil
}

func NewMySQLDBInterface(dbURL string) (repository.DBInterface, error) {
	return &testDB{}, nil
}

func (m *testDB) Close() {}
