package api

import (
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
	SetBody([]byte) error

	New() Request
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

	New() Response
}

type BaseRequest struct {
	User *database.User // the user making the request (nil if not authenticated)

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
	fmt.Println("72", b.Body)
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
	b.StatusCode = statusCode
	return nil
}
func (b *BaseResponse) SetBody(body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("consume response body: %w", err)
	}
	fmt.Println(string(data))
	if err := json.Unmarshal(data, &b.Body); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

type NewAccountRequest struct {
	BaseRequest
	Body NewAccountRequestBody
}

func (r *NewAccountRequest) New() Request {
	return &NewAccountRequest{}
}
func (r *NewAccountRequest) GetMethod() string {
	return http.MethodPost
}
func (r *NewAccountRequest) GetPath() string {
	return "/api/accounts/new"
}

type NewAccountRequestBody struct {
	Name           string `json:"name"`
	MasterPassword string `json:"master_password"`
}

type NewAccountResponse struct {
	BaseResponse
	Body NewAccountResponseBody
}

func (r *NewAccountResponse) New() Response {
	return &NewAccountResponse{}
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

func (r *NewPasswordRequest) New() Request {
	return &NewPasswordRequest{}
}
func (r *NewPasswordRequest) GetMethod() string {
	return http.MethodPost
}
func (r *NewPasswordRequest) GetPath() string {
	return "/api/passwords/new"
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

func (r *NewPasswordResponse) New() Response {
	return &NewPasswordResponse{}
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

func (r *RetrievePasswordRequest) New() Request {
	return &RetrievePasswordRequest{}
}
func (r *RetrievePasswordRequest) GetMethod() string {
	return http.MethodPost
}
func (r *RetrievePasswordRequest) GetPath() string {
	return "/api/passwords/retrieve"
}

type RetrievePasswordResponse struct {
	BaseResponse
	Body RetrievePasswordResponseBody
}

func (r *RetrievePasswordResponse) New() Response {
	return &RetrievePasswordResponse{}
}

type RetrievePasswordResponseBody struct {
	Value string `json:"value"`
}

type ListPasswordsRequest struct {
	BaseRequest
	Header struct {
		Authorization string
	}
}

func (r *ListPasswordsRequest) New() Request {
	return &ListPasswordsRequest{}
}
func (r *ListPasswordsRequest) GetMethod() string {
	return http.MethodGet
}
func (r *ListPasswordsRequest) GetPath() string {
	return "/api/passwords/list"
}

func (r *ListPasswordsRequest) GetHeader() map[string]string {
	return map[string]string{
		"Authorization": r.Header.Authorization,
	}
}
func (r *ListPasswordsRequest) SetHeader(header map[string]string) error {
	if auth, ok := header["Authorization"]; ok {
		r.Header.Authorization = auth
	} else {
		return fmt.Errorf("missing Authorization header")
	}
	return nil
}

type ListPasswordsResponse struct {
	BaseResponse
	Body ListPasswordsResponseBody
}

func (r *ListPasswordsResponse) New() Response {
	return &ListPasswordsResponse{}
}

type ListPasswordsResponseBody struct {
	Passwords []database.Password `json:"passwords"`
}
