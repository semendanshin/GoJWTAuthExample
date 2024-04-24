package response

import (
	"encoding/json"
	"go/types"
	"net/http"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type Data interface {
	any
}

type Response[T Data] struct {
	Status string `json:"status"`
	Data   T      `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

func NewSuccessResponse[T Data](data T) *Response[T] {
	return &Response[T]{
		Status: StatusSuccess,
		Data:   data,
	}
}

func NewErrorResponse(err string) *Response[types.Nil] {
	return &Response[types.Nil]{
		Status: StatusError,
		Error:  err,
	}
}

func (r *Response[T]) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func WriteWithError(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := NewErrorResponse(err)
	_ = json.NewEncoder(w).Encode(resp)
}

func WriteWithSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := NewSuccessResponse(data)
	_ = json.NewEncoder(w).Encode(resp)
}
