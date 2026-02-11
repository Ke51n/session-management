package manager

// import (
// 	"errors"
// 	"log"
// 	"sync"
// 	"time"

// 	my_models "session-demo/models"
// 	my_service "session-demo/service"

// 	"github.com/google/uuid"
// )

// // StreamState 流式生成状态
// type StreamState struct {
// 	SessionID    string                      `json:"session_id"`    // 会话ID
// 	MessageID    string                      `json:"message_id"`    // 消息ID
// 	Query        string                      `json:"query"`         // 用户查询
// 	FullResponse string                      `json:"full_response"` // 完整响应（逐步构建）
// 	Chunks       []string                    `json:"chunks"`        // 所有chunk
// 	IsBreak      bool                        `json:"is_break"`      // 是否中断
// 	IsCompleted  bool                        `json:"is_completed"`  // 是否完成
// 	CreatedAt    time.Time                   `json:"created_at"`    // 创建时间
// 	UpdatedAt    time.Time                   `json:"updated_at"`    // 更新时间
// 	Clients      map[string]chan StreamChunk // 连接的客户端
// 	Mu           sync.RWMutex                `json:"-"` // 添加互斥锁

// }

// // StreamChunk 流式chunk
// type StreamChunk struct {
// 	ChunkID     int    `json:"chunk_id"`     // 分块ID
// 	Content     string `json:"content"`      // 内容
// 	IsCompleted bool   `json:"is_completed"` // 是否完成
// 	IsBreak     bool   `json:"is_break"`     // 是否中断
// 	MessageID   string `json:"message_id"`   // 消息ID
// 	SessionID   string `json:"session_id"`   // 会话ID
// }

// // StreamManager 流状态管理器
// type StreamManager struct {
// 	Streams map[string]*StreamState // sessionID_messageID -> StreamState
// 	Mu      sync.RWMutex
// }

// var GlobalStreamManager = StreamManager{
// 	Streams: make(map[string]*StreamState),
// }

// // 获取或创建流状态
// func (sm *StreamManager) GetOrCreateStream(sessionID, messageID, query string, resume bool) *StreamState {
// 	// 清理过期流，自定义过期时间
// 	go sm.cleanupExpiredStreams()
// 	// 若是恢复流，直接返回
// 	if resume {
// 		return sm.Streams[sessionID+"_"+messageID]
// 	}
// 	sm.Mu.Lock()
// 	defer sm.Mu.Unlock()

// 	// 说明是创建新流
// 	if messageID == "" {
// 		messageID = uuid.New().String()
// 	}
// 	// 创建新流
// 	stream := &StreamState{
// 		SessionID:    sessionID,
// 		MessageID:    messageID,
// 		Query:        query,
// 		FullResponse: "",
// 		Chunks:       []string{},
// 		IsCompleted:  false,
// 		CreatedAt:    time.Now(),
// 		UpdatedAt:    time.Now(),
// 		Clients:      make(map[string]chan StreamChunk),
// 	}

// 	sm.Streams[sessionID+"_"+messageID] = stream

// 	return stream
// }

// // 添加chunk到流
// func (sm *StreamManager) AddChunk(stream *StreamState, chunk string) {
// 	sm.Mu.Lock()
// 	defer sm.Mu.Unlock()

// 	if stream, exists := sm.Streams[stream.SessionID+"_"+stream.MessageID]; exists {
// 		stream.Chunks = append(stream.Chunks, chunk)
// 		stream.FullResponse += chunk
// 		stream.UpdatedAt = time.Now()
// 	}
// }

// // 标记流完成
// func (sm *StreamManager) CompleteStream(streamKey string) {
// 	sm.Mu.Lock()
// 	defer sm.Mu.Unlock()

// 	if stream, exists := sm.Streams[streamKey]; exists {
// 		stream.IsCompleted = true
// 		stream.UpdatedAt = time.Now()

