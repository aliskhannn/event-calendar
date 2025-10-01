package response

import (
	"encoding/json"
	"net/http"
)

// Success represents the JSON structure for a successful HTTP response.
// It contains a single field, Result, which holds the response data.
type Success struct {
	Result interface{} `json:"result"` // The data to be returned in the response
}

// Error represents the JSON structure for an error HTTP response.
// It contains a single field, Message, which holds the error message.
type Error struct {
	Message string `json:"error"` // The error message describing the failure
}

// JSON writes a JSON response to the provided HTTP response writer.
// It sets the Content-Type header to application/json, writes the specified status code,
// and encodes the provided data as JSON.
//
// Parameters:
//   - w: The HTTP response writer to send the response.
//   - status: The HTTP status code for the response.
//   - data: The data to be encoded as JSON in the response body.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// OK sends a successful HTTP response with a 200 OK status code.
// It wraps the provided result in a Success struct and encodes it as JSON.
//
// Parameters:
//   - w: The HTTP response writer to send the response.
//   - result: The data to be included in the response.
func OK(w http.ResponseWriter, result interface{}) {
	JSON(w, http.StatusOK, Success{Result: result})
}

// Created sends a successful HTTP response with a 201 Created status code.
// It wraps the provided result in a Success struct and encodes it as JSON.
//
// Parameters:
//   - w: The HTTP response writer to send the response.
//   - result: The data to be included in the response.
func Created(w http.ResponseWriter, result interface{}) {
	JSON(w, http.StatusCreated, Success{Result: result})
}

// Fail sends an error HTTP response with the specified status code.
// It wraps the provided error message in an Error struct and encodes it as JSON.
//
// Parameters:
//   - w: The HTTP response writer to send the response.
//   - status: The HTTP status code for the error response (e.g., 400, 404, 500).
//   - err: The error containing the message to be included in the response.
func Fail(w http.ResponseWriter, status int, err error) {
	JSON(w, status, Error{Message: err.Error()})
}
