package httputil

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// JSONStatusError is a Huma StatusError whose JSON body is exactly
// {"message":"..."}. It also forces application/json instead of
// application/problem+json.
type JSONStatusError struct {
	status  int
	Message string `json:"message"`
}

func (e *JSONStatusError) Error() string {
	return e.Message
}

func (e *JSONStatusError) GetStatus() int {
	return e.status
}

func (e *JSONStatusError) ContentType(string) string {
	return "application/json"
}

func init() {
	huma.NewError = func(status int, message string, _ ...error) huma.StatusError {
		if status == http.StatusUnprocessableEntity {
			status = http.StatusBadRequest
			if message == "" || message == http.StatusText(http.StatusUnprocessableEntity) {
				message = http.StatusText(http.StatusBadRequest)
			}
		}
		if message == "" {
			message = http.StatusText(status)
		}
		return &JSONStatusError{status: status, Message: message}
	}
}

// WriteStatusError writes err using the standard {"message":"..."} JSON shape.
// Huma status errors keep their status and message; anything else becomes a 500.
func WriteStatusError(w http.ResponseWriter, err error) {
	if se, ok := errors.AsType[huma.StatusError](err); ok {
		WriteError(w, se.GetStatus(), se.Error())
		return
	}
	WriteErrorAuto(w, err)
}
