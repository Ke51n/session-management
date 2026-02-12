package constant

import (
	"net/http"
	"session-demo/response"
)

var (
	ErrUnauthorized = &response.BizError{HttpStatus: http.StatusUnauthorized, Code: 401, Msg: "Unauthorized"}
	ErrBadRequest   = &response.BizError{HttpStatus: http.StatusBadRequest, Code: 400, Msg: "Bad Request"}
)
