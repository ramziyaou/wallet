package main

import (
	"log"
	"os"
	"wallet/wallet/delivery"
	"wallet/wallet/repository/mysql"
	"wallet/wallet/usecase"

	"github.com/buaazp/fasthttprouter"
	"github.com/subosito/gotenv"
	"github.com/valyala/fasthttp"
)

func init() {
	gotenv.Load()
}

func main() {
	r := fasthttprouter.New()

	dbConn, err := mysql.NewMySQLDBInterface(os.Getenv("DATA_SOURCE"))
	if err != nil {
		log.Fatalf("Db interface create error: %v", err)
	}

	defer dbConn.Close()
	walletListUsecase := usecase.NewWalletListUsecase(dbConn)
	transferUsecase := usecase.NewTransferUsecase(dbConn)
	topUpUsecase := usecase.NewTopUpUsecase(dbConn)
	addWalletUsecase := usecase.NewAddWalletUsecase(dbConn)
	getWalletsUsecase := usecase.NewGetWalletsUsecase(dbConn)
	getTransactionsUsecase := usecase.NewGetTransactionsUsecase(dbConn)

	delivery.NewGetInfoHandler(r, getWalletsUsecase)
	delivery.NewGetTransactionsHandler(r, getTransactionsUsecase)
	delivery.NewAddWalletHandler(r, addWalletUsecase)
	delivery.NewTopUpHandler(r, topUpUsecase)
	delivery.NewTransferHandler(r, transferUsecase)
	delivery.NewWalletListHandler(r, walletListUsecase)
	fasthttp.ListenAndServe(":8070", r.Handler)
}
