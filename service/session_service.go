package service

import (
	"errors"
	"fmt"
	"log"
	model "session-demo/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 会话服务
type SessionService struct {
	DB *gorm.DB
}

// 创建会话服务
func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{DB: db}
}

// buildContextFromRootInMemory 从内存构建从 root 到 target 的上下文
func (s *SessionService) buildContextFromRootInMemory(sessionID, targetMessageID string) ([]*model.Message, error) {
	// 加载会话所有消息到内存
	var allMessages []model.Message
	if err := s.DB.Where("session_id = ?", sessionID).Find(&allMessages).Error; err != nil {
		return nil, fmt.Errorf("load messages: %w", err)
	}

	// 构建内存消息映射
	msgMap := make(map[string]*model.Message, len(allMessages))
	for i := range allMessages {
		msg := &allMessages[i]
		msgMap[msg.ID] = msg
	}

	// 查找目标消息
	targetMsg, ok := msgMap[targetMessageID]
	if !ok {
		return nil, errors.New("target message not found")
	}

	// 回溯路径
	path := []*model.Message{}
	current := targetMsg
	for current != nil {
		path = append(path, current)
		if current.ParentID == nil {
			break
		}
		parent, exists := msgMap[*current.ParentID]
		if !exists {
			return nil, fmt.Errorf("broken parent chain at %s", *current.ParentID)
		}
		current = parent
	}

	// 反转：root → target
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// 转为上下文
	var context []*model.Message
	for _, msg := range path {
		context = append(context, msg)
	}
	return context, nil
}

// CreateMessage 创建新消息
func (s *SessionService) CreateMessage(sessionID string, parentID *string, role, content string) (*model.Message, error) {
	msg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		ParentID:  parentID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.DB.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// GetSessionMessages 获取会话所有消息（用于前端渲染树）
func (s *SessionService) GetSessionMessages(convID string) ([]model.Message, error) {
	var msgs []model.Message
	err := s.DB.Where("session_id = ?", convID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

// Regenerate 重新生成回答
func (s *SessionService) Regenerate(userID, sessionID, parentMessageID string) (*model.Message, error) {
	// 验证会话归属
	var conv model.Session
	if err := s.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return nil, errors.New("session not found or access denied")
	}

	// 获取父消息（必须是 user 消息）
	var parentMsg model.Message
	if err := s.DB.Where("id = ? AND session_id = ?", parentMessageID, sessionID).First(&parentMsg).Error; err != nil {
		return nil, errors.New("parent message not found")
	}
	if parentMsg.Role != "user" {
		return nil, errors.New("can only regenerate after user message")
	}

	// 查找是否已有 version_group_id，应该是有的，它们的version_group_id相同
	var existing model.Message
	if err := s.DB.Where("parent_id = ? AND role = 'assistant'", parentMessageID).First(&existing).Error; err != nil {
		return nil, errors.New("no existing assistant message found")
	}

	// 构建上下文（到 parentMessageID 为止）
	context, err := s.buildContextFromRootInMemory(sessionID, parentMessageID)

	if err != nil {
		return nil, err
	}
	log.Println("重新生成消息，建立context:", context)

	// TODO: 调用大模型（此处模拟）
	modelResponse := "这是重新生成的回答。时间戳：" + time.Now().Format(time.RFC3339)

	// 保存新回答
	newMsg, err := s.CreateMessage(sessionID, &parentMessageID, "assistant", modelResponse)
	if err != nil {
		return nil, err
	}

	return newMsg, nil
}

// EditAndResend 编辑并重发
func (s *SessionService) EditAndResend(userID, sessionID, targetMessageID, newContent string) (*model.Message, error) {
	var conv model.Session
	if err := s.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return nil, errors.New("session not found")
	}

	var targetMsg model.Message
	if err := s.DB.Where("id = ? AND session_id = ?", targetMessageID, sessionID).First(&targetMsg).Error; err != nil {
		return nil, errors.New("target message not found")
	}
	if targetMsg.Role != "user" {
		return nil, errors.New("can only edit user messages")
	}

	// 创建新 user 消息（继承原 parent）
	newUserMsg, err := s.CreateMessage(sessionID, targetMsg.ParentID, "user", newContent)
	if err != nil {
		return nil, err
	}

	// 构建上下文（到 newUserMsg 为止）
	context, err := s.buildContextFromRootInMemory(sessionID, newUserMsg.ID)
	if err != nil {
		return nil, err
	}
	//TODO:
	_ = context

	// TODO: 调用大模型
	modelResponse := "这是基于修改后问题的回答。"

	// 保存新 assistant 消息
	newAssistantMsg, err := s.CreateMessage(sessionID, &newUserMsg.ID, "assistant", modelResponse)
	if err != nil {
		return nil, err
	}

	return newAssistantMsg, nil
}

func (s *SessionService) ListByProject(userID, projectID string) ([]model.Session, error) {
	// TODO: 查询数据库
	// 示例：SELECT * FROM sessions WHERE project_id = ? AND user_id = ?
	if projectID == "invalid" {
		return nil, errors.New("project not found")
	}
	// 模拟返回数据
	return []model.Session{
		{ID: "sess-1", Title: "需求讨论", CreatedAt: time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 1, 30, 10, 5, 0, 0, time.UTC)},
		{ID: "sess-2", Title: "技术方案", CreatedAt: time.Date(2026, 1, 29, 15, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 1, 29, 15, 10, 0, 0, time.UTC)},
	}, nil
}
