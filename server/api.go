package server

import (
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func APINewAccount(s *Server) fiber.Handler {
	type expectedBody struct {
		Name           string `json:"name"`
		MasterPassword string `json:"master_password"`
	}
	return HandlerJSON(func(c *fiber.Ctx, body expectedBody) error {
		id, sessionCode, code, err := s.db.SaveUser(body.Name, body.MasterPassword)
		if err != nil {
			return APIError(c, fiber.StatusBadRequest, err, "failed to save user")
		}

		slog.Info("saved user", "name", body.Name, "id", id)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"user_id":      id,
			"session_code": sessionCode,
			"code":         code,
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
	return HandlerJSON(func(c *fiber.Ctx, body expectedBody) error {
		err := s.db.SavePassword(body.UserID, body.Name, body.Key2, body.Value)
		if err != nil {
			return APIError(c, fiber.StatusBadRequest, err, err.Error()) // make sure err does NOT info leak otherwise we're so cooked
		}
		slog.Info("saved password", "name", body.Name, "user_id", body.UserID)

		return c.Status(fiber.StatusOK).SendString("")
	})
}

func APIRetrievePassword(s *Server) fiber.Handler {
	type expectedBody struct {
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
		Key2   string `json:"key2"` // this is session code + code
	}
	return HandlerJSON(func(c *fiber.Ctx, body expectedBody) error {
		val, err := s.db.RetrievePassword(body.UserID, body.Name, body.Key2)
		if err != nil {
			return APIError(c, http.StatusBadRequest, err, "password not found or incorrect code")
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"value": string(val),
		})
	})
}
