package service

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	constant "session-demo/const"
	"session-demo/models"
	"session-demo/response"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 创建会话服务
func NewSessionService(db *gorm.DB) *DBService {
	return &DBService{DB: db}
}

// buildContextFromRootInMemory 从内存构建从 root 到 target 的上下文
func (s *DBService) buildContextFromRootInMemory(sessionID, targetMessageID string) ([]*models.Message, error) {
	// 加载会话所有消息到内存
	var allMessages []models.Message
	if err := s.DB.Where("session_id = ?", sessionID).Find(&allMessages).Error; err != nil {
		return nil, fmt.Errorf("load messages: %w", err)
	}

	// 构建内存消息映射
	msgMap := make(map[string]*models.Message, len(allMessages))
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
	path := []*models.Message{}
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
	var context []*models.Message
	for _, msg := range path {
		context = append(context, msg)
	}
	return context, nil
}

// CreateMessage 创建新消息
func (s *DBService) CreateMessage(sessionID string, parentID *string, role, content string) (*models.Message, error) {
	msg := &models.Message{
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
func (s *DBService) GetSessionMessages(sessionId string) ([]models.Message, error) {
	var msgs []models.Message
	err := s.DB.Where("session_id = ?", sessionId).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

// Regenerate 重新生成回答
func (s *DBService) Regenerate(userID, sessionID, parentMessageID string) (*models.Message, error) {
	// 验证会话归属
	var conv models.Session
	if err := s.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return nil, errors.New("session not found or access denied")
	}

	// 获取父消息（必须是 user 消息）
	var parentMsg models.Message
	if err := s.DB.Where("id = ? AND session_id = ?", parentMessageID, sessionID).First(&parentMsg).Error; err != nil {
		return nil, errors.New("parent message not found")
	}
	if parentMsg.Role != constant.RoleUser {
		return nil, errors.New("can only regenerate after user message")
	}

	// 查找是否已有 version_group_id，应该是有的，它们的version_group_id相同
	var existing models.Message
	if err := s.DB.Where("parent_id = ? AND role = ?", parentMessageID, constant.RoleAssistant).First(&existing).Error; err != nil {
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
	newMsg, err := s.CreateMessage(sessionID, &parentMessageID, constant.RoleAssistant, modelResponse)
	if err != nil {
		return nil, err
	}

	return newMsg, nil
}

// EditAndResend 编辑并重发
func (s *DBService) EditAndResend(userID, sessionID, targetMessageID, newContent string) (*models.Message, error) {
	var conv models.Session
	if err := s.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return nil, errors.New("session not found")
	}

	var targetMsg models.Message
	if err := s.DB.Where("id = ? AND session_id = ?", targetMessageID, sessionID).First(&targetMsg).Error; err != nil {
		return nil, errors.New("target message not found")
	}
	if targetMsg.Role != constant.RoleUser {
		return nil, errors.New("can only edit user messages")
	}

	// 创建新 user 消息（继承原 parent）
	newUserMsg, err := s.CreateMessage(sessionID, targetMsg.ParentID, constant.RoleUser, newContent)
	if err != nil {
		return nil, err
	}

	// 构建上下文（到 newUserMsg 为止）
	context, err := s.buildContextFromRootInMemory(sessionID, newUserMsg.ID)
	if err != nil {
		return nil, err
	}
	//todo
	_ = context

	// TODO: 调用大模型
	modelResponse := "这是基于修改后问题的回答。"

	// 保存新 assistant 消息
	newAssistantMsg, err := s.CreateMessage(sessionID, &newUserMsg.ID, constant.RoleAssistant, modelResponse)
	if err != nil {
		return nil, err
	}

	return newAssistantMsg, nil
}

// ListSessionsInProject 列出某个项目下的所有会话
func ListSessionsInProject(userID string, projectID string) ([]models.Session, error) {
	if projectID == "" {
		return nil, &response.BizError{HttpStatus: http.StatusBadRequest, Code: 400, Msg: "项目ID不能为空"}
	}
	//  查询数据库
	var sessions []models.Session
	err := Dbservice.DB.Where("project_id = ? AND user_id = ?", projectID, userID).Find(&sessions).Error
	if err != nil {
		return nil, response.WrapError(404, "查询会话失败", err)
	}
	return sessions, nil

}

// ListAllSessions 列出用户的所有会话
func ListAllSessions(userID string) ([]models.Session, error) {
	// 查询数据库
	var sessions []models.Session
	err := Dbservice.DB.Where("user_id = ? AND deleted = ?", userID, false).Find(&sessions).Error
	if err != nil {
		return nil, response.WrapError(500, "查询会话失败", err)
	}
	return sessions, nil
}

// 创建会话
func CreateSession(userID string, projectID string, query string) (*models.Session, error) {
	title := genTitleFromQuery(query)
	session := &models.Session{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Deleted:   false,
	}
	if err := Dbservice.DB.Create(session).Error; err != nil {
		return nil, response.WrapError(500, "创建会话失败", err)
	}
	return session, nil
}

// MoveSessionToProject 移动会话到项目
func MoveSessionToProject(userID, sessionID string, projectID string) error {

	// 判断项目和用户是否存在
	if projectID != "" {
		_, err := GetProjectById(userID, projectID)
		if err != nil {
			return err
		}

	}

	// 验证会话归属
	var conv models.Session
	if err := Dbservice.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return response.WrapError(500, "查询会话失败", err)
	}

	// 更新会话项目ID
	conv.ProjectID = projectID
	conv.UpdatedAt = time.Now()
	if err := Dbservice.DB.Save(&conv).Error; err != nil {
		return response.WrapError(500, "报错会话失败", err)
	}
	return nil
}

// ListSessionsNotInProject 列出不在任何项目中的会话
func ListSessionsNotInProject(userID string) ([]models.Session, error) {
	// 查询数据库
	var sessions []models.Session
	err := Dbservice.DB.Where("project_id = '' AND user_id = ?", userID).Find(&sessions).Error
	if err != nil {
		return nil, response.WrapError(500, "查询会话失败", err)
	}
	return sessions, nil
}

// UpdateSession 更新会话标题
func UpdateSession(userID, sessionID string, title string) error {
	// 验证会话归属
	var conv models.Session
	if err := Dbservice.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return response.WrapError(500, "验证会话失败", err)
	}

	// 更新会话标题
	conv.Title = title
	conv.UpdatedAt = time.Now()
	if err := Dbservice.DB.Save(&conv).Error; err != nil {
		return response.WrapError(500, "更新会话失败", err)
	}
	return nil
}

