package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/goccy/go-reflect"
	"github.com/gofiber/fiber/v2"
)

// NOTE: make the json parser strict, NO extra fields and NO missing fields
func parseJSONBody[T any](c *fiber.Ctx) (T, error) {
	var data T

	// possibly use bytebufferpool for mem
	decoder := json.NewDecoder(bytes.NewBuffer(c.Body()))
	decoder.DisallowUnknownFields()
	defer c.Request().CloseBodyStream()
	if err := decoder.Decode(&data); err != nil {
		return data, fmt.Errorf("json error: %w", err)
	}

	v := reflect.ValueOf(data)
	t := v.Type()
	if t.Kind() != reflect.Struct { // validation not performed here
		return data, nil
	}

	// basic validation for required fields (or fields not tagged with `required:"false"`)
	// this can't catch boolean or number fields that aren't provided, since their zero values are considered valid
	for i := range t.NumField() {
		structField := t.Field(i)
		fieldValue := v.Field(i)
		if structField.Tag.Get("required") == "false" {
			continue
		}

		switch structField.Type.Kind() {
		case reflect.String:
			if fieldValue.String() == "" {
				return data, fmt.Errorf("field %s not provided (empty)", structField.Name)
			}
		case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
			if fieldValue.IsNil() {
				return data, fmt.Errorf("field %s not provided (null/nil)", structField.Name)
			}
		}
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
