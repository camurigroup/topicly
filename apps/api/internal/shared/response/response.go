package response

import (
	"encoding/json"
	"net/http"
)

// Meta carries pagination metadata for list responses.
type Meta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// ErrorBody is the error portion of the response envelope.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Envelope is the consistent response shape for the whole API.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *ErrorBody  `json:"error"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// JSON writes an envelope with the given status code.
func JSON(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// OK writes a 200 success response.
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Envelope{Success: true, Data: data})
}

// Created writes a 201 success response.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Envelope{Success: true, Data: data})
}

// Paginated writes a 200 success response with pagination metadata.
func Paginated(w http.ResponseWriter, data interface{}, meta Meta) {
	JSON(w, http.StatusOK, Envelope{Success: true, Data: data, Meta: &meta})
}

// Error writes a failure response with a machine-readable code.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, Envelope{Success: false, Error: &ErrorBody{Code: code, Message: message}})
}
