package service

import (
	"errors"
	"fmt"
	"log"
	my_models "session-demo/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 创建会话服务
func NewSessionService(db *gorm.DB) *DBService {
	return &DBService{DB: db}
}

// buildContextFromRootInMemory 从内存构建从 root 到 target 的上下文
func (s *DBService) buildContextFromRootInMemory(sessionID, targetMessageID string) ([]*my_models.Message, error) {
	// 加载会话所有消息到内存
	var allMessages []my_models.Message
	if err := s.DB.Where("session_id = ?", sessionID).Find(&allMessages).Error; err != nil {
		return nil, fmt.Errorf("load messages: %w", err)
	}

	// 构建内存消息映射
	msgMap := make(map[string]*my_models.Message, len(allMessages))
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
	path := []*my_models.Message{}
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
	var context []*my_models.Message
	for _, msg := range path {
		context = append(context, msg)
	}
	return context, nil
}

// CreateMessage 创建新消息
func (s *DBService) CreateMessage(sessionID string, parentID *string, role, content string) (*my_models.Message, error) {
	msg := &my_models.Message{
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
func (s *DBService) GetSessionMessages(convID string) ([]my_models.Message, error) {
	var msgs []my_models.Message
	err := s.DB.Where("session_id = ?", convID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

// Regenerate 重新生成回答
func (s *DBService) Regenerate(userID, sessionID, parentMessageID string) (*my_models.Message, error) {
	// 验证会话归属
	var conv my_models.Session
	if err := s.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return nil, errors.New("session not found or access denied")
	}

	// 获取父消息（必须是 user 消息）
	var parentMsg my_models.Message
	if err := s.DB.Where("id = ? AND session_id = ?", parentMessageID, sessionID).First(&parentMsg).Error; err != nil {
		return nil, errors.New("parent message not found")
	}
	if parentMsg.Role != "user" {
		return nil, errors.New("can only regenerate after user message")
	}

	// 查找是否已有 version_group_id，应该是有的，它们的version_group_id相同
	var existing my_models.Message
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
func (s *DBService) EditAndResend(userID, sessionID, targetMessageID, newContent string) (*my_models.Message, error) {
	var conv my_models.Session
	if err := s.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return nil, errors.New("session not found")
	}

	var targetMsg my_models.Message
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

// ListByProject 列出某个项目下的所有会话
func ListByProject(userID, projectID string) ([]my_models.Session, error) {
	// TODO: 查询数据库
	if projectID == "invalid" {
		return nil, errors.New("project not found")
	}
	//  查询数据库
	var sessions []my_models.Session
	log.Println("Listing sessions for userID:", userID, "projectID:", projectID)
	err := My_dbservice.DB.Where("project_id = ? AND user_id = ?", projectID, userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil

}

// 创建会话
func CreateSession(userID string, projectID *uint64, title string) (*my_models.Session, error) {
	session := &my_models.Session{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Deleted:   false,
	}
	if err := My_dbservice.DB.Create(session).Error; err != nil {
		return nil, err
	}
	log.Println("Created session:", session)
	return session, nil
}

// 这是服务层：建一个会话并对话，sse流式响应模型回答，要求流式响应，sse方式
