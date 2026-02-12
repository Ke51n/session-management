package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	constant "session-demo/const"
	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
)

// NewStreamChatHandler 新对话
func NewStreamChatHandler(req *restful.Request, resp *restful.Response) {
	sessionID := req.PathParameter("sessionId")

	//鉴权
	var reqBody requests.StreamChatReq
	if err := req.ReadEntity(&reqBody); err != nil {
		response.WriteBizError(resp, err)
		return
	}

	service.NewStreamChatInSession(sessionID, reqBody, req, resp)

	//todo 保存用户消息
	var parentId *string
	if reqBody.LastMsgID != "" {
		parentId = &reqBody.LastMsgID
	}
	userMsgId := uuid.NewString()
	service.CreateAndSaveMessage(
		userMsgId,
		sessionID,
		parentId,
		constant.RoleUser,
		nil,
		reqBody.QueryInfo.Files,
		reqBody.QueryInfo.Query,
		len(reqBody.QueryInfo.Files),
		"completed",
		false,
		nil,
		nil,
	)
	// 获取流
	assistantMsgId := uuid.NewString()
	//先保存助手消息占位，标识状态
	service.CreateAndSaveMessage(
		assistantMsgId,
		sessionID,
		&userMsgId, "assistant", nil, nil, "",
		0, "processing", false, nil, nil)

	stream := service.GlobalStreamManager.GetOrCreateStream(sessionID, assistantMsgId, userMsgId, reqBody.QueryInfo.Query, false)

	go service.StreamChat(stream, reqBody, sessionID, resp)

	dealStreamChat(stream, false, req, resp)
	log.Println("deal stream chat: new  done :", stream.MessageID)
}

// ResumeStreamChatHandler 续传对话
func ResumeStreamChatHandler(req *restful.Request, resp *restful.Response) {
	sessionID := req.PathParameter("sessionId")

	//鉴权
	_, err := service.QuerySession(sessionID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	var reqBody requests.ResumeStreamChatReq
	if err := req.ReadEntity(&reqBody); err != nil {
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	// 获取或创建流状态
	stream := service.GlobalStreamManager.GetOrCreateStream(sessionID, reqBody.MessageID, "", "", true)
	if stream == nil {
		resp.WriteError(http.StatusNotFound, errors.New("流不存在"))
		return
	}
	dealStreamChat(stream, true, req, resp)
	log.Println("deal stream chat: resume done :", stream.MessageID)
}

func dealStreamChat(stream *service.StreamState, resume bool, req *restful.Request, resp *restful.Response) {
	// 获取客户端ID，每次请求都生成一个新的ID，防止多个客户端用同一个id同时请求导致数据混乱
	clientID := uuid.NewString()

	// 1. 设置 SSE Header
	writer := resp.ResponseWriter
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 8. 发送连接成功事件
	// 告知客户端连接已建立，并返回会话和消息ID信息
	// is_resume 字段指示当前是否为续传模式
	service.SendSSE(writer, flusher, "connected", map[string]any{
		"message_id": stream.MessageID,
		"session_id": stream.SessionID,
		"history":    stream.FullResponse,
	})
	// 7. 注册客户端接收通道
	// 将当前连接注册到流管理器中，以便接收广播消息
	// 返回一个只读通道 chunkChan，用于接收 StreamChunk
	chunkChan := service.GlobalStreamManager.RegisterClient(stream, clientID)
	if chunkChan == nil {
		resp.WriteError(http.StatusInternalServerError, errors.New("无法注册客户端"))
		return
	}
	// 10. 监听消息并推送
	// 创建一个带有取消功能的上下文，用于监听客户端断开连接
	ctx, cancel := context.WithTimeout(req.Request.Context(), 30*time.Minute)
	defer cancel()
	for {
		select {
		// 监听来自 streamManager 的消息通道
		case chunk, ok := <-chunkChan:
			if !ok {
				// 添加错误日志
				log.Printf("chunkChan closed for client %s, message %s",
					clientID, stream.MessageID)
				// 通道关闭，通常意味着流被管理器强制关闭或发生错误
				service.SendSSE(writer, flusher, "complete", map[string]any{
					"message_id":      stream.MessageID,
					"session_id":      stream.SessionID,
					"partial_content": stream.FullResponse,
				})
				log.Println("deal stream chat:  !ok  :", stream.MessageID)
				return
			}

			// 处理结束标志
			if chunk.IsCompleted || chunk.IsBreak {
				service.SendSSE(writer, flusher, "complete", map[string]any{
					"message_id":   stream.MessageID,
					"session_id":   stream.SessionID,
					"full_content": stream.FullResponse,
					"is_final":     stream.IsCompleted,
					"is_break":     stream.IsBreak,
				})
				return
			}

			// 发送普通的数据分块 (chunk)
			service.SendSSE(writer, flusher, "chunk", map[string]any{
				"session_id": stream.SessionID,
				"message_id": stream.MessageID,
				"chunk_id":   chunk.ChunkID,
				"content":    chunk.Content,
				"is_final":   false,
			})

		// 监听客户端连接状态
		case <-ctx.Done():
			// 客户端断开连接（如关闭浏览器标签页）
			// 循环退出，触发 defer streamManager.UnregisterClient
			log.Println("deal stream chat:  <-ctx.Done()  :", clientID)
			return
		}
	}
}

// BreakStreamChatHandler 中断流
func BreakStreamChatHandler(req *restful.Request, resp *restful.Response) {
	sessionID := req.PathParameter("sessionId")

	var reqBody requests.BreakStreamChatReq
	if err := req.ReadEntity(&reqBody); err != nil {
		response.WriteBizError(resp, &response.BizError{HttpStatus: http.StatusBadRequest, Msg: err.Error()})
		return
	}

	// 中断流
	exists, err := service.GlobalStreamManager.BreakStream(sessionID, reqBody.MessageID)
	if !exists {
		resp.WriteError(http.StatusNotFound, errors.New("stream not found"))
		return
	}
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, response.BreakStreamChatResponse{Success: exists, Message: "break stream success"})
}
