package service

import (
	"log"
	my_models "session-demo/models"
	my_requests "session-demo/requests"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
)

func StreamChat(stream *StreamState, reqBody my_requests.StreamChatReq, sessionID string, resp *restful.Response) {

	// 构造最终prompt
	prompt := buildFinalPrompt(reqBody, sessionID)
	_ = prompt
	log.Println("Final Prompt:", prompt)

	// 调用模型层处理流式对话
	// stream := my_models.StreamChat(reqBody)

	streamChatInner(stream, prompt, sessionID)

}

// buildFinalPrompt 构建最终prompt
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
