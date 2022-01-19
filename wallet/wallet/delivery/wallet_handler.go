package delivery

import (
	"fmt"
	"log"
	"wallet/myerrors"
	"wallet/wallet/delivery/middleware"
	"wallet/wallet/delivery/response"
	"wallet/wallet/usecase"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type AddWalletHandler struct {
	uc usecase.AddWalletUsecase
}

// AddWallet hadnles creation of new wallet
func (h *AddWalletHandler) AddWallet(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|AddWallet endpoint hit")
	IIN, _, ok := getIINAndRole(ctx)
	if !ok {
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "")
		return
	}
	log.Println("INFO|getIIN successful")
	account, err := h.uc.MakeWallet(IIN)
	if err != nil {
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO|Created new wallet for user %s under account%s\n", IIN, account)
	response.ResponseJSON(ctx, account)
}

// NewAddWalletHandler sets /add route
func NewAddWalletHandler(r *fasthttprouter.Router, uc usecase.AddWalletUsecase) {
	handler := &AddWalletHandler{
		uc: uc,
	}
	r.GET("/add", middleware.ProcessTokenMiddleware(handler.AddWallet))
}

type GetInfoHandler struct {
	uc usecase.GetWalletsUsecase
}

// GetInfo hadnles retrieval of user account info
func (h *GetInfoHandler) GetInfo(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|GetInfo hit")
	tokenIIN, isAdmin, ok := getIINAndRole(ctx)
	if !ok {
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "Failed to get IIN and/or role")
		return
	}
	if isAdmin {
		IIN := string(ctx.Request.Header.Peek("iin"))
		if IIN == "" {
			response.RespondWithError(ctx, fasthttp.StatusBadRequest, "Failed to get IIN")
			return
		}
		log.Println("received iin", IIN)
		tokenIIN = IIN
	}

	log.Println("INFO|retrieved IIN successfully, sending IIN", tokenIIN)
	wallets, err := h.uc.GetWallets(tokenIIN)
	if err != nil {
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	response.ResponseWallets(ctx, wallets)
}

func NewGetInfoHandler(r *fasthttprouter.Router, uc usecase.GetWalletsUsecase) {
	handler := &GetInfoHandler{
		uc: uc,
	}
	r.GET("/info", middleware.ProcessTokenMiddleware(handler.GetInfo))
}

type WalletListHandler struct {
	uc usecase.WalletListUsecase
}

// GetWalletList handles retrieval of user account list
func (h *WalletListHandler) GetWalletList(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|GetWalletList endpoint hit")
	IIN, _, ok := getIINAndRole(ctx)
	if !ok {
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "Failed to get IIN and/or role")
		return
	}
	log.Println("INFO|getIINAndRole successful")

	walletList, err := h.uc.GetWalletList(IIN)
	if err != nil {
		fmt.Println("ERROR|GetWalletList handler:", err)
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if len(walletList) == 0 {
		log.Println("INFO|GetWalletList handler: user has no accounts")
		response.ResponseJSON(ctx, "empty")
		return
	}
	response.ResponseWalletList(ctx, walletList)
}

// NewWalletListHandler sets /wallets route
func NewWalletListHandler(r *fasthttprouter.Router, uc usecase.WalletListUsecase) {
	handler := &WalletListHandler{
		uc: uc,
	}
	r.GET("/wallets", middleware.ProcessTokenMiddleware(handler.GetWalletList))
}

type TopUpHandler struct {
	uc usecase.TopUpUsecase
}

// TopUp hadnles account replenishment
func (h *TopUpHandler) TopUp(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|Topup endpoint hit")
	IIN, _, ok := getIINAndRole(ctx)
	if !ok {
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "")
		return
	}
	values, ok := getTopupValues(ctx)
	if !ok {
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "invalid form data")
		return
	}
	accountNo, amount := values[0], values[1]
	fmt.Println(accountNo, amount, "acc#, amt")
	res, err := h.uc.TopUp(IIN, accountNo, amount)
	if err != nil {
		log.Println("ERROR|Topup handler:", err)
		if err == myerrors.ErrInvalidAmt {
			response.RespondWithError(ctx, fasthttp.StatusBadRequest, "invalid amount")
			return
		}
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO|Account %s amount updated to%s\n", accountNo, res)
	response.ResponseJSON(ctx, res)
}

func NewTopUpHandler(r *fasthttprouter.Router, uc usecase.TopUpUsecase) {
	handler := &TopUpHandler{
		uc: uc,
	}
	r.GET("/topup", middleware.ProcessTokenMiddleware(handler.TopUp))
}

// getTopupValues retrieves account number and topup amount from headers
func getTopupValues(ctx *fasthttp.RequestCtx) ([]string, bool) {
	var values []string
	for _, header := range []string{"account", "amount"} {
		value := string(ctx.Request.Header.Peek(header))
		if value == "" {
			return nil, false
		}
		fmt.Println(value)
		values = append(values, value)
	}
	return values, true
}