// 		// 通知所有客户端完成
// 		for clientID, ch := range stream.Clients {
// 			ch <- StreamChunk{IsCompleted: true}
// 			close(ch)
// 			delete(stream.Clients, clientID)
// 		}
// 		//删除流
// 		delete(sm.Streams, streamKey)

// 	}

// }

// // 客户端注册监听
// func (sm *StreamManager) RegisterClient(stream *StreamState, clientID string) <-chan StreamChunk {
// 	ch := stream.Clients[clientID]
// 	if ch == nil {
// 		ch = make(chan StreamChunk, 100)
// 		stream.Clients[clientID] = ch
// 	}

// 	// 如果流已完成，立即发送完成信号
// 	if stream.IsCompleted {
// 		go func() {
// 			ch <- StreamChunk{IsCompleted: true}
// 			close(ch)
// 		}()
// 		return nil
// 	}
// 	return ch

// }

// // 客户端注销
// func (sm *StreamManager) UnregisterClient(sessionID, clientID string) {
// 	sm.Mu.Lock()
// 	defer sm.Mu.Unlock()

// 	if stream, exists := sm.Streams[sessionID]; exists {
// 		if ch, exists := stream.Clients[clientID]; exists {
// 			close(ch)
// 			delete(stream.Clients, clientID)
// 		}
// 	}
// }

// // 获取未完成的chunks
// func (sm *StreamManager) GetPendingChunks(messageID string) []StreamChunk {
// 	sm.Mu.RLock()
// 	defer sm.Mu.RUnlock()

// 	if stream, exists := sm.Streams[messageID]; exists && !stream.IsCompleted {
// 		var pending []StreamChunk
// 		for i := 0; i < len(stream.Chunks); i++ {
// 			pending = append(pending, StreamChunk{
// 				ChunkID:     i,
// 				Content:     stream.Chunks[i],
// 				IsCompleted: false,
// 			})
// 		}
// 		log.Printf("GetPendingChunks: %s, %d, stream:%v", messageID, len(pending), stream)

// 		return pending
// 	}

// 	return nil
// }

// // 清理过期流，自定义过期时间
// func (sm *StreamManager) cleanupExpiredStreams() {
// 	sm.Mu.Lock()
// 	defer sm.Mu.Unlock()

// 	cutoff := time.Now().Add(-10 * time.Minute)
// 	for streamKey, stream := range sm.Streams {
// 		if stream.UpdatedAt.Before(cutoff) {
// 			for clientID, ch := range stream.Clients {
// 				close(ch)
// 				delete(stream.Clients, clientID)
// 			}
// 			delete(sm.Streams, streamKey)
// 		}
// 	}
// }

// // BreakStream 中断流，通知所有客户端中断，关闭流
// // 返回是否存在该流，是否成功中断
// func (sm *StreamManager) BreakStream(sessionID, messageID string) (bool, error) {
// 	sm.Mu.Lock()
// 	defer sm.Mu.Unlock()

// 	if stream, exists := sm.Streams[sessionID+"_"+messageID]; exists {
// 		stream.IsBreak = true
// 		stream.UpdatedAt = time.Now()

// 		// 通知所有客户端中断
// 		for clientID, ch := range stream.Clients {
// 			ch <- StreamChunk{IsBreak: true}
// 			close(ch)
// 			delete(stream.Clients, clientID)
// 		}
// 		//删除流
// 		delete(sm.Streams, sessionID+"_"+messageID)

// 		//消息入库
// 		assistantMsg := my_models.Message{
// 			SessionID: stream.SessionID,
// 			ID:        stream.MessageID, // 使用相同的messageID
// 			Role:      "assistant",
// 			Content:   stream.FullResponse,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 			Metadata: map[string]any{
// 				"model": "mock_llm",
// 			},
// 		}
// 		my_service.My_dbservice.DB.Create(&assistantMsg)

// 		return true, nil
// 	}
// 	return false, errors.New("stream not found")
// }
