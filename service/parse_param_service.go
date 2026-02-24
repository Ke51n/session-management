package service

import (
	"log"
	"net/http"
	"session-demo/response"

	"github.com/emicklei/go-restful/v3"
)

// 定义一个接口，约束传入的 req 必须有 ReadEntity 方法
type EntityReader interface {
	ReadEntity(v any) error
}

// service/param.go
// 这个函数只负责解析，不负责写响应，保持 Service/Helper 层的纯净
func BindRequestBody[T any](req *restful.Request) (*T, error) {
	var body T
	// 注意：这里我们假设框架传进来的是 *restful.Request
	if err := req.ReadEntity(&body); err != nil {
		return nil, response.WrapError(http.StatusBadRequest, "参数解析失败", err)
	}
	log.Println("BindRequestBody body:", body)
	return &body, nil
}
