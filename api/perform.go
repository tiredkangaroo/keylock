package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func PerformRequest[X Response, T Request](host string, req T, response X) error {
	u, err := url.Parse(host)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	u.Path = req.GetPath()
	u.RawQuery = req.GetQuery().Encode()

	body, err := req.GetBody()
	if err != nil {
		return fmt.Errorf("body: %w", err)
	}
	httpReq, err := http.NewRequest(req.GetMethod(), u.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	httpReq.Header = make(http.Header)
	for k, v := range req.GetHeader() {
		httpReq.Header.Set(k, v)
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	if err := fillResponse(httpResp, response); err != nil {
		return fmt.Errorf("fill response: %w", err)
	}
	return nil
}

// looks like a disaster of a function but here's what it does:
// provides a function that takes in a request (type T) and returns a response and an error -> we give u the fiber handler for it
func Handler[T Request, X Response](handler func(T) (X, error)) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var req T
		req = req.New().(T)

		if err := req.SetMethod(c.Method()); err != nil {
			return apiErr(c, http.StatusBadRequest, fmt.Errorf("set method: %w", err))
		}
		if err := req.SetPath(c.Path()); err != nil {
			return apiErr(c, http.StatusBadRequest, fmt.Errorf("set path: %w", err))
		}
		if err := req.SetQuery(expand(c.Queries())); err != nil {
			return apiErr(c, http.StatusBadRequest, fmt.Errorf("set query: %w", err))
		}
		if err := req.SetHeader(shrink(c.GetReqHeaders())); err != nil {
			return apiErr(c, http.StatusBadRequest, fmt.Errorf("set header: %w", err))
		}
		if err := req.SetBody(c.Body()); err != nil {
			return apiErr(c, http.StatusBadRequest, fmt.Errorf("set body: %w", err))
		}

		resp, err := handler(req)
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, fmt.Errorf("handler: %w", err))
		}
		return SendResponse(c, resp)
	}
}

func SendResponse[T Response](c *fiber.Ctx, response T) error {
	for key, value := range response.GetHeader() {
		c.Set(key, value)
	}
	for _, cookie := range response.GetCookies() {
		c.Cookie(cookie)
	}
	body, err := response.GetBody()
	if err != nil {
		return fmt.Errorf("get body: %w", err)
	}
	fmt.Println("86", string(body))
	return c.Status(response.GetStatusCode()).Send(body)
}

func fillResponse[X Response](resp *http.Response, response X) error {
	if resp.Header != nil { // avoid nil deref
		h := make(map[string]string, len(resp.Header))
		for key, values := range resp.Header {
			if len(values) == 0 {
				continue // skip empty headers, avoid oob
			}
			h[key] = values[0] // use first value
		}
		if err := response.SetHeaders(h); err != nil {
			return fmt.Errorf("set headers: %w", err)
		}
	}
	if err := response.SetCookies(resp.Cookies()); err != nil {
		return fmt.Errorf("set cookies: %w", err)
	}
	if err := response.SetStatusCode(resp.StatusCode); err != nil {
		return fmt.Errorf("set status code: %w", err)
	}
	if err := response.SetBody(resp.Body); err != nil {
		return fmt.Errorf("set body: %w", err)
	}
	return nil
}
