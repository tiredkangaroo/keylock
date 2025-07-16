package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/database"
)

// really redesigned fiber out here ✌️
// i wish this was auto generated
// i should make a codegen tool + a package for this
// this is similar to grpc but better because i dont like protobuf

const scheme = "http"

type Request interface {
	FromCtx(*fiber.Ctx) (Request, error)
	HTTPRequest() (*http.Request, error)
}

type Response interface {
	FromResp(*http.Response) (Response, error)
	Send(*fiber.Ctx) error
}

// new account request (/api/accounts/new)
type NewAccountRequest struct {
	Body NewAccountRequestBody
}
type NewAccountRequestBody struct {
	Name           string `json:"name"`
	MasterPassword string `json:"master_password"`
}

func (r *NewAccountRequest) FromCtx(c *fiber.Ctx) (Request, error) {
	r = &NewAccountRequest{}
	if err := c.BodyParser(&r.Body); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}
	if r.Body.Name == "" || r.Body.MasterPassword == "" {
		return nil, fmt.Errorf("name and master_password are required")
	}
	return r, nil
}

func (r *NewAccountRequest) HTTPRequest() (*http.Request, error) {
	body, err := json.Marshal(r.Body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return &http.Request{
		Method: http.MethodPost,
		URL:    &url.URL{Scheme: scheme, Path: "/api/accounts/new"},
		Body:   io.NopCloser(bytes.NewBuffer(body)),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}, nil
}

// new account response (/api/accounts/new)

type NewAccountResponse struct {
	Body NewAccountResponseBody
}
type NewAccountResponseBody struct {
	UserID      int64  `json:"user_id"`
	SessionCode string `json:"session_code"`
	Code        string `json:"code"`
}

func (r *NewAccountResponse) FromResp(resp *http.Response) (Response, error) {
	r = new(NewAccountResponse)
	err := decodeResponseBody(resp, &r.Body)
	return r, err
}
func (r *NewAccountResponse) Send(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8) // wow look how specific my mime type is
	c.Status(http.StatusOK)
	return c.JSON(r.Body)
}

type NewPasswordRequest struct {
	Body NewPasswordRequestBody
}

type NewPasswordRequestBody struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Key2   string `json:"key2"`
	Value  string `json:"value"`
}

func (r *NewPasswordRequest) FromCtx(c *fiber.Ctx) (Request, error) {
	r = &NewPasswordRequest{}
	if err := c.BodyParser(&r.Body); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}
	// add validation if needed
	return r, nil
}

func (r *NewPasswordRequest) HTTPRequest() (*http.Request, error) {
	body, err := json.Marshal(r.Body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return &http.Request{
		Method: http.MethodPost,
		URL:    &url.URL{Scheme: scheme, Path: "/api/passwords/new"},
		Body:   io.NopCloser(bytes.NewBuffer(body)),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}, nil
}

type NewPasswordResponse struct{}

func (r *NewPasswordResponse) FromResp(resp *http.Response) (Response, error) {
	// no response body expected
	return r, nil
}

func (r *NewPasswordResponse) Send(c *fiber.Ctx) error {
	c.Status(http.StatusOK)
	return nil
}

// --- RetrievePassword ---

type RetrievePasswordRequest struct {
	Body RetrievePasswordRequestBody
}

type RetrievePasswordRequestBody struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Key2   string `json:"key2"`
}

func (r *RetrievePasswordRequest) FromCtx(c *fiber.Ctx) (Request, error) {
	r = &RetrievePasswordRequest{}
	if err := c.BodyParser(&r.Body); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}
	// validation if needed
	return r, nil
}

func (r *RetrievePasswordRequest) HTTPRequest() (*http.Request, error) {
	body, err := json.Marshal(r.Body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return &http.Request{
		Method: http.MethodPost,
		URL:    &url.URL{Scheme: scheme, Path: "/api/passwords/retrieve"},
		Body:   io.NopCloser(bytes.NewBuffer(body)),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}, nil
}

type RetrievePasswordResponse struct {
	Body RetrievePasswordResponseBody
}

type RetrievePasswordResponseBody struct {
	Value string `json:"value"`
}

func (r *RetrievePasswordResponse) FromResp(resp *http.Response) (Response, error) {
	r = &RetrievePasswordResponse{}
	err := decodeResponseBody(resp, &r.Body)
	return r, err
}

func (r *RetrievePasswordResponse) Send(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(http.StatusOK)
	return c.JSON(r.Body)
}

// --- ListPasswords ---

type ListPasswordsRequest struct {
	Header struct {
		Authorization string
	}
}

func (r *ListPasswordsRequest) FromCtx(c *fiber.Ctx) (Request, error) {
	r = &ListPasswordsRequest{}
	r.Header.Authorization = c.Get("Authorization")
	if r.Header.Authorization == "" {
		return nil, fmt.Errorf("missing Authorization header")
	}
	return r, nil
}

func (r *ListPasswordsRequest) HTTPRequest() (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, "/api/passwords/list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", r.Header.Authorization)
	return req, nil
}

type ListPasswordsResponse struct {
	Body ListPasswordsResponseBody
}

type ListPasswordsResponseBody struct {
	Passwords []database.Password `json:"passwords"`
}

func (r *ListPasswordsResponse) FromResp(resp *http.Response) (Response, error) {
	r = &ListPasswordsResponse{}
	err := decodeResponseBody(resp, &r.Body)
	return r, err
}

func (r *ListPasswordsResponse) Send(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(http.StatusOK)
	return c.JSON(r.Body)
}
