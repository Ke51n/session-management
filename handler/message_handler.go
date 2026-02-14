package handler

import (
	"log"
	"net/http"
	constant "session-demo/const"
	"session-demo/models"
	"session-demo/pkg/auth"
	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

// NewChatHandler 新对话
func NewChatHandler(req *restful.Request, resp *restful.Response) {

	reqBody, err := service.BindRequestBody[requests.StreamChatReq](req)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	userId := auth.GetUserID(req)
	sessionID := req.PathParameter("sessionId")

	streamChatDto := models.StreamChatDto{
		UserId:    userId,
		SessionId: sessionID,
		LastMsgID: reqBody.LastMsgID,
		ProjectID: reqBody.ProjectId,
		Query:     reqBody.QueryInfo.Query,
		Files:     reqBody.QueryInfo.Files,
		Req:       req,
		Resp:      resp,
	}

	if err := service.NewStreamChatInSession(streamChatDto); err != nil {
		response.WriteBizError(resp, err)
		return
	}
	response.WriteSuccess(resp, http.StatusOK, nil)

}

// ResumeStreamChatHandler 续传对话
func ResumeStreamChatHandler(req *restful.Request, resp *restful.Response) {
	//获取会话ID
	sessionID := req.PathParameter("sessionId")
	//获取用户ID
	userId := auth.GetUserID(req)

	//解析请求体
	reqBody, err := service.BindRequestBody[requests.ResumeStreamChatReq](req)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	log.Printf("Resume request: sessionId=%s, body=%+v", sessionID, reqBody)

	//调用服务层
	err = service.ResumeStreamChat(userId, sessionID, reqBody, req, resp)
	if err != nil {
		log.Println("ResumeStreamChatHandler error:", err)
		response.WriteBizError(resp, err)
		return
	}
	//返回响应
	response.WriteSuccess(resp, http.StatusOK, nil)
}

// BreakStreamChatHandler 中断流
func BreakStreamChatHandler(req *restful.Request, resp *restful.Response) {

	// 1. 解析参数
	reqBody, err := service.BindRequestBody[requests.BreakStreamChatReq](req)

	// 2. 统一处理解析错误 (Handler 负责 HTTP 响应)
	if err != nil {
		// 这里你可以写入 400 Bad Request
		response.WriteBizError(resp, err)
		return
	}

	//验证会话归属
	sessionID := req.PathParameter("sessionId")
	userId := auth.GetUserID(req)
	if _, err := service.QuerySession(userId, sessionID); err != nil {
		response.WriteBizError(resp, err)
		return
	}

	// 中断流
	exists, err := service.GlobalStreamManager.BreakStream(sessionID, reqBody.MessageID)
	if !exists {
		response.WriteBizError(resp, constant.ErrStreamNotFound)
		return
	}
	if err != nil {
		log.Println("BreakStreamChatHandler error:", err)
		response.WriteBizError(resp, constant.ErrInternalServer)
		return
	}
	response.WriteSuccess(resp, http.StatusOK, nil)
}
