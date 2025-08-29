package response

import (
	"encoding/json"
	"net/http"
)

type Success struct {
	Result interface{} `json:"result"`
}

type Error struct {
	Message string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func OK(w http.ResponseWriter, result interface{}) {
	JSON(w, http.StatusOK, Success{Result: result})
}

func Created(w http.ResponseWriter, result interface{}) {
	JSON(w, http.StatusCreated, Success{Result: result})
}

func Fail(w http.ResponseWriter, status int, err error) {
	JSON(w, status, Error{Message: err.Error()})
}
