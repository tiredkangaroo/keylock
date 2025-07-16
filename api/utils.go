package api

import "github.com/gofiber/fiber/v2"

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
