package middleware

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"wallet/myerrors"
	"wallet/wallet/delivery/response"

	"github.com/golang-jwt/jwt"
	"github.com/valyala/fasthttp"
)

const TOKEN = "token"

var ErrInvalidToken = errors.New("Invalid token") // TODO? hz kak

// ProcessTokenAndCtxMiddleware extracts IIN from token and populates it to context
func ProcessTokenMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if err := parseToken(ctx); err != nil {
			log.Println("ERROR|ProcessTokenMiddleware", err)
			response.RespondWithError(ctx, fasthttp.StatusForbidden, "Invalid token")
			return
		}
		log.Println("INFO|Token successfuly processed")
		next(ctx)
	}
}

// extractIINAndRole extracts IIN and role from token
func extractIINAndRole(token string) (string, bool, error) {
	fmt.Println("extracting claims from ", token)
	JWTToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Failed to extract token metadata, unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		log.Printf("ERROR|Native parse err:%q", err.Error())
		return "", false, err
	}
	claims, ok := JWTToken.Claims.(jwt.MapClaims)

	var IIN string
	var admin bool

	if ok && JWTToken.Valid {
		log.Println("INFO|Middlewrare: token ok & valid")
		admin, ok = claims["admin"].(bool)
		if !ok {
			return "", false, fmt.Errorf("Field admin not found")
		}
		_, ok = claims["username"].(string)
		if !ok {
			return "", false, fmt.Errorf("Field username not found")
		}
		_, ok = claims["createdAt"].(string)
		if !ok {
			return "", false, fmt.Errorf("Field userts not found")
		}

		IIN, ok = claims["iin"].(string)
		if !ok {
			return "", false, fmt.Errorf("Field iin not found")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return "", false, fmt.Errorf("Field exp not found")
		}

		expiredTime := time.Unix(int64(exp), 0)
		if time.Now().After(expiredTime) {
			return "", false, myerrors.ErrTokenExpired
		}
		log.Println("INFO|Middleware: everything ok, passing IIN and role from parsed token")
		return string(IIN), admin, nil
	}
	log.Println("INFO|Middleware: token not ok or not valid")
	return "", false, myerrors.ErrInvalidToken
}

// parseToken gets token from request header and populates IIN to context
func parseToken(ctx *fasthttp.RequestCtx) error {
	token := string(ctx.Request.Header.Peek("token"))
	if token == "" {
		return fmt.Errorf("Couldn't find token")
	}
	IIN, isAdmin, err := extractIINAndRole(token)
	if err != nil {
		return err
	}
	ctx.SetUserValue("IIN", IIN)
	log.Println("setting isadmin", isAdmin)
	ctx.SetUserValue("isAdmin", isAdmin)
	return nil
}
