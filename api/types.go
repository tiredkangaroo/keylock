package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

// really redesigned fiber out here ✌️
// i wish this was auto generated
// i should make a codegen tool + a package for this
// this is similar to grpc but better because i dont like protobuf

type Request interface {
	GetHeader() map[string]string
	GetMethod() string
	GetPath() string
	GetQuery() url.Values
	GetBody() ([]byte, error)

	SetHeader(map[string]string) error
	SetMethod(string) error
	SetPath(string) error
	SetQuery(url.Values) error
	SetBody() error
}

type Response interface {
	GetHeader() map[string]string
	GetCookies() []*fiber.Cookie
	GetStatusCode() int
	GetBody() ([]byte, error)

	SetHeaders(map[string]string) error
	SetCookies([]*http.Cookie) error
	SetStatusCode(int) error
	SetBody(io.Reader) error
}

type BaseRequest struct {
	Header struct{}
	Method string
	Path   string
	Query  struct{}
	Body   struct{}
}

func (b *BaseRequest) GetHeader() map[string]string {
	return map[string]string{}
}
func (b *BaseRequest) GetMethod() string {
	return b.Method
}
func (b *BaseRequest) GetPath() string {
	return b.Path
}
func (b *BaseRequest) GetQuery() url.Values {
	return url.Values{}
}
func (b *BaseRequest) GetBody() ([]byte, error) {
	return json.Marshal(b.Body)
}
func (b *BaseRequest) SetHeader(header map[string]string) error {
	return nil
}
func (b *BaseRequest) SetMethod(method string) error {
	b.Method = method
	return nil
}
func (b *BaseRequest) SetPath(path string) error {
	b.Path = path
	return nil
}
func (b *BaseRequest) SetQuery(query url.Values) error {
	return nil
}
func (b *BaseRequest) SetBody(body []byte) error {
	if err := json.Unmarshal(body, &b.Body); err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}
	return nil
}

type BaseResponse struct {
	Header     struct{}
	Cookies    struct{}
	StatusCode int
	Body       struct{}
}

func (b *BaseResponse) GetHeader() map[string]string {
	return map[string]string{}
}
func (b *BaseResponse) GetCookies() []*fiber.Cookie {
	return nil
}
func (b *BaseResponse) GetStatusCode() int {
	return b.StatusCode
}
func (b *BaseResponse) GetBody() ([]byte, error) {
	return json.Marshal(b.Body)
}
func (b *BaseResponse) SetHeaders(header map[string]string) error {
	return nil
}
func (b *BaseResponse) SetCookies(cookies []*http.Cookie) error {
	return nil
}
func (b *BaseResponse) SetStatusCode(statusCode int) error {
	if statusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}
	b.StatusCode = statusCode
	return nil
}
func (b *BaseResponse) SetBody(body io.Reader) error {
	if err := json.NewDecoder(body).Decode(&b.Body); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	return nil
}

type NewAccountRequest struct {
	BaseRequest
	Body NewAccountRequestBody
}
type NewAccountRequestBody struct {
	Name           string `json:"name"`
	MasterPassword string `json:"master_password"`
}

type NewAccountResponse struct {
	BaseResponse
	Body NewAccountResponseBody
}

type NewAccountResponseBody struct {
	UserID      int64  `json:"user_id"`
	SessionCode string `json:"session_code"`
	Code        string `json:"code"`
}

type NewPasswordRequest struct {
	BaseRequest
	Body NewPasswordRequestBody
}
type NewPasswordRequestBody struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Key2   string `json:"key2"`
	Value  string `json:"value"`
}

type NewPasswordResponse struct {
	BaseResponse
}

type RetrievePasswordRequest struct {
	BaseRequest
	Body RetrievePasswordRequestBody
}
type RetrievePasswordRequestBody struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Key2   string `json:"key2"`
}

type RetrievePasswordResponse struct {
	BaseResponse
	Body struct {
		Value string `json:"value"`
	}
}
