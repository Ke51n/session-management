package handler

import (
	constant "session-demo/const"
	"session-demo/pkg/auth"
	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

// NewStreamChatHandler 新对话
func NewStreamChatHandler(req *restful.Request, resp *restful.Response) {

	service.NewStreamChatInSession(req, resp)

}

// ResumeStreamChatHandler 续传对话
func ResumeStreamChatHandler(req *restful.Request, resp *restful.Response) {
	sessionID := req.PathParameter("sessionId")
	userId := auth.GetUserID(req)

	//鉴权
	_, err := service.QuerySession(userId, sessionID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	var reqBody requests.ResumeStreamChatReq
	if err := req.ReadEntity(&reqBody); err != nil {
		response.WriteBizError(resp, constant.ErrBadRequest)
		return
	}

	service.ResumeStreamChat(userId, sessionID, reqBody, req, resp)
}

// BreakStreamChatHandler 中断流
func BreakStreamChatHandler(req *restful.Request, resp *restful.Response) {
	sessionID := req.PathParameter("sessionId")
	var reqBody requests.BreakStreamChatReq
	if err := req.ReadEntity(&reqBody); err != nil {
		response.WriteBizError(resp, constant.ErrBadRequest)
		return
	}

	//验证会话归属
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
		response.WriteBizError(resp, err)
		return
	}
	response.SuccessResp(nil)
}
