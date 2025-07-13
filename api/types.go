package api

type NewAccountRequest struct {
	Name           string `json:"name"`
	MasterPassword string `json:"master_password"`
}

type NewAccountResponse struct {
	UserID      int64  `json:"user_id"`
	SessionCode string `json:"session_code"`
	Code        string `json:"code"`
}

type NewPasswordRequest struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Key2   string `json:"key2"`
	Value  string `json:"value"`
}

type NewPasswordResponse struct{}

type RetrievePasswordRequest struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Key2   string `json:"key2"`
}

type RetrievePasswordResponse struct {
	Value string `json:"value"`
}

// type RetrievePasswordRequest struct {
// 	UserID int64  `json:"user_id"`
