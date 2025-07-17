package main

import (
	"log/slog"

	"github.com/tiredkangaroo/keylock/cache"
	"github.com/tiredkangaroo/keylock/config"
	"github.com/tiredkangaroo/keylock/database"
	"github.com/tiredkangaroo/keylock/server"
)

type Key interface {
	Value() ([]byte, error)
}

func main() {
	config.Init()
	database.Init()
	cache.Init()

	db, err := database.Database()
	if err != nil {
		slog.Error("connecting to database failed (fatal)", "error", err)
		return
	}
	defer db.Close()

	slog.Info("opened database")

	s := &server.Server{}
	s.Init(db)
	if err := s.Start(); err != nil {
		slog.Error("server failed (fatal)", "error", err)
		return
	}
}
