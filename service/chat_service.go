package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	constant "session-demo/const"
	"session-demo/models"
	my_models "session-demo/models"
	"session-demo/pkg/auth"
	"session-demo/requests"
	"session-demo/response"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
)

// 在已有会话中新对话
func NewStreamChatInSession(req *restful.Request, resp *restful.Response) {

	var reqBody requests.StreamChatReq
	if err := req.ReadEntity(&reqBody); err != nil {
		response.WriteBizError(resp, err)
		return
	}
	userId := auth.GetUserID(req)
	sessionID := req.PathParameter("sessionId")

	//检查session和lastMessageId 有效性
	_, err := GetSessionById(userId, sessionID)
	if err != nil {
		response.WriteBizError(resp, err)
		log.Println("NewStreamChatInSession: session not found:", sessionID)
		return
	}

	//检查lastMessageId 有效性
	if reqBody.LastMsgID != "" {
		_, err := GetMessageById(sessionID, reqBody.LastMsgID)
		if err != nil {
			response.WriteBizError(resp, err)
			return
		}
	}

	//保存用户消息

	//保存助手消息占位,标识processing

	var parentId *string
	if reqBody.LastMsgID != "" {
		parentId = &reqBody.LastMsgID
	}
	userMsgId := uuid.NewString()

	userMsg := &my_models.Message{
		ID:        userMsgId,
		SessionID: sessionID,
		ParentID:  parentId,

		Role:       constant.RoleUser,
		Steps:      nil,
		Files:      reqBody.QueryInfo.Files,
		Content:    reqBody.QueryInfo.Query,
		TokenCount: len(reqBody.QueryInfo.Query),

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    constant.MessageStatusCompleted,
		Deleted:   false,

		Extension: nil,
		Metadata:  nil,
	}

	if err := CreateAndSaveMessage(userMsg); err != nil {
		response.WriteBizError(resp, err)
		return
	}
	assistantMsgId := uuid.NewString()
	//先保存助手消息占位，标识状态
	assistantMsg := &my_models.Message{
		ID:        assistantMsgId,
		SessionID: sessionID,
		ParentID:  &userMsgId,

		Role:       constant.RoleAssistant,
		Steps:      nil,
		Files:      nil,
		Content:    "",
		TokenCount: 0,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    constant.MessageStatusProcessing,
		Deleted:   false,

		Extension: nil,
		Metadata:  nil,
	}
	if err := CreateAndSaveMessage(assistantMsg); err != nil {
		response.WriteBizError(resp, err)
		return
	}

	//获取流
	stream := GlobalStreamManager.GetOrCreateStream(sessionID, assistantMsgId, userMsgId, reqBody.QueryInfo.Query, false)

	// 启动流式对话处理
	go StreamChatStarter(userId, stream, reqBody, sessionID, resp)

	dealStreamResponse(stream, false, req, resp)
	log.Println("deal stream chat: new  done :", stream.MessageID)
}

// StreamChatStarter 启动流式对话处理
func StreamChatStarter(userId string, stream *StreamState, reqBody requests.StreamChatReq, sessionID string, resp *restful.Response) {

	// 构造最终prompt
	prompt := buildFinalPrompt(userId, reqBody, sessionID)
	_ = prompt
	log.Println("Final Prompt:", prompt)

	// 调用模型层处理流式对话
	// stream := my_models.StreamChat(reqBody)

	streamChatInner(stream, prompt, sessionID)

}

// buildFinalPrompt 构建最终prompt
func buildFinalPrompt(userId string, reqBody requests.StreamChatReq, sessionID string) string {
	//构造历史上下文
	history := buildHistoryContext(sessionID, reqBody.LastMsgID)

	//查询session
	session, err := GetSessionById(userId, sessionID)
	if err != nil {
		return ""
	}

	//获取项目级别的prompt模板
	project, err := GetProjectById(session.ProjectID)
	if err != nil {
		return ""
	}

	// 合并历史上下文和当前查询
	finalPrompt := "系统指令:" + project.CustomInstruction + "\n\n对话历史:" + history + "\n\n当前问题:" + reqBody.QueryInfo.Query

	return finalPrompt
}

