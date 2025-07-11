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

// NOTE: maybe retrive should be a GET
// NOTE: too much repeated code, NEEDS refactoring

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
	api.Post("/passwords/new", s.apiPasswordsNew)
	api.Post("/passwords/retrieve", s.apiPasswordsRetrieve)

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

func (s *Server) apiPasswordsNew(c *fiber.Ctx) error {
	rawdata := c.Body()
	if len(rawdata) == 0 {
		return apiErrorStringBadRequest(c, "request body is empty")
	}

	type expected struct {
		Code   string `json:"code"`
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
		Value  string `json:"value"`
	}
	var data expected
	err := json.Unmarshal(rawdata, &data)
	if err != nil {
		slog.Error("unmarshalling request body failed", "error", err)
		return apiErrorStringBadRequest(c, "invalid request body")
	}
	if data.Name == "" || data.Value == "" || data.Code == "" {
		return apiErrorStringBadRequest(c, "one or more required fields are empty")
	}

	pwd := &database.Password{
		UserID: data.UserID,
		Name:   data.Name,
		Value:  []byte(data.Value),
	}
	err = s.db.SavePassword(data.Code, pwd)
	if err != nil {
		slog.Error("saving password failed", "error", err)
		return apiErrorStringBadRequest(c, "failed to save password")
	}
	slog.Info("saved password", "name", data.Name, "user_id", data.UserID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"password_id": pwd.ID,
	})
}

func (s *Server) apiPasswordsRetrieve(c *fiber.Ctx) error {
	rawdata := c.Body()
	if len(rawdata) == 0 {
		return apiErrorStringBadRequest(c, "request body is empty")
	}

	type expected struct {
		Code   string `json:"code"`
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
	}
	var data expected
	err := json.Unmarshal(rawdata, &data)
	if err != nil {
		slog.Error("unmarshalling request body failed", "error", err)
		return apiErrorStringBadRequest(c, "invalid request body")
	}
	if data.Name == "" || data.Code == "" {
		return apiErrorStringBadRequest(c, "one or more required fields are empty")
	}

	pwd := &database.Password{
		UserID: data.UserID,
		Name:   data.Name,
	}
	err = s.db.RetrievePassword(data.Code, pwd)
	if err != nil {
		slog.Error("retrieving password failed", "error", err)
		return apiErrorStringBadRequest(c, "failed to retrieve password")
	}
	slog.Info("retrieved password", "name", data.Name, "user_id", data.UserID, "value", string(pwd.Value))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"value": string(pwd.Value),
	})
}

func apiErrorStringBadRequest(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": err,
	})
}
