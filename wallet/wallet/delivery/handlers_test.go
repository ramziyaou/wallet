package delivery

import (
	"fmt"
	"net"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type headerData struct {
	key   string
	value string
}

var testTable = []struct {
	name               string
	url                string
	method             string
	params             []headerData
	expectedStatusCode int
}{
	{"get-transactions", "/transactions", "GET", []headerData{
		{key: "account", value: "account"},
	}, fasthttp.StatusOK},
	{"get-transfer", "/transfer", "GET", []headerData{
		{key: "from", value: "account"},
		{key: "to", value: "account"},
		{key: "amount", value: "123"},
	}, fasthttp.StatusOK},
	{"get-topup", "/topup", "GET", []headerData{
		{key: "account", value: "123"},
		{key: "amount", value: "123"},
	}, fasthttp.StatusOK},
	{"get-add", "/add", "GET", []headerData{
		{key: "iin", value: "910815450350"},
	}, fasthttp.StatusOK},
	{"get-info", "/info", "GET", []headerData{
		{key: "iin", value: "910815450350"},
	}, fasthttp.StatusOK},
	{"get-walletlist", "/wallets", "GET", []headerData{
		{key: "iin", value: "12345"},
	}, fasthttp.StatusOK},
}

var testTableErr = []struct {
	name               string
	url                string
	method             string
	params             []headerData
	expectedStatusCode int
}{
	{"get-transactions", "/transactions", "GET", []headerData{}, fasthttp.StatusBadRequest},
	{"get-transactions", "/transactions", "GET", []headerData{
		{key: "account", value: "wrong"},
	}, fasthttp.StatusInternalServerError},
	{"get-transactions", "/transactions", "GET", []headerData{
		{key: "account", value: "abc"},
	}, fasthttp.StatusForbidden},
	{"get-transfer", "/transfer", "GET", []headerData{}, fasthttp.StatusBadRequest},
	{"get-transfer", "/transfer", "GET", []headerData{
		{key: "from", value: "wrong"},
		{key: "to", value: "abc"},
		{key: "amount", value: "123"},
	}, fasthttp.StatusInternalServerError},
	{"get-transfer", "/transfer", "GET", []headerData{
		{key: "from", value: "abc"},
		{key: "to", value: "wrong"},
		{key: "amount", value: "123"},
	}, fasthttp.StatusInternalServerError},
	{"get-transfer", "/transfer", "GET", []headerData{
		{key: "from", value: "wrongg"},
		{key: "to", value: "abc"},
		{key: "amount", value: "123"},
	}, fasthttp.StatusInternalServerError},
	{"get-transfer", "/transfer", "GET", []headerData{
		{key: "from", value: "ok"},
		{key: "to", value: "ok"},
		{key: "amount", value: "insufficient"},
	}, fasthttp.StatusBadRequest},
	{"get-topup", "/topup", "GET", []headerData{
		{key: "account", value: "KZT0000000001"},
		{key: "amount", value: "-123"},
	}, fasthttp.StatusBadRequest},
	{"get-topup", "/topup", "GET", []headerData{}, fasthttp.StatusBadRequest},
	{"get-topup", "/topup", "GET", []headerData{
		{key: "account", value: "wrong"},
		{key: "amount", value: "76"},
	}, fasthttp.StatusInternalServerError},
	{"get-info", "/info", "GET", []headerData{
		{key: "iin", value: "wrong"},
	}, fasthttp.StatusForbidden},
	{"get-info", "/info", "GET", []headerData{}, fasthttp.StatusBadRequest},
}

func TestHandlers(t *testing.T) {
	r := getRoutes()

	ln := fasthttputil.NewInmemoryListener()
	defer func() {
		_ = ln.Close()
	}()

	s := &fasthttp.Server{
		Handler: r,
	}

	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.SetRequestURI("http://test.com")
	// valid thru Mon Sep 08 2053 19:21:28 GMT+0600
	access, err := GenerateTestToken()
	if err != nil {
		t.Error("Couldn't generate token", err)
		return
	}
	req.Header.Add("token", access)
	for _, tt := range testTable {
		fmt.Println("Testing", tt.name, "******************************************************************************************")

		req.Header.SetMethod(fasthttp.MethodGet)
		req.SetRequestURI("http://test.com" + tt.url)
		for _, h := range tt.params {
			req.Header.Add(h.key, h.value)
		}

		err := c.Do(req, res)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode() != tt.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, res.StatusCode())
		}
	}

	if res.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("unexpected status code %d. Expecting %d", res.StatusCode(), fasthttp.StatusOK)
	}
}

func TestHandlersError(t *testing.T) {
	r := getRoutes()

	ln := fasthttputil.NewInmemoryListener()
	defer func() {
		_ = ln.Close()
	}()

	s := &fasthttp.Server{
		Handler: r,
	}

	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	req.SetRequestURI("http://test.com")
	for _, tt := range testTableErr {
		fmt.Println("Testing", tt.name, "******************************************************************************************")
		// valid thru Mon Sep 08 2053 19:21:28 GMT+0600
		access, err := GenerateTestToken()
		if err != nil {
			t.Error("Couldn't generate token", err)
			return
		}
		req.Header.Add("token", access)
		req.Header.SetMethod(fasthttp.MethodGet)
		req.SetRequestURI("http://test.com" + tt.url)
		for _, h := range tt.params {
			req.Header.Add(h.key, h.value)
		}

		err = c.Do(req, res)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode() != tt.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", tt.name, tt.expectedStatusCode, res.StatusCode())
		}
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}
}