// buildHistoryContext 构建历史上下文, 从lastMsgID开始, 直到根消息
func buildHistoryContext(sessionID string, lastMsgID string) string {
	if lastMsgID == "" {
		return ""
	}
	// 从数据库查询历史消息
	var messages []models.Message
	Dbservice.DB.Where("session_id = ?", sessionID).Order("id").Find(&messages)

	msgMap := make(map[string]models.Message)
	for _, msg := range messages {
		msgMap[msg.ID] = msg
	}

	// 构建历史上下文
	var historyContext strings.Builder
	var messageId = &lastMsgID
	for messageId != nil && msgMap[*messageId].ParentID != nil {
		msg := msgMap[*messageId]
		historyContext.WriteString(msg.Role + ": " + msg.Content + "\n")
		messageId = msg.ParentID
	}

	return historyContext.String()
}

func streamChatInner(stream *StreamState, prompt string, sessionID string) {
	// 调用模型层处理流式对话
	// stream := my_models.StreamChat(prompt)

	// 5. 模拟 AI 流式回复
	reply := `春江潮水连海平，海上明月共潮生①。
	滟滟随波千万里②，何处春江无月明！
	江流宛转绕芳甸③，月照花林皆似霰④。
	空里流霜不觉飞⑤，汀上白沙看不见⑥。
	江天一色无纤尘⑦，皎皎空中孤月轮⑧。
	江畔何人初见月？江月何年初照人？`

	for i, ch := range reply {

		content := string(ch)
		stream.Mu.Lock()
		stream.Chunks = append(stream.Chunks, content)
		stream.UpdatedAt = time.Now()
		stream.FullResponse += content
		stream.Mu.Unlock()

		chunk := StreamChunk{
			ChunkID: i,
			Content: string(ch),
		}
		broadcastChunk(stream, chunk)
		time.Sleep(500 * time.Millisecond)
	}
	GlobalStreamManager.CompleteStream(stream.SessionID + "_" + stream.MessageID)
	// 6. 结束标记
	log.Println("Completed StreamChatService for key:", stream.SessionID+"_"+stream.MessageID)

}
func broadcastChunk(stream *StreamState, chunk StreamChunk) {
	stream.Mu.RLock() // 只需要读锁
	defer stream.Mu.RUnlock()

	for _, clientChan := range stream.Clients {
		select {
		case clientChan <- chunk: // 只读取map，不修改它
			// 发送成功
		default:
			log.Println("Client channel is full, skipping")
			// 客户端可能已断开，跳过
			// 即使这里发现客户端有问题，也不立即删除
			// 可以标记或异步处理
		}
	}
}

// dealStreamResponse 处理流式响应
func dealStreamResponse(stream *StreamState, resume bool, req *restful.Request, resp *restful.Response) {
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
	SendSSE(writer, flusher, "connected", map[string]any{
		"message_id": stream.MessageID,
		"session_id": stream.SessionID,
		"history":    stream.FullResponse,
	})
	// 7. 注册客户端接收通道
	// 将当前连接注册到流管理器中，以便接收广播消息
	// 返回一个只读通道 chunkChan，用于接收 StreamChunk
	chunkChan := GlobalStreamManager.RegisterClient(stream, clientID)
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
				SendSSE(writer, flusher, "complete", map[string]any{
					"message_id":      stream.MessageID,
					"session_id":      stream.SessionID,
					"partial_content": stream.FullResponse,
				})
				log.Println("deal stream chat:  !ok  :", stream.MessageID)
				return
			}

			// 处理结束标志
			if chunk.IsCompleted || chunk.IsBreak {
				SendSSE(writer, flusher, "complete", map[string]any{
					"message_id":   stream.MessageID,
					"session_id":   stream.SessionID,
					"full_content": stream.FullResponse,
					"is_final":     stream.IsCompleted,
					"is_break":     stream.IsBreak,
				})
				return
			}

			// 发送普通的数据分块 (chunk)
			SendSSE(writer, flusher, "chunk", map[string]any{
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

func ResumeStreamChat(userId, sessionID string, reqBody requests.ResumeStreamChatReq, req *restful.Request, resp *restful.Response) {
	// 验证会话归属
	if _, err := QuerySession(userId, sessionID); err != nil {
		log.Println("failed to query session:", err)
		response.WriteBizError(resp, err)
		return
	}

	// 获取或创建流状态
	stream := GlobalStreamManager.GetOrCreateStream(sessionID, reqBody.MessageID, "", "", true)
	if stream == nil {
		resp.WriteError(http.StatusNotFound, errors.New("流不存在"))
		return
	}
	dealStreamResponse(stream, true, req, resp)
	log.Println("deal stream chat: resume done :", stream.MessageID)

}
