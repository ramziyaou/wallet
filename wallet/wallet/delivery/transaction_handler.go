package delivery

import (
	"log"
	"wallet/myerrors"
	"wallet/wallet/delivery/middleware"
	"wallet/wallet/delivery/response"
	"wallet/wallet/usecase"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type TransferHandler struct {
	uc usecase.TransferUsecase
}

// TransferHandler handles account transactions
func (h *TransferHandler) Transfer(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|Transfer endpoint hit")
	values, ok := getTransferValues(ctx)
	if !ok {
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "Invalid form data")
		return
	}
	from, to, amount := values[0], values[1], values[2]

	IIN, ok := ctx.UserValue(("IIN")).(string)
	if !ok {
		log.Println("ERROR|Failed to get IIN from ctx")
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "")
		return
	}

	if err := h.uc.MakeTransfer(from, to, amount, IIN); err != nil {
		if err == myerrors.ErrInsufficientFunds {
			log.Println("ERROR|Insufficient funds to deduct from account")
			response.RespondWithError(ctx, fasthttp.StatusBadRequest, "insufficient funds")
			return
		}
		log.Println("ERROR|", err)
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	log.Println("INFO|Successfully set transfer status back to 0")
	response.ResponseJSON(ctx, "success!")
}

// NewTransferHandler sets /transfer route
func NewTransferHandler(r *fasthttprouter.Router, uc usecase.TransferUsecase) {
	handler := &TransferHandler{
		uc: uc,
	}
	r.GET("/transfer", middleware.ProcessTokenMiddleware(handler.Transfer))
}

// getTransferValues retrieves account numbers and trasnfer amount from request headers
func getTransferValues(ctx *fasthttp.RequestCtx) ([]string, bool) {
	var values []string
	for _, header := range []string{"from", "to", "amount"} {
		value := string(ctx.Request.Header.Peek(header))
		if value == "" {
			return nil, false
		}
		values = append(values, value)
	}
	return values, true
}

type GetTransactionsHandler struct {
	uc usecase.GetTransactionsUsecase
}

// GetTransacions handles retieval of account transaction history
func (h *GetTransactionsHandler) GetTransactions(ctx *fasthttp.RequestCtx) {
	log.Println("INFO|GetTransactions hit")

	account := string(ctx.Request.Header.Peek("account"))
	log.Println(account)
	if account == "" {
		log.Println("ERROR|Couldn't find accountno")
		response.RespondWithError(ctx, fasthttp.StatusBadRequest, "accountno not provided")
		return
	}

	IIN, ok := ctx.UserValue(("IIN")).(string)
	if !ok {
		log.Println("ERROR|Failed to get IIN from ctx")
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "")
		return
	}
	isAdmin, ok := ctx.UserValue("isAdmin").(bool)
	if !ok {
		log.Println("ERROR|Failed to get role from ctx")
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "")
		return
	}

	transactions, err := h.uc.GetTransactions(IIN, account, isAdmin)
	if err != nil {
		log.Println("ERROR|Getting transactions:", err)
		if err == myerrors.ErrIINMismatch {
			response.RespondWithError(ctx, fasthttp.StatusForbidden, err.Error())
			return
		}
		response.RespondWithError(ctx, fasthttp.StatusInternalServerError, "")
		return
	}
	response.ResponseTransactions(ctx, transactions)
}

// NewGetTransactionsHandler sets /transactions route
func NewGetTransactionsHandler(r *fasthttprouter.Router, uc usecase.GetTransactionsUsecase) {
	handler := &GetTransactionsHandler{
		uc: uc,
	}
	r.GET("/transactions", middleware.ProcessTokenMiddleware(handler.GetTransactions))
}
