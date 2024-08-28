package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	ErrMsgInvalidUrl           = "field is not a valid URL. Field:"
	ErrMSgMissingRequiredField = "field is required. Field:"
	ErrMsgUnexpected           = "invalid field or unecpected rule. Field:"
)

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMessages = append(errMessages, fmt.Sprintf("%s %s", ErrMSgMissingRequiredField, err.Field()))
		case "url":
			errMessages = append(errMessages, fmt.Sprintf("%s %s", ErrMsgInvalidUrl, err.Field()))
		default:
			errMessages = append(errMessages, fmt.Sprintf("%s %s", ErrMsgUnexpected, err.Field()))
		}

	}
	return Error(strings.Join(errMessages, ", "))
}
