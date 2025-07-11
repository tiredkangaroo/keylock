package server

import (
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/database"
)

// i puffified my fiber 😭 oh how i wish puff was working

func APINewAccount(s *Server) fiber.Handler {
	type expectedBody struct {
		Name string `json:"name"`
	}
	return HandlerJSON(func(c *fiber.Ctx, body expectedBody) error {
		user := &database.User{
			Name: body.Name,
		}

		err := s.db.SaveUser(user)
		if err != nil {
			return APIError(c, fiber.StatusBadRequest, err, "failed to save user")
		}

		slog.Info("saved user", "name", body.Name, "id", user.ID)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"user_id": user.ID,
		})
	})
}

func APINewPassword(s *Server) fiber.Handler {
	type expectedBody struct {
		Code   string `json:"code"`
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
		Value  string `json:"value"`
	}
	return HandlerJSON(func(c *fiber.Ctx, body expectedBody) error {
		pwd := &database.Password{
			UserID: body.UserID,
			Name:   body.Name,
			Value:  []byte(body.Value),
		}
		err := s.db.SavePassword(body.Code, pwd)
		if err != nil {
			return APIError(c, fiber.StatusBadRequest, err, "failed to save password")
		}
		slog.Info("saved password", "name", body.Name, "user_id", body.UserID, "code", body.Code)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"password_id": pwd.ID,
		})
	})
}

func APIRetrievePassword(s *Server) fiber.Handler {
	type expectedBody struct {
		Code   string `json:"code"`
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
	}
	return HandlerJSON(func(c *fiber.Ctx, body expectedBody) error {
		pwd := &database.Password{
			UserID: body.UserID,
			Name:   body.Name,
		}
		err := s.db.RetrievePassword(body.Code, pwd)
		if err != nil {
			return APIError(c, http.StatusBadRequest, err, "password not found or incorrect code")
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"value": string(pwd.Value),
		})
	})
}
