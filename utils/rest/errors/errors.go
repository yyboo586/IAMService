package errors

import (
	"fmt"
)

type HTTPError struct {
	Code    int
	Msg     string
	Details map[string]interface{}
}

func NewHTTPError(code int, msg string, details map[string]interface{}) *HTTPError {
	return &HTTPError{
		Code:    code,
		Msg:     msg,
		Details: details,
	}
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("错误码: %d, 错误信息: %v, 详细错误信息: %v\n", e.Code, e.Msg, e.Details)
}

func (e *HTTPError) StatusCode() int {
	return e.Code
}
