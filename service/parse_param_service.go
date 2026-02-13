package service

import (
	"net/http"
	constant "session-demo/const"
	"session-demo/response"

	"github.com/emicklei/go-restful/v3"
)

// 定义一个接口，约束传入的 req 必须有 ReadEntity 方法
type EntityReader interface {
	ReadEntity(v interface{}) error
}

// 泛型函数签名修改如下
func ReadAndBindReq[T any](req EntityReader, resp *restful.Response) (*T, bool) {
	// 1. 创建一个 T 类型的实例
	// 因为 T 是 any (interface{})，我们需要 new(T) 来分配内存并返回指针
	var reqBody T

	// 2. 这里的 req.ReadEntity 是你使用的特定框架的方法
	// 注意：如果为了通用性，你可能需要传入一个能调用 ReadEntity 的接口
	if err := req.ReadEntity(&reqBody); err != nil {
		// 3. 统一错误处理
		response.WriteBizError(resp, constant.BuildBizError(http.StatusBadRequest, 400, err.Error()))

		// 返回 nil 和 false 表示失败，调用方应立即 return
		return nil, false
	}

	// 4. 成功返回
	return &reqBody, true
}

// service/param.go
// 这个函数只负责解析，不负责写响应，保持 Service/Helper 层的纯净
func BindRequestBody[T any](req *restful.Request) (*T, error) {
	var body T
	// 注意：这里我们假设框架传进来的是 *restful.Request
	if err := req.ReadEntity(&body); err != nil {
		return nil, response.WrapError(http.StatusBadRequest, "参数解析失败", err)
	}
	return &body, nil
}
