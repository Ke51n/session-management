package handler

import (
	"net/http"
	auth "session-demo/pkg/auth"
	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

// 创建一个会话并对话，sse流式响应CreateSessioAndChatHandler
func CreateSessioAndChatHandler(req *restful.Request, resp *restful.Response) {
	userID := auth.GetUserID(req)

	// 1. 解析参数
	reqData, err := service.BindRequestBody[requests.CreateSessionAndChatReq](req)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	//服务层
	if err := service.CreateSessionAndChat(userID, reqData, req, resp); err != nil {
		response.WriteBizError(resp, err)
		return
	}

	response.WriteSuccess(resp, http.StatusOK, nil)
	// //模型消息写数据库
	// var steps []models.StepNode = []models.StepNode{
	// 	{
	// 		Type: "thought",
	// 		Text: "我需要先思考一下。",
	// 	},
	// 	{
	// 		Type: "plan",
	// 		Text: "我计划先检索相关信息，然后生成回答。",
	// 	}}
	// if strings.Contains(reqData.Query, "天气") {
	// 	steps = []models.StepNode{
	// 		{
	// 			Type: "tool_call",
	// 			Name: "get_weather",
	// 			Text: "{'id':'tool_call_1','tool':'get_weather','args':{'location':'Beijing'}}",
	// 		},
	// 		{
	// 			Type: "tool_return",
	// 			Name: "get_weather",
	// 			Text: "天气信息：上海当前温度为15度，多云。",
	// 		},
	// 	}
	// }
	// if strings.Contains(reqData.Query, "苹果") {
	// 	steps = append(steps, models.StepNode{
	// 		Type: "tool_call",
	// 		Name: "read_file",
	// 		Text: "{'id':'tool_call_1','tool':'read_file','args':{'file_path':'/path/to/apple.txt'}}",
	// 	})
	// 	steps = append(steps, models.StepNode{
	// 		Type: "tool_return",
	// 		Name: "read_file",
	// 		Text: "苹果是一种红色的水果，通常用于 pies。",
	// 	})
	// }

}

// 查询某个会话的所有消息
func ListMessagesBySessionHandler(req *restful.Request, resp *restful.Response) {

	userID := auth.GetUserID(req)
	sessionID := req.PathParameter("sessionId")
	// 调用服务层
	messages, err := service.ListMessagesBySession(userID, sessionID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	response.WriteSuccess(resp, http.StatusOK, response.SuccessResp(messages))
}

func ListSessionsNotInProjectHandler(req *restful.Request, resp *restful.Response) {

	userID := auth.GetUserID(req)
	// 调用服务层
	sessions, err := service.ListSessionsNotInProject(userID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	response.WriteSuccess(resp, http.StatusOK, response.SuccessResp(sessions))
}

// / MoveSessionToProjectHandler 移动一个会话到某个指定项目
func MoveSessionToProjectHandler(req *restful.Request, resp *restful.Response) {

	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)
	sessionID := req.PathParameter("sessionId")

	reqData, err := service.BindRequestBody[requests.MoveSessionToProjectReq](req)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	// 调用服务层
	err = service.MoveSessionToProject(userID, sessionID, reqData.ProjectID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	response.WriteSuccess(resp, http.StatusOK, response.SuccessResp(nil))

}

func UpdateSessionHandler(req *restful.Request, resp *restful.Response) {

	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)
	sessionID := req.PathParameter("sessionId")

	reqData, err := service.BindRequestBody[requests.UpdateSessionReq](req)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	// 调用服务层
	err = service.UpdateSession(userID, sessionID, reqData.Title)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	response.WriteSuccess(resp, http.StatusOK, response.SuccessResp(nil))

}
