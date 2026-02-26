package service

import (
	"fmt"
	"net/http"
	constant "session-demo/const"
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

// 更新消息的特定字段
func updateMessageFields(messageId string, updates map[string]any) error {
	result := Dbservice.DB.Model(&my_models.Message{}).
		Where("id = ?", messageId).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	// 可选：检查是否找到记录
	if result.RowsAffected == 0 {
		return fmt.Errorf("message with id %s not found", messageId)
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

// 删除会话的一条消息，及后续消息
func DeleteMessage(sessionID, messageID string) error {
	//先查询会话的所有消息
	messages, err := ListMessagesBySession("", sessionID)
	if err != nil {
		return err
	}

	//建立消息的父子关系map，key是父消息id，value是子消息id集合
	messageRelationMap := make(map[string][]string)
	messageMap := make(map[string]my_models.Message)
	for _, msg := range messages {
		messageMap[msg.ID] = msg
		if msg.ParentID != nil {
			messageRelationMap[*msg.ParentID] = append(messageRelationMap[*msg.ParentID], msg.ID)
		}
	}

	//检查要删除的消息是否存在,且是用户消息
	msg, exists := messageMap[messageID]
	if !exists || msg.Deleted || msg.Role != constant.RoleUser {
		return constant.ErrInvalidMessageID
	}

	//根据父子关系找到所有要删除的消息
	var toDelete []string
	var curParentIDs = []string{messageID}
	for len(curParentIDs) > 0 {
		toDelete = append(toDelete, curParentIDs...)
		var nextParentIDs []string
		for _, parentID := range curParentIDs {
			if childrenIDs, exists := messageRelationMap[parentID]; exists {
				nextParentIDs = append(nextParentIDs, childrenIDs...)
			}
		}
		curParentIDs = nextParentIDs
	}

	// 删除索引及后续消息
	for _, msgID := range toDelete {
		if err := updateMessageFields(msgID, map[string]any{"deleted": true}); err != nil {
			return err
		}
	}

	return nil
}
