package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/config"
	"github.com/tiredkangaroo/keylock/database"
	"github.com/tiredkangaroo/keylock/server/middlewares"
	"github.com/tiredkangaroo/keylock/web"
)

// NOTE: maybe retrive should be a GET

type Server struct {
	db *database.DB
}

func (s *Server) Init(db *database.DB) {
	s.db = db
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp4", config.DefaultConfig.Addr) // fiber is yet to support ipv6
	if err != nil {
		return fmt.Errorf("creating listener at addr %s failed: %w", config.DefaultConfig.Addr, err)
	}
	slog.Info("listening on addr", "addr", listener.Addr().String())

	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})

	sessionMiddleware := middlewares.SessionMiddleware(s.db)

	webGroup := app.Group("")
	web.SetGroup(s.db, sessionMiddleware, webGroup)

	api := app.Group("/api")
	api.Post("/accounts/new", APINewAccount(s))
	api.Post("/passwords/new", sessionMiddleware, APINewPassword(s))
	api.Post("/passwords/retrieve", sessionMiddleware, APIRetrievePassword(s))
	api.Get("/passwords/list", sessionMiddleware, APIListPasswords(s))

	return app.Listener(listener)
}
