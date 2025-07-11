package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/config"
	"github.com/tiredkangaroo/keylock/database"
)

type Server struct {
	db *database.DB
}

func (s *Server) Init(db *database.DB) {
	s.db = db
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp4", config.DefaultConfig.Addr) // NOTE: research, put tcp4 bc i dont think fasthttp supports ipv6
	if err != nil {
		return fmt.Errorf("creating listener at addr %s failed: %w", config.DefaultConfig.Addr, err)
	}
	slog.Info("listening on addr", "addr", listener.Addr().String())

	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})

	api := app.Group("/api")
	api.Post("/accounts/new", s.apiAccountsNew)

	return app.Listener(listener)
}

func (s *Server) apiAccountsNew(c *fiber.Ctx) error {
	rawdata := c.Body()
	if len(rawdata) == 0 {
		return apiErrorStringBadRequest(c, "request body is empty")
	}

	type expected struct {
		Name string `json:"name"`
	}
	var data expected
	err := json.Unmarshal(rawdata, &data)
	if err != nil {
		slog.Error("unmarshalling request body failed", "error", err)
		return apiErrorStringBadRequest(c, "invalid request body")
	}
	if data.Name == "" {
		return apiErrorStringBadRequest(c, "one or more required fields are empty")
	}

	user := &database.User{
		Name: data.Name,
	}
	err = s.db.SaveUser(user)
	if err != nil {
		slog.Error("saving user failed", "error", err)
		return apiErrorStringBadRequest(c, "failed to save user")
	}
	slog.Info("saved user", "name", data.Name, "id", user.ID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user_id": user.ID,
	})
}

func apiErrorStringBadRequest(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": err,
	})
}
