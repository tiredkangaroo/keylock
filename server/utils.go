package server

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// NOTE: make the json parser strict, NO extra fields and NO missing fields
func parseJSONBody[T any](c *fiber.Ctx) (T, error) {
	var data T

	rawdata := c.Body()
	if len(rawdata) == 0 {
		return data, fmt.Errorf("request body is empty")
	}
	if err := json.Unmarshal(rawdata, &data); err != nil {
		return data, fmt.Errorf("invalid request body: %w", err)
	}
	return data, nil
}

// HandlerJSON returns a fiber.Handler that expects a request with JSON request body and unmarshals it into the provided type.
// It will call the provided function with the fiber context and the unmarshaled body.
func HandlerJSON[T any](h func(c *fiber.Ctx, body T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body, err := parseJSONBody[T](c)
		if err != nil {
			return APIError(c, fiber.StatusBadRequest, err, "invalid request body")
		}
		return h(c, body)
	}
}

func APIError(c *fiber.Ctx, status int, real error, display string) error {
	slog.Error(display, "error", real)
	return c.Status(status).JSON(fiber.Map{
		"error": display,
	})
}
