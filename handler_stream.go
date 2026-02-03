// handler_stream.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	my_models "session-demo/models"
	my_service "session-demo/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 辅助函数：发送SSE事件
func sendSSEEvent(c *gin.Context, flusher http.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)

	// 正确的SSE格式：event: <name>\ndata: <json>\n\n
	fmt.Fprintf(c.Writer, "event: %s\n", event)
	fmt.Fprintf(c.Writer, "data: %s\n\n", string(jsonData))

	flusher.Flush()
}

// 流式对话接口（支持恢复）
// 该方法处理客户端的流式对话请求，支持断点续传。
// 核心逻辑包括：参数校验、SSE连接建立、流状态管理（创建或恢复）、
// 客户端注册、实时消息推送以及连接断开处理。
func handleDialogStreamWithResumeApi(c *gin.Context) {
	var req DialogRequest

	// 1. 绑定请求参数
	// 解析 JSON 请求体到 DialogRequest 结构中，如果失败则返回 400 错误
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 调用核心处理函数
	handleDialogStreamWithResumeInner(c, req)
}

func handleDialogStreamWithResumeInner(c *gin.Context, req DialogRequest) {

	// 2. 验证会话
	// 检查 SessionID 是否存在，是否属于当前用户，且未被删除
	// 如果验证失败，返回 404 错误
	var session my_models.Session
	if my_service.My_dbservice.DB.Where("id = ? AND user_id = ? AND deleted = ?",
		req.SessionID, req.UserID, false).First(&session).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	// 3. 设置SSE响应头
	// 必须设置 Content-Type 为 text/event-stream
	// Cache-Control: no-cache 防止缓存
	// Connection: keep-alive 保持长连接
	// X-Accel-Buffering: no 防止 Nginx 等代理服务器缓冲响应
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// 4. 检查是否支持 Flusher
	// http.Flusher 用于将缓冲区的内容立即发送给客户端，是 SSE 的基础
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "不支持流式"})
		return
	}

	// 立即发送响应头，建立连接
	c.Status(http.StatusOK)
	flusher.Flush()

	// 5. 生成或获取客户端ID
	// X-Client-ID 用于区分同一用户的不同连接（如多标签页）
	// 如果客户端未提供，则由服务器生成
	clientID := c.GetHeader("X-Client-ID")
	if clientID == "" {
		clientID = uuid.NewString()
	}

	// 6. 获取或创建流状态
	// streamManager 负责维护全局的流状态
	// 如果是新请求，会创建新流；如果是续传，会返回现有流
	stream := streamManager.GetOrCreateStream(req.SessionID, req.Query)

	// 7. 注册客户端接收通道
	// 将当前连接注册到流管理器中，以便接收广播消息
	// 返回一个只读通道 chunkChan，用于接收 StreamChunk
	chunkChan := streamManager.RegisterClient(stream, clientID)
	if chunkChan == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法注册客户端"})
		return
	}
	// 确保在函数退出（连接断开）时注销客户端
	defer streamManager.UnregisterClient(stream.SessionID, clientID)

	// 8. 发送连接成功事件
	// 告知客户端连接已建立，并返回会话和消息ID信息
	// is_resume 字段指示当前是否为续传模式
	sendSSEEvent(c, flusher, "connected", gin.H{
		"message_id": stream.MessageID,
		"session_id": stream.SessionID,
		"timestamp":  time.Now().Unix(),
	})

	// 9. 处理新任务或续传逻辑
	if len(stream.Chunks) == 0 {
		// CASE A: 新任务
		// 启动一个新的 goroutine 来生成 LLM 响应
		go generateStreamResponse(stream, req)

		// 发送 start 事件，告知客户端生成开始
		sendSSEEvent(c, flusher, "start", gin.H{
			"message_id": stream.MessageID,
			"query":      stream.Query,
			"timestamp":  time.Now().Unix(),
		})
	} else {
		// CASE B: 续传/恢复
		// 发送 resume 事件，包含已生成的完整历史内容 (history_output)
		// 客户端可以直接展示 history_output，然后等待后续 chunk
		sendSSEEvent(c, flusher, "resume", gin.H{
			"message_id":     stream.MessageID,
			"total_chunks":   len(stream.Chunks),
			"timestamp":      time.Now().Unix(),
			"history_output": stream.Chunks, // 把已经生成的内容返回（包括前端中断时候的内容）
		})
	}

	// 10. 监听消息并推送
	// 创建一个带有取消功能的上下文，用于监听客户端断开连接
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	for {
		select {
		// 监听来自 streamManager 的消息通道
		case chunk, ok := <-chunkChan:
			if !ok {
				// 通道关闭，通常意味着流被管理器强制关闭或发生错误
				sendSSEEvent(c, flusher, "complete", gin.H{
					"message_id":   stream.MessageID,
					"full_content": stream.FullResponse,
					"timestamp":    time.Now().Unix(),
					"ok":           false,
				})
				return
			}

			// 处理结束标志
			if chunk.IsFinal {
				sendSSEEvent(c, flusher, "complete", gin.H{
					"message_id":   stream.MessageID,
					"full_content": stream.FullResponse,
					"timestamp":    time.Now().Unix(),
					"chunk_id":     chunk.ChunkID,
					"content":      chunk.Content,
					"ok":           true,
					"is_final":     true,
				})
				return
			}

			// 发送普通的数据分块 (chunk)
			sendSSEEvent(c, flusher, "chunk", gin.H{
				"message_id": stream.MessageID,
				"chunk_id":   chunk.ChunkID,
				"content":    chunk.Content,
				"is_final":   false,
				"timestamp":  time.Now().Unix(),
			})

		// 监听客户端连接状态
		case <-ctx.Done():
			// 客户端断开连接（如关闭浏览器标签页）
			// 循环退出，触发 defer streamManager.UnregisterClient
			return
		}
	}
}

