// stream_manager.go
package main

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StreamState 流式生成状态
type StreamState struct {
	MessageID    string                      `json:"message_id"`    // 消息ID
	SessionID    string                      `json:"session_id"`    // 会话ID
	Query        string                      `json:"query"`         // 用户查询
	FullResponse string                      `json:"full_response"` // 完整响应（逐步构建）
	Chunks       []string                    `json:"chunks"`        // 所有chunk
	SentIndex    int                         `json:"sent_index"`    // 已发送到的索引
	IsCompleted  bool                        `json:"is_completed"`  // 是否完成
	CreatedAt    time.Time                   `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time                   `json:"updated_at"`    // 更新时间
	Clients      map[string]chan StreamChunk // 连接的客户端
	mu           sync.RWMutex                `json:"-"` // 添加互斥锁

}

// StreamChunk 流式chunk
type StreamChunk struct {
	ChunkID    int    `json:"chunk_id"`
	Content    string `json:"content"`
	IsFinal    bool   `json:"is_final"`
	IsContinue bool   `json:"is_continue"` // 是否是继续之前的中断
}

// StreamManager 流状态管理器
type StreamManager struct {
	streams map[string]*StreamState // messageID -> StreamState
	mu      sync.RWMutex
}

var streamManager = &StreamManager{
	streams: make(map[string]*StreamState),
}

// 获取或创建流状态
func (sm *StreamManager) GetOrCreateStream(sessionID, query string) *StreamState {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 查找该会话未完成的流
	for _, stream := range sm.streams {
		if stream.SessionID == sessionID && !stream.IsCompleted {
			return stream
		}
	}

	// 创建新流
	messageID := uuid.NewString()
	stream := &StreamState{
		MessageID:    messageID,
		SessionID:    sessionID,
		Query:        query,
		FullResponse: "",
		Chunks:       []string{},
		SentIndex:    -1,
		IsCompleted:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Clients:      make(map[string]chan StreamChunk),
	}

	sm.streams[messageID] = stream

	// 清理过期流（24小时）
	go sm.cleanupExpiredStreams()

	return stream
}

// 添加chunk到流
func (sm *StreamManager) AddChunk(messageID string, chunk string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if stream, exists := sm.streams[messageID]; exists {
		stream.Chunks = append(stream.Chunks, chunk)
		stream.FullResponse += chunk
		stream.UpdatedAt = time.Now()
	}
}

// 标记流完成
func (sm *StreamManager) CompleteStream(messageID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if stream, exists := sm.streams[messageID]; exists {
		stream.IsCompleted = true
		stream.UpdatedAt = time.Now()

		// 通知所有客户端完成
		for clientID, ch := range stream.Clients {
			ch <- StreamChunk{IsFinal: true}
			close(ch)
			delete(stream.Clients, clientID)
		}
	}
}

// 客户端注册监听
func (sm *StreamManager) RegisterClient(messageID, clientID string) <-chan StreamChunk {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if stream, exists := sm.streams[messageID]; exists {
		ch := make(chan StreamChunk, 100)
		stream.Clients[clientID] = ch

		// 如果流已完成，立即发送完成信号
		if stream.IsCompleted {
			go func() {
				ch <- StreamChunk{IsFinal: true}
				close(ch)
			}()
		}

		return ch
	}

	return nil
}

// 客户端注销
func (sm *StreamManager) UnregisterClient(messageID, clientID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if stream, exists := sm.streams[messageID]; exists {
		if ch, exists := stream.Clients[clientID]; exists {
			close(ch)
			delete(stream.Clients, clientID)
		}
	}
}

// 获取未完成的chunks
func (sm *StreamManager) GetPendingChunks(messageID string) []StreamChunk {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if stream, exists := sm.streams[messageID]; exists && !stream.IsCompleted {
		var pending []StreamChunk
		for i := stream.SentIndex + 1; i < len(stream.Chunks); i++ {
			pending = append(pending, StreamChunk{
				ChunkID: i,
				Content: stream.Chunks[i],
				IsFinal: false,
			})
		}
		log.Printf("GetPendingChunks: %s, %d, stream:%v", messageID, len(pending), stream)

		return pending
	}

	return nil
}

// 清理过期流
func (sm *StreamManager) cleanupExpiredStreams() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for msgID, stream := range sm.streams {
		if stream.UpdatedAt.Before(cutoff) {
			for clientID, ch := range stream.Clients {
				close(ch)
				delete(stream.Clients, clientID)
			}
			delete(sm.streams, msgID)
		}
	}
}
