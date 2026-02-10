package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	my_manager "session-demo/manager"
	my_models "session-demo/models"
	my_requests "session-demo/requests"
	my_service "session-demo/service"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
)

var streamManager = &my_manager.StreamManager{
	Streams: make(map[string]*my_manager.StreamState),
}

func StreamChatHandler(req *restful.Request, resp *restful.Response) {
	sessionID := req.PathParameter("sessionId")

	var reqBody my_requests.StreamChatReq
	if err := req.ReadEntity(&reqBody); err != nil {
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	// 从请求头中获取用户ID
	userID := getUserIdFromHeader(req, resp)
	if userID == "" {
		resp.WriteError(http.StatusUnauthorized, errors.New("未授权"))
		return
	}

	// 验证会话归属
	var session *my_models.Session
	var err error
	if session, err = my_service.GetSessionById(sessionID); err != nil {
		resp.WriteError(http.StatusForbidden, err)
		return
	} else if session.UserID != userID {
		resp.WriteError(http.StatusForbidden, errors.New("会话归属错误"))
		return
	}

	// 获取或创建流状态
	stream := streamManager.GetOrCreateStream(sessionID, "", reqBody.QueryInfo.Query)

	// go my_service.StreamChat(stream, reqBody, session, resp)

	// 5. 生成或获取客户端ID
	// X-Client-ID 用于区分同一用户的不同连接（如多标签页）
	// 如果客户端未提供，则由服务器生成
	clientID := req.HeaderParameter("X-Client-ID")
	if clientID == "" {
		clientID = uuid.NewString()
	}

	// 7. 注册客户端接收通道
	// 将当前连接注册到流管理器中，以便接收广播消息
	// 返回一个只读通道 chunkChan，用于接收 StreamChunk
	chunkChan := streamManager.RegisterClient(stream, clientID)
	if chunkChan == nil {
		resp.WriteError(http.StatusInternalServerError, errors.New("无法注册客户端"))
		return
	}

	// 确保在函数退出（连接断开）时注销客户端
	defer streamManager.UnregisterClient(stream.SessionID, clientID)

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
	my_service.SendSSE(writer, flusher, "connected", map[string]any{
		"message_id": stream.MessageID,
		"session_id": stream.SessionID,
	})

	// 9. 处理新任务
	// 启动一个新的 goroutine 来生成 LLM 响应
	go my_service.StreamChat(stream, reqBody, session, resp)
	// 10. 监听消息并推送
	// 创建一个带有取消功能的上下文，用于监听客户端断开连接
	ctx, cancel := context.WithCancel(req.Request.Context())
	defer cancel()

	for {
		select {
		// 监听来自 streamManager 的消息通道
		case chunk, ok := <-chunkChan:
			if !ok {
				// 通道关闭，通常意味着流被管理器强制关闭或发生错误
				my_service.SendSSE(writer, flusher, "complete", map[string]any{
					"message_id":   stream.MessageID,
					"session_id":   stream.SessionID,
					"full_content": stream.FullResponse,
				})
				log.Println("StreamChatHandler:  !ok  :", stream.MessageID)
				return
			}

			// 处理结束标志
			if chunk.IsFinal {
				my_service.SendSSE(writer, flusher, "complete", map[string]any{
					"message_id":   stream.MessageID,
					"session_id":   stream.SessionID,
					"full_content": stream.FullResponse,
					"chunk_id":     chunk.ChunkID,
					"content":      chunk.Content,
					"ok":           "true",
					"is_final":     "true",
				})
				log.Println("StreamChatHandler:  is_final  :", stream.MessageID)
				return
			}

			// 发送普通的数据分块 (chunk)
			my_service.SendSSE(writer, flusher, "chunk", map[string]any{
				"session_id": stream.SessionID,
				"message_id": stream.MessageID,
				"chunk_id":   chunk.ChunkID,
				"content":    chunk.Content,
				"is_final":   false,
			})
			log.Println("StreamChatHandler:  chunk  :", stream.MessageID)

		// 监听客户端连接状态
		case <-ctx.Done():
			// 客户端断开连接（如关闭浏览器标签页）
			// 循环退出，触发 defer streamManager.UnregisterClient
			return
		}
	}

}
