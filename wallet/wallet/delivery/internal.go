package delivery

import (
	"github.com/valyala/fasthttp"
)

func getHeaderValues(ctx *fasthttp.RequestCtx, headers ...string) bool {
	for _, header := range headers {
		value := string(ctx.Request.Header.Peek(header))
		if value == "" {
			return false
		}
		ctx.SetUserValue(header, value)
	}
	return true
}

func getIINAndRole(ctx *fasthttp.RequestCtx) (string, bool, bool) {
	IIN, ok := ctx.UserValue("IIN").(string)
	if !ok {
		return "", false, false
	}
	isAdmin, ok := ctx.UserValue("isAdmin").(bool)
	if !ok {
		return "", false, false
	}
	return IIN, isAdmin, true
}
