package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func expand(i map[string]string) map[string][]string {
	v := make(map[string][]string)
	for key, val := range i {
		v[key] = []string{val}
	}
	return v
}
func shrink(i map[string][]string) map[string]string {
	m := make(map[string]string)
	for key, val := range i {
		m[key] = val[0]
	}
	return m
}
func apiErr(c *fiber.Ctx, code int, err error) error {
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func parseBody[T any](c *fiber.Ctx, dst *T) error {
	if err := c.BodyParser(dst); err != nil {
		return fmt.Errorf("parse request body: %w", err)
	}
	return nil
}

func marshalRequestBody[T any](method, path string, body T) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return &http.Request{
		Method: method,
		URL:    &url.URL{Scheme: scheme, Path: path},
		Body:   io.NopCloser(bytes.NewBuffer(data)),
	}, nil
}

func decodeResponseBody[T any](resp *http.Response, dst *T) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d (body: %s)", resp.StatusCode, body)
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}
