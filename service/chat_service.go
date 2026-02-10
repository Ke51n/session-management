package service

import (
	"log"
	"net/http"
	my_manager "session-demo/manager"
	my_models "session-demo/models"
	my_requests "session-demo/requests"
	my_utils "session-demo/utils"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
)

func StreamChat(stream *my_manager.StreamState, reqBody my_requests.StreamChatReq, session *my_models.Session, resp *restful.Response) {

	// 从请求体中获取上一条消息ID
	lastMsgID := reqBody.LastMsgID

	// 生成当前query消息ID
	queryMsgId := uuid.New().String()

	//保存用户请求
	if err := CreateAndSaveMessage(queryMsgId, session.ID, &lastMsgID, my_utils.RoleUser,
		nil, reqBody.QueryInfo.Files, reqBody.QueryInfo.Query,
		len(reqBody.QueryInfo.Query), false, nil, nil); err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// 构造最终prompt
	prompt := buildFinalPrompt(reqBody, session.ID)
	_ = prompt
	log.Println("Final Prompt:", prompt)

	// 调用模型层处理流式对话
	// stream := my_models.StreamChat(reqBody)

	dealStreamChat(stream, prompt, session.ID, resp)

}

func buildFinalPrompt(reqBody my_requests.StreamChatReq, sessionID string) string {
	//构造历史上下文
	history := buildHistoryContext(sessionID, reqBody.LastMsgID)

	//查询session
	session, err := GetSessionById(sessionID)
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
	var messages []my_models.Message
	My_dbservice.DB.Where("session_id = ?", sessionID).Order("id").Find(&messages)

	msgMap := make(map[string]my_models.Message)
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

func dealStreamChat(stream *my_manager.StreamState, prompt string, sessionID string, resp *restful.Response) {
	// 调用模型层处理流式对话
	// stream := my_models.StreamChat(prompt)

	// 1. 设置 SSE Header
	// writer := resp.ResponseWriter

	// writer.Header().Set("Content-Type", "text/event-stream")
	// writer.Header().Set("Cache-Control", "no-cache")
	// writer.Header().Set("Connection", "keep-alive")

	// flusher, ok := writer.(http.Flusher)
	// if !ok {
	// http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
	// return
	// }
	// 4. SSE 发送 session_id和message_id
	// msgId := uuid.New().String()
	// SendSSE(writer, flusher, "id", map[string]any{"session_id": sessionID, "message_id": msgId})

	// 5. 模拟 AI 流式回复
	reply := `春江潮水连海平，海上明月共潮生①。
	滟滟随波千万里②，何处春江无月明！
	江流宛转绕芳甸③，月照花林皆似霰④。
	空里流霜不觉飞⑤，汀上白沙看不见⑥。
	江天一色无纤尘⑦，皎皎空中孤月轮⑧。
	江畔何人初见月？江月何年初照人？`

	for i, ch := range reply {
		stream.Mu.Lock()
		stream.Chunks = append(stream.Chunks, string(ch))
		stream.UpdatedAt = time.Now()
		stream.FullResponse += string(ch)
		// SendSSE(writer, flusher, "message", map[string]any{"content": string(ch)})

		// 广播给所有客户端
		for _, clientChan := range stream.Clients {
			select {
			case clientChan <- my_manager.StreamChunk{
				ChunkID: i,
				Content: string(ch),
			}:
			default:
				log.Println("Client channel is full, skipping")
				// 客户端可能已断开，跳过
			}
		}

		stream.Mu.Unlock()

		time.Sleep(1000 * time.Millisecond)
	}
	my_manager.GlobalStreamManager.CompleteStream(stream.SessionID + "_" + stream.MessageID)
	// 6. 结束标记
	// SendSSE(writer, flusher, "done", map[string]any{"done": "true"})
	log.Println("Completed StreamChatService for sessionID:", sessionID)

}