// 生成流式响应的协程
func generateStreamResponse(stream *StreamState, req DialogRequest) {
	// 查找历史消息
	var historyMessages []my_models.Message
	my_service.My_dbservice.DB.Where("session_id = ? AND deleted = ?", req.SessionID, false).
		Order("created_at ASC").
		Find(&historyMessages)

	prompt := buildPrompt(historyMessages, req.Query)

	// 保存用户消息到数据库
	userMsg := my_models.Message{
		ID:        uuid.NewString(),
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Query,
		Files:     req.Files,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"source": "new_ask",
		},
	}
	my_service.My_dbservice.DB.Create(&userMsg)

	// 生成chunks
	chunks := mockLLMStream(prompt)

	// 逐步发送chunks
	for i, chunk := range chunks {
		time.Sleep(200 * time.Millisecond) // 模拟生成延迟

		// 保存到流状态
		streamManager.AddChunk(stream.SessionID, chunk)

		// // 更新已发送索引（需要在锁外进行广播）
		// stream.mu.Lock()
		// stream.SentIndex += 1
		// stream.mu.Unlock()
		// 广播给所有客户端
		stream.mu.RLock()
		for _, clientChan := range stream.Clients {
			select {
			case clientChan <- StreamChunk{
				ChunkID: i,
				Content: chunk,
				IsFinal: i == len(chunks)-1,
			}:
			default:
				log.Println("Client channel is full, skipping")
				// 客户端可能已断开，跳过
			}
		}

		stream.mu.RUnlock()
	}

	// 标记流完成，
	defer streamManager.CompleteStream(stream.SessionID)

	// 保存助手回复到数据库
	assistantMsg := my_models.Message{
		ID:        stream.MessageID, // 使用相同的messageID
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   stream.FullResponse,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"model":     "mock_llm",
			"stream":    true,
			"resumable": true,
		},
	}
	my_service.My_dbservice.DB.Create(&assistantMsg)

	// 更新会话时间
	my_service.My_dbservice.DB.Model(&my_models.Session{ID: req.SessionID}).Update("updated_at", time.Now())
}

// 查询流状态接口
func handleStreamStatus(c *gin.Context) {
	sessionID := c.Query("session_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要session_id或message_id"})
		return
	}

	streamManager.mu.RLock()
	defer streamManager.mu.RUnlock()

	var result []gin.H

	for _, stream := range streamManager.streams {
		if sessionID != "" && stream.SessionID == sessionID {
			result = append(result, gin.H{
				"message_id":   stream.MessageID,
				"session_id":   stream.SessionID,
				"query":        stream.Query,
				"is_completed": stream.IsCompleted,
				"total_chunks": len(stream.Chunks),
				"created_at":   stream.CreatedAt,
				"updated_at":   stream.UpdatedAt,
				"client_count": len(stream.Clients),
			})
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(result),
		"streams": result,
	})
}
