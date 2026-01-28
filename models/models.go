package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

type Project struct {
	ID        string    `gorm:"primaryKey"`
	UserID    string    `gorm:"not null;index"`
	Name      string    `gorm:"not null;default:'新项目'"`
	IsDefault bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
	Deleted   bool      `gorm:"not null;default:false"`
	Version   int64     `gorm:"not null;default:1"`    // 更新次数
	Extension string    `gorm:"type:json; column:ext"` // 扩展字段（存 JSON 字符串）

}

// Session 会话表
type Session struct {
	ID              string    `gorm:"type:char(36);primaryKey"`
	ParentID        *string   `gorm:"type:char(36)"`                              // 父会话ID（如果是子会话）
	ProjectID       string    `gorm:"type:varchar(64);not null;default:'system'"` // 项目ID（关联项目）
	ActiveTailMsgID *string   `gorm:"type:char(36); column:active_tail_msg_id"`   // 该会话活跃激活路径的尾消息ID
	UserID          string    `gorm:"type:varchar(64);not null;index"`
	Title           string    `gorm:"type:varchar(255);not null"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
	Source          string    `gorm:"type:varchar(32);not null"`
	Deleted         bool      `gorm:"not null;default:false"`
	Archived        bool      `gorm:"not null;default:false"`
	ShareLink       *string   `gorm:"type:varchar(255)"`     // 分享链接（如果启用分享）
	Version         int64     `gorm:"not null;default:1"`    // 会话的版本（更新次数）
	Extension       string    `gorm:"type:json; column:ext"` // 扩展字段（存 JSON 字符串）

}

// Message 消息表
type Message struct {
	ID             string                 `gorm:"type:char(36);primaryKey"`
	SessionID      string                 `gorm:"type:varchar(64);not null;index"` // 所属会话ID
	ParentID       *string                `gorm:"index:idx_parent"`
	Role           string                 `gorm:"type:varchar(20);not null"` // "user" 或 "assistant"
	Type           string                 `gorm:"type:varchar(20);not null"` // text、img、file等
	Content        string                 `gorm:"type:longtext;not null"`
	TokenCount     int                    `gorm:"default:0"`                                         // 建议添加：消息token数统计
	VersionGroupID string                 `gorm:"type:varchar(64);not null;index:idx_version_group"` // 必填！
	CreatedAt      time.Time              `gorm:"not null"`
	UpdatedAt      time.Time              `gorm:"not null"`
	Deleted        bool                   `gorm:"not null;default:false"`
	Extension      string                 `gorm:"type:json; column:ext"` // 扩展字段（存 JSON 字符串）
	Metadata       map[string]interface{} `gorm:"type:text;serializer:json"`
}

// Session 创建前的钩子
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	// 初始化 Extension 字段
	if s.Extension == "" {
		s.Extension = "{}"
	}

	// 确保其他必填字段
	if s.Source == "" {
		s.Source = "user_create"
	}
	if s.ProjectID == "" {
		s.ProjectID = "system"
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = time.Now()
	}

	return nil
}

// Message 创建前的钩子
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	// 初始化 Extension 字段
	if m.Extension == "" {
		m.Extension = "{}"
	}
	// 初始化 Metadata 字段
	if m.Metadata == nil {
		m.Metadata = map[string]interface{}{}
	}

	// 设置默认消息类型
	if m.Type == "" {
		m.Type = "text"
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}

	return nil
}

// TableName 自定义表名
func (Session) TableName() string {
	return "my_test_sessions"
}

func (Message) TableName() string {
	return "my_test_messages"
}
