package constant

import (
	"net/http"
	"session-demo/response"
)

var (
	ErrUnauthorized   = &response.BizError{HttpStatus: http.StatusUnauthorized, Code: 401, Msg: "Unauthorized"}
	ErrBadRequest     = &response.BizError{HttpStatus: http.StatusBadRequest, Code: 400, Msg: "Bad Request"}
	ErrStreamNotFound = &response.BizError{HttpStatus: http.StatusNotFound, Code: 404, Msg: "Stream Not Found"}
	ErrInternalServer = &response.BizError{HttpStatus: http.StatusInternalServerError, Code: 500, Msg: "Internal Server Error"}
)
