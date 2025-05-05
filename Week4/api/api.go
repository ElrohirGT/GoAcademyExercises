package api

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	InnerReason error
}

func NewFromError(err error) APIError {
	return APIError{InnerReason: err}
}

func (self APIError) JSONIfyAndRespond(w http.ResponseWriter, code int) {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	http.Error(w, string(bytes), code)
}
