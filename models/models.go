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
	ID        string         `gorm:"primaryKey"`
	UserID    string         `gorm:"not null;index"`
	Title     string         `gorm:"not null;default:'新项目'"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	Deleted   bool           `gorm:"not null;default:false"`    //是否删除
	Version   int64          `gorm:"not null;default:1"`        // 更新次数
	Extension map[string]any `gorm:"type:text;serializer:json"` // 扩展字段

}

// Session 会话表
type Session struct {
	ID        string         `gorm:"type:char(36);primaryKey"`
	ProjectID string         `gorm:"type:varchar(64);not null;default:'system'"` // 项目ID（关联项目）
	UserID    string         `gorm:"type:varchar(64);not null;index"`
	Title     string         `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	Source    string         `gorm:"type:varchar(32);not null"`
	Deleted   bool           `gorm:"not null;default:false"`
	Archived  bool           `gorm:"not null;default:false"`
	ShareLink *string        `gorm:"type:varchar(255)"`         // 分享链接（如果启用分享）
	Extension map[string]any `gorm:"type:text;serializer:json"` // 扩展字段
}

// Message 消息表
type Message struct {
	ID         string         `gorm:"type:char(36);primaryKey"`
	SessionID  string         `gorm:"type:varchar(64);not null;index"` // 所属会话ID
	ParentID   *string        `gorm:"index:idx_parent"`                //父消息id
	Role       string         `gorm:"type:varchar(20);not null"`       // "user" 或 "assistant"
	Files      []File         `gorm:"type:text;serializer:json"`       // 关联文件列表（JSON 字符串）
	ToolCalls  []ToolCall     `gorm:"type:text;serializer:json"`       // 助手发起的调用（可选）
	Content    string         `gorm:"type:longtext;not null"`
	TokenCount int            `gorm:"default:0"` // 建议添加：消息token数统计
	CreatedAt  time.Time      `gorm:"not null"`
	UpdatedAt  time.Time      `gorm:"not null"`
	Deleted    bool           `gorm:"not null;default:false"`    //是否删除
	Extension  map[string]any `gorm:"type:text;serializer:json"` // 扩展字段（存 JSON 字符串）
	Metadata   map[string]any `gorm:"type:text;serializer:json"` //其他信息
}

type File struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	URL       string         `json:"url"`
	Status    string         `json:"status"`
	Type      string         `json:"type"` // 文件类型（例如："image"、"document"）
	Extension map[string]any `gorm:"type:text;serializer:json"`
}

// ToolCall 助手调用工具的记录
type ToolCall struct {
	ID       string `json:"id"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
	Output string `json:"output"` // 工具执行结果（可选）
}

// BeforeCreate 钩子函数，在创建记录前调用
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	// 初始化 Extension 字段
	if s.Extension == nil {
		s.Extension = map[string]any{}
	}
	// 确保其他必填字段
	if s.Source == "" {
		s.Source = "user_create"
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
	if m.Extension == nil {
		m.Extension = map[string]any{}
	}
	// 初始化 Metadata 字段
	if m.Metadata == nil {
		m.Metadata = map[string]any{}
	}

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}

	return nil
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	// 初始化 Extension 字段
	if p.Extension == nil {
		p.Extension = map[string]any{}
	}
	// 确保其他必填字段
	if p.Title == "" {
		p.Title = "新项目"
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
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

func (Project) TableName() string {
	return "my_test_projects"
}
