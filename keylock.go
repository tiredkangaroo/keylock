package main

import (
	"log/slog"
	"net"

	"github.com/tiredkangaroo/keylock/config"
	"github.com/valyala/fasthttp"
)

type Key interface {
	Value() ([]byte, error)
}

func main() {
	listener, err := net.Listen("tcp4", config.DefaultConfig.Addr) // NOTE: research, put tcp4 bc i dont think fasthttp supports ipv6
	if err != nil {
		slog.Error("creating listener at addr failed (fatal)", "addr", config.DefaultConfig.Addr, "error", err)
		return
	}
	slog.Info("listening on addr", "addr", listener.Addr().String())

	err = fasthttp.Serve(listener, func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("hello!")
	})
	if err != nil {
		slog.Error("serving requests failed (fatal)", "addr", listener.Addr().String(), "error", err)
		return
	}
}
