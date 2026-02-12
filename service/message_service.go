package service

import (
	"log"
	"net/http"
	my_models "session-demo/models"
	"session-demo/response"
	"time"
)

// 保存一条消息到数据库
func CreateAndSaveMessage(messageID string, sessionID string, parentID *string, role string,
	steps []my_models.StepNode, files []my_models.File, content string, tokenCount int, status string, deleted bool,
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
		Status:    status,
		Deleted:   deleted,

		Extension: extension,
		Metadata:  metadata,
	}
	if err := Dbservice.DB.Create(msg).Error; err != nil {
		return err
	}
	log.Printf("创建消息成功: %v", msg)
	return nil
}

// 查询会话的所有消息
func ListMessagesBySession(userID, sessionID string) ([]my_models.Message, error) {
	var messages []my_models.Message
	err := Dbservice.DB.Where("session_id = ? AND deleted = ?", sessionID, false).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// 更新消息状态，全量更新
func updateMessageById(message *my_models.Message) error {
	if err := Dbservice.DB.Save(message).Error; err != nil {
		return err
	}
	return nil
}

// 查询会话的一条消息
func GetMessageById(sessionID, messageID string) (*my_models.Message, error) {
	var message my_models.Message
	err := Dbservice.DB.Where("id = ? AND session_id = ?", messageID, sessionID).
		First(&message).Error
	if err != nil {
		return nil, &response.BizError{HttpStatus: http.StatusNotFound, Msg: err.Error()}
	}
	return &message, nil
}
