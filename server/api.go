package server

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/api"
)

func APINewAccount(s *Server) fiber.Handler {
	return api.Handler(func(c *fiber.Ctx, req *api.NewAccountRequest) (*api.NewAccountResponse, error) {
		id, sessionCode, code, err := s.db.SaveUser(req.Body.Name, req.Body.MasterPassword)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				slog.Warn("user already exists", "name", req.Body.Name)
				return nil, fmt.Errorf("user already exists with name \"%s\"", req.Body.Name)
			}
			return nil, err
		}

		sessionID, err := newSessionForUser(id)
		if err != nil {
			return nil, fmt.Errorf("session: %w", err)
		}

		slog.Info("saved user", "name", req.Body.Name, "id", id)
		return &api.NewAccountResponse{
			Cookies: api.NewAccountResponseCookies{
				Session: sessionID,
			},
			Body: api.NewAccountResponseBody{
				UserID:      id,
				SessionCode: sessionCode,
				Code:        code,
			},
		}, nil
	})
}

func APINewPassword(s *Server) fiber.Handler {
	return api.Handler(func(c *fiber.Ctx, req *api.NewPasswordRequest) (*api.NewPasswordResponse, error) {
		user := getUser(c)

		err := s.db.SavePassword(user.ID, req.Body.Name, req.Body.Key2, req.Body.Value)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		slog.Info("saved password", "name", req.Body.Name, "user_id", user.ID)
		return &api.NewPasswordResponse{}, nil
	})
}

func APIRetrievePassword(s *Server) fiber.Handler {
	return api.Handler(func(c *fiber.Ctx, req *api.RetrievePasswordRequest) (*api.RetrievePasswordResponse, error) {
		val, err := s.db.RetrievePassword(req.Body.UserID, req.Body.Name, req.Body.Key2)
		if err != nil {
			return nil, fmt.Errorf("password not found or incorrect code: %w", err)
		}
		return &api.RetrievePasswordResponse{
			Body: api.RetrievePasswordResponseBody{
				Value: string(val),
			},
		}, nil
	})
}

func APIListPasswords(s *Server) fiber.Handler {
	return api.Handler(func(c *fiber.Ctx, req *api.ListPasswordsRequest) (*api.ListPasswordsResponse, error) {
		user := getUser(c)
		passwords, err := s.db.ListPasswords(user.ID)
		if err != nil {
			return nil, fmt.Errorf("list passwords: %w", err)
		}
		slog.Info("listing passwords", "user_id", user.ID, "count", len(passwords))
		return &api.ListPasswordsResponse{
			Body: api.ListPasswordsResponseBody{
				Passwords: passwords,
			},
		}, nil
	})
}
