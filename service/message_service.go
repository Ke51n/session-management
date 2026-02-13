package service

import (
	"net/http"
	my_models "session-demo/models"
	"session-demo/response"
)

// 保存一条消息到数据库
func CreateAndSaveMessage(msg *my_models.Message) error {
	if err := Dbservice.DB.Create(msg).Error; err != nil {
		return response.WrapError(500, "创建消息失败", err)
	}
	return nil
}

// 查询会话的所有消息
func ListMessagesBySession(userID, sessionID string) ([]my_models.Message, error) {
	var messages []my_models.Message
	err := Dbservice.DB.Where("session_id = ? AND deleted = ?", sessionID, false).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, response.WrapError(500, "查询消息失败", err)
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
		return nil, response.WrapError(http.StatusNotFound, "查询消息失败", err)
	}
	return &message, nil
}
