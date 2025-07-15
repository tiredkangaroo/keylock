package server

import (
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/api"
)

func APINewAccount(s *Server) fiber.Handler {
	return HandlerJSON(func(c *fiber.Ctx, body api.NewAccountRequest) error {
		id, sessionCode, code, err := s.db.SaveUser(body.Name, body.MasterPassword)
		if err != nil {
			return APIError(c, fiber.StatusBadRequest, err, "failed to save user")
		}

		slog.Info("saved user", "name", body.Name, "id", id)
		return c.Status(fiber.StatusOK).JSON(api.NewAccountResponse{
			UserID:      id,
			SessionCode: sessionCode,
			Code:        code,
		})
	})
}

func APINewPassword(s *Server) fiber.Handler {
	type expectedBody struct {
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
		Key2   string `json:"key2"`
		Value  string `json:"value"`
	}
	return HandlerJSON(func(c *fiber.Ctx, body api.NewPasswordRequest) error {
		err := s.db.SavePassword(body.UserID, body.Name, body.Key2, body.Value)
		if err != nil {
			return APIError(c, fiber.StatusBadRequest, err, err.Error()) // make sure err does NOT info leak otherwise we're so cooked
		}
		slog.Info("saved password", "name", body.Name, "user_id", body.UserID)

		return c.Status(fiber.StatusOK).JSON(api.NewPasswordResponse{})
	})
}

func APIRetrievePassword(s *Server) fiber.Handler {
	return HandlerJSON(func(c *fiber.Ctx, body api.RetrievePasswordRequest) error {
		val, err := s.db.RetrievePassword(body.UserID, body.Name, body.Key2)
		if err != nil {
			return APIError(c, http.StatusBadRequest, err, "password not found or incorrect code")
		}
		return c.Status(fiber.StatusOK).JSON(api.RetrievePasswordResponse{
			Value: string(val),
		})
	})
}

func APIListPasswords(s *Server) fiber.Handler {
	return HandlerJSON(func(c *fiber.Ctx, data api.ListPasswordsRequest) error {
		passwords, err := s.db.ListPasswords(data.UserID)
		if err != nil {
			return APIError(c, http.StatusBadRequest, err, "failed to list passwords")
		}
		return c.Status(fiber.StatusOK).JSON(api.ListPasswordsResponse{
			Passwords: passwords,
		})
	})
}
