package middleware

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var td = map[string]string{
	"Error when parsing token": "",
	"Couldn't find IIN":        "jgq1&2w_347192",
	"Couldn't find role":       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDAwMjQ5MTksImlhdCI6MTY0MDAyNDg5OSwiaWluIjoiOTEwODE1NDUwMzUwIiwidXNlcm5hbWUiOiJhc3MiLCJ1c2VydHMiOiIyMDIxLTEyLTE5IDEwOjM2OjM2In0.ouFMo3rLUHELjupcV8yqbiu0-_jWELiZ1pE-r5kht5M",
}

func TestHostClientMultipleAddrs(t *testing.T) {
	ln := fasthttputil.NewInmemoryListener()

	s := &fasthttp.Server{
		Handler: ProcessTokenMiddleware(func(ctx *fasthttp.RequestCtx) {
			ctx.Write(ctx.Host())
			ctx.SetConnectionClose()
		}),
	}
	serverStopCh := make(chan struct{})
	go func() {
		if err := s.Serve(ln); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		close(serverStopCh)
	}()

	dialsCount := make(map[string]int)
	c := &fasthttp.HostClient{
		Addr: "foo,bar,baz",
		Dial: func(addr string) (net.Conn, error) {
			dialsCount[addr]++
			return ln.Dial()
		},
	}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI("http://foobar/baz/aaa?bbb=ddd")
	// Invalid token test
	testNo := 0
	for _, v := range td {
		testNo++
		log.Println("No", testNo)
		req.Header.Add("token", v)
		err := c.Do(req, resp)
		if err != nil {
			t.Fatal("something went wrong", err)
		}
		if resp.StatusCode() != fasthttp.StatusForbidden {
			t.Fatalf("unexpected status code %d. Expecting %d", resp.StatusCode(), fasthttp.StatusForbidden)
		}
		log.Println(testNo, "PASS")
	}
	// Valid token test
	req.Header.Set("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6ZmFsc2UsImV4cCI6MTY0MDAyNDkxOSwiaWF0IjoxNjQwMDI0ODk5LCJpaW4iOiI5MTA4MTU0NTAzNTAiLCJ1c2VybmFtZSI6ImFzcyIsInVzZXJ0cyI6IjIwMjEtMTItMTkgMTA6MzY6MzYifQ.UhJzzMEq7GxY_rBdD5rwsn-CaewsXmKfa9jA0Xt-b0Q")

	for i := 0; i < 9; i++ {
		err := c.Do(req, resp)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			t.Fatalf("unexpected status code %d. Expecting %d", resp.StatusCode(), fasthttp.StatusOK)
		}
		if string(resp.Body()) != "foobar" {
			t.Fatalf("unexpected body %q. Expecting %q", resp.Body(), "foobar")
		}
	}

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	select {
	case <-serverStopCh:
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}

	if len(dialsCount) != 3 {
		t.Fatalf("unexpected dialsCount size %d. Expecting 3", len(dialsCount))
	}
	for _, k := range []string{"foo", "bar", "baz"} {
		if dialsCount[k] != 3 {
			t.Fatalf("unexpected dialsCount for %q. Expecting 3", k)
		}
	}
}
