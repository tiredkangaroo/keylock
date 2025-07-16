package server

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/api"
)

func APINewAccount(s *Server) fiber.Handler {
	return api.Handler(func(req *api.NewAccountRequest) (*api.NewAccountResponse, error) {
		id, sessionCode, code, err := s.db.SaveUser(req.Body.Name, req.Body.MasterPassword)
		if err != nil {
			return nil, err
		}

		slog.Info("saved user", "name", req.Body.Name, "id", id)
		return &api.NewAccountResponse{
			Body: api.NewAccountResponseBody{
				UserID:      id,
				SessionCode: sessionCode,
				Code:        code,
			},
		}, nil
	})
}

func APINewPassword(s *Server) fiber.Handler {
	return api.Handler(func(req *api.NewPasswordRequest) (*api.NewPasswordResponse, error) {
		err := s.db.SavePassword(req.Body.UserID, req.Body.Name, req.Body.Key2, req.Body.Value)
		if err != nil {
			return nil, err
		}
		slog.Info("saved password", "name", req.Body.Name, "user_id", req.Body.UserID)
		return &api.NewPasswordResponse{}, nil
	})
}

func APIRetrievePassword(s *Server) fiber.Handler {
	return api.Handler(func(req *api.RetrievePasswordRequest) (*api.RetrievePasswordResponse, error) {
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
	return api.Handler(func(req *api.ListPasswordsRequest) (*api.ListPasswordsResponse, error) {
		// passwords, err := s.db.ListPasswords(req.User.ID)
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to list passwords: %w", err)
		// }
		return &api.ListPasswordsResponse{
			Body: api.ListPasswordsResponseBody{},
		}, nil
	})
}
