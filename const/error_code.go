package constant

import (
	"net/http"
	"session-demo/response"
)

var (
	// 内部服务器错误
	ErrInternalServer = &response.BizError{HttpStatus: http.StatusInternalServerError, Code: 500, Msg: "Internal Server Error"}
	// 认证错误
	ErrUnauthorized = &response.BizError{HttpStatus: http.StatusUnauthorized, Code: 401, Msg: "Unauthorized"}

	// 请求错误
	ErrBadRequest = &response.BizError{HttpStatus: http.StatusBadRequest, Code: 400, Msg: "Bad Request"}
	// 流不存在错误
	ErrStreamNotFound = &response.BizError{HttpStatus: http.StatusNotFound, Code: 404, Msg: "Stream Not Found"}

	// 项目ID不匹配错误
	ErrProjectIDNotMatch = &response.BizError{HttpStatus: http.StatusBadRequest, Code: 400, Msg: "Project ID Not Match"}
	ErrProjectNotFound   = &response.BizError{HttpStatus: http.StatusNotFound, Code: 404, Msg: "Project Not Found"}

	// 查询消息错误
	ErrQueryMessageError = &response.BizError{HttpStatus: http.StatusNotFound, Code: 404, Msg: "Query Message Error"}
	// 创建消息错误
	ErrCreateMessageError = &response.BizError{HttpStatus: http.StatusInternalServerError, Code: 500, Msg: "Create Message Error"}

	// 无效的消息ID错误
	ErrInvalidMessageID = &response.BizError{HttpStatus: http.StatusBadRequest, Code: 400, Msg: "Invalid Message ID"}
)

func BuildBizError(httpStatus int, code int, msg string) *response.BizError {
	return &response.BizError{HttpStatus: httpStatus, Code: code, Msg: msg}
}
