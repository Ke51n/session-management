package utils

// import (
// 	"session-demo/response"

// 	"github.com/emicklei/go-restful/v3"
// )

// // RouteConfig 定义一个配置函数类型
// type RouteConfig func(*restful.RouteBuilder)

// // CommonConfig 封装所有接口都有的公共配置（错误返回、公共参数等）
// func CommonConfig() RouteConfig {
// 	return func(b *restful.RouteBuilder) {
// 		b.Consumes(restful.MIME_JSON).
// 			// 统一配置公共错误返回
// 			Returns(400, "Bad Request", response.CommonResponse{}).
// 			Returns(401, "Unauthorized", response.CommonResponse{}).
// 			Returns(500, "Internal Server Error", response.CommonResponse{})
// 	}
// }

// // SSERoute 专门用于 SSE 流式接口的配置
// // 参数：文档说明、请求体类型、处理器
// func SSERoute(doc string, reqType string, handler restful.RouteFunction) RouteConfig {
// 	return func(b *restful.RouteBuilder) {
// 		b.To(handler).
// 			Doc(doc).
// 			Produces("text/event-stream"). // SSE 特有属性
// 			Param(ws.PathParameter("sessionId", "Session ID").DataType("string").Required(true)).
// 			Param(ws.BodyParameter("request", "Request Body").DataType(reqType)).
// 			Returns(200, "OK", nil) // SSE 成功响应
// 		// 应用公共配置
// 		CommonConfig()(b)
// 	}
// }

// // StandardRoute 专门用于普通 JSON 接口的配置
// // 参数：文档说明、请求体类型、响应体模型、处理器
// func StandardRoute(doc string, reqType string, respModel interface{}, handler restful.RouteFunction) RouteConfig {
// 	return func(b *restful.RouteBuilder) {
// 		b.To(handler).
// 			Doc(doc).
// 			Produces(restful.MIME_JSON). // 普通 JSON 接口
// 			Param(ws.PathParameter("sessionId", "Session ID").DataType("string").Required(true)).
// 			Param(ws.BodyParameter("request", "Request Body").DataType(reqType)).
// 			Returns(200, "OK", respModel)
// 		// 应用公共配置
// 		CommonConfig()(b)
// 	}
// }

// // Register 封装最终的注册动作（辅助函数，减少调用代码量）
// func Register(builder *restful.RouteBuilder, configs ...RouteConfig) {
// 	for _, config := range configs {
// 		config(builder)
// 	}
// 	ws.Route(builder)
// }