// 根据id查询session
func QuerySession(userId, sessionID string) (*models.Session, *response.BizError) {
	// 验证会话归属
	var conv models.Session
	if err := Dbservice.DB.Where("id = ? AND user_id = ?", sessionID, userId).First(&conv).Error; err != nil {
		return nil, response.WrapError(500, "查询会话失败", err)
	}
	return &conv, nil
}

func GetSessionById(userId, sessionID string) (*models.Session, error) {
	// 验证会话归属
	var conv models.Session
	if err := Dbservice.DB.Where("id = ? AND user_id = ?", sessionID, userId).First(&conv).Error; err != nil {
		return nil, response.WrapError(404, "查询会话失败", err)
	}
	return &conv, nil
}

// 从查询内容生成标题（简单截取前20字符）可以用更复杂的逻辑，比如llm总结，go协程异步，拿到后sse广播给所有客户端
func genTitleFromQuery(query string) string {
	runes := []rune(query) // 转为 rune 切片（每个元素是一个 Unicode 字符）
	if len(runes) > 20 {
		return string(runes[:20])
	}
	return query
}

func DeleteSession(userID, sessionID string) error {
	// 验证会话归属
	var conv models.Session
	if err := Dbservice.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&conv).Error; err != nil {
		return response.WrapError(500, "查询会话失败", err)
	}

	// 删除会话
	conv.Deleted = true
	conv.UpdatedAt = time.Now()
	if err := Dbservice.DB.Save(&conv).Error; err != nil {
		return response.WrapError(500, "删除会话失败", err)
	}
	return nil
}
