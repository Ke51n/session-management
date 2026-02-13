package models

import "github.com/emicklei/go-restful/v3"

// StreamChatDto 流式对话内部请求体
type StreamChatDto struct {
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`

	LastMsgID string `json:"lastMsgID"`

	ProjectID string `json:"project_id"`
	Query     string `json:"query"`
	Files     []File `json:"files"`

	Req  *restful.Request  `json:"-"`
	Resp *restful.Response `json:"-"`
}
