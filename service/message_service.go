package service

import (
	my_models "session-demo/models"
	"time"
)

// 保存一条消息到数据库
func CreateAndSaveMessage(messageID string, sessionID string, parentID *string, role string,
	steps []my_models.StepNode, files []my_models.File, content string, tokenCount int, deleted bool,
	extension, metadata map[string]any) error {

	msg := &my_models.Message{
		ID:        messageID,
		SessionID: sessionID,
		ParentID:  parentID,

		Role:       role,
		Steps:      steps,
		Files:      files,
		Content:    content,
		TokenCount: tokenCount,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Deleted:   deleted,

		Extension: extension,
		Metadata:  metadata,
	}
	if err := My_dbservice.DB.Create(msg).Error; err != nil {
		return err
	}
	return nil
}

// 查询会话的所有消息
func ListMessagesBySession(userID, sessionID string) ([]my_models.Message, error) {
	var messages []my_models.Message
	err := My_dbservice.DB.Where("session_id = ? AND deleted = ?", sessionID, false).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
