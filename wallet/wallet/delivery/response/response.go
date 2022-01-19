package response

import (
	"encoding/json"
	"wallet/domain"

	"github.com/valyala/fasthttp"
)

func RespondWithError(ctx *fasthttp.RequestCtx, status int, message string) {
	ctx.SetStatusCode(status)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:      false,
			Message: message,
		},
	)
}

func ResponseJSON(ctx *fasthttp.RequestCtx, data string) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:      true,
			Message: data,
		},
	)
}

func ResponseWalletList(ctx *fasthttp.RequestCtx, wl []string) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:         true,
			WalletList: wl,
		},
	)
}

func ResponseWallets(ctx *fasthttp.RequestCtx, wallets []domain.Wallet) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:      true,
			Wallets: wallets,
		},
	)
}

func ResponseTransactions(ctx *fasthttp.RequestCtx, ts []domain.Transaction) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(
		domain.Response{
			OK:           true,
			Transactions: ts,
		},
	)
}
