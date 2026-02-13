package response

import (
	"fmt"
	"log"
	"net/http"
	"session-demo/models"

	"github.com/emicklei/go-restful/v3"
)

type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func SuccessResp(data any) *CommonResponse {
	return &CommonResponse{
		Code:    0, // 惯例：0 代表成功
		Message: "success",
		Data:    data,
	}
}
func ErrorResp(code int, message string) *CommonResponse {
	return &CommonResponse{
		Code:    code,
		Message: message,
		Data:    nil, // 错误通常没有数据负载
	}
}

// BizError 定义业务错误接口，方便 Service 层返回具体错误
type BizError struct {
	HttpStatus int    `json:"http_status"`
	Code       int    `json:"code"`
	Msg        string `json:"message"`
}

func (e *BizError) Error() string {
	return e.Msg
}

// WrapError 是一个统一入口
// 它做了三件事：
// 1. 打印详细错误日志（供后台看）
// 2. 包装成一个对外的 BizError（供前台看）
// 3. 隐藏敏感细节
func WrapError(businessCode int, publicMsg string, internalErr error) *BizError {
	// 只有当 internalErr 不为 nil 时才处理
	if internalErr == nil {
		return nil
	}
	// 这里可以接入你项目的日志框架，自动带上堆栈信息
	// 例如：log.Errorf("Internal Error: %+v", internalErr)
	fmt.Printf("[ERROR] %v\n", internalErr)
	return &BizError{
		HttpStatus: http.StatusInternalServerError,
		Code:       businessCode,
		Msg:        publicMsg,
	}
}

// 统一错误响应写入
func WriteBizError(resp *restful.Response, err error) {
	if bizErr, ok := err.(*BizError); ok {
		resp.WriteHeaderAndEntity(bizErr.HttpStatus, CommonResponse{
			Code:    bizErr.Code,
			Message: bizErr.Msg,
		})
		return
	}
	log.Printf("Unhandled error type: %T, message: %s", err, err.Error())
	// 非业务错误，统一视为内部错误，不暴露细节给前端
	resp.WriteHeaderAndEntity(http.StatusInternalServerError, CommonResponse{
		Code:    -1,
		Message: "Internal Server Error",
	})
}

func WriteSuccess(resp *restful.Response, httpStatus int, data any) {
	resp.WriteHeaderAndEntity(httpStatus, SuccessResp(data))
}

// 查询某个项目下的所有会话响应结构
type ListSessionsResponse struct {
	Data    []models.Session `json:"data"`
	Success bool             `json:"success"`
}

// 创建项目响应结构
type CreateOrEditProjectResponse struct {
	Success   bool   `json:"success"`
	ProjectID string `json:"project_id"`
}

// 查询所有项目响应结构
type ListProjectsResponse struct {
	Data    []models.Project `json:"data"`
	Success bool             `json:"success"`
}

// 更新项目标题响应结构
type UpdateProjectResponse struct {
	Success bool `json:"success"`
}

// 删除项目响应结构
type DeleteProjectResponse struct {
	Success bool `json:"success"`
}
type ListMessagesResponse struct {
	Data    []models.Message `json:"data"`
	Success bool             `json:"success"`
}

// MoveSessionToProjectResponse 移动会话到项目响应结构
type MoveSessionToProjectResponse struct {
	Success bool `json:"success"`
}

type UpdateSessionTitleResponse struct {
	Success bool `json:"success"`
}

// BreakStreamChatResponse 中断流式对话响应结构
type BreakStreamChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
