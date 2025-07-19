package api

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func PerformRequest[X Response, T Request](host string, req T) (X, error) {
	var zerov X
	httpReq, err := req.HTTPRequest()
	if err != nil {
		return zerov, fmt.Errorf("make http request: %w", err)
	}
	httpReq.URL.Host = host
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return zerov, fmt.Errorf("perform http request: %w", err)
	}
	defer httpResp.Body.Close()
	resp, err := zerov.FromResp(httpResp)
	if err != nil {
		return zerov, fmt.Errorf("make response: %w", err)
	}
	return resp.(X), nil
}

// looks like a disaster of a function but here's what it does:
// provides a function that takes in a request (type T) and returns a response and an error -> we give u the fiber handler for it
func Handler[T Request, X Response](handler func(*fiber.Ctx, T) (X, error)) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var zerov T
		req, err := zerov.FromCtx(c)
		if err != nil {
			return apiErr(c, http.StatusBadRequest, fmt.Errorf("parse request: %w", err))
		}
		resp, err := handler(c, req.(T))
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, err)
		}
		return resp.Send(c)
	}
}
