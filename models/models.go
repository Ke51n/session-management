package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

type JSONMap map[string]any // 避免重复定义

// Project 项目表
type Project struct {
	ID     string `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID string `gorm:"not null;index" json:"user_id"`
	Title  string `gorm:"not null;default:'新项目'" json:"title"`

	CustomInstruction string   `gorm:"type:longtext" json:"custom_instruction"` // 自定义指令
	Files             FileList `gorm:"type:longtext" json:"files"`              // 文件列表
	ToolsConfig       JSONMap  `gorm:"type:longtext" json:"tools_config"`       // 工具配置
	ModelSvcsConfig   JSONMap  `gorm:"type:longtext" json:"model_svcs_config"`  // 模型服务配置

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `gorm:"not null;default:false" json:"deleted"`      //是否删除
	Version   int64     `gorm:"not null;default:1" json:"version"`          // 更新次数
	Extension JSONMap   `gorm:"type:text;serializer:json" json:"extension"` // 扩展字段

}

// Session 会话表
type Session struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	ProjectID string    `gorm:"column:project_id;index" json:"project_id"` // 项目ID（关联项目）
	UserID    string    `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	Source    string    `gorm:"type:varchar(32);not null" json:"source"`    // 会话来源，如"user_create"
	Deleted   bool      `gorm:"not null;default:false" json:"deleted"`      // 是否删除
	Archived  bool      `gorm:"not null;default:false" json:"archived"`     // 是否归档
	ShareLink *string   `gorm:"type:varchar(255)" json:"share_link"`        //
	Extension JSONMap   `gorm:"type:text;serializer:json" json:"extension"` // 扩展字段（存 JSON 字符串）
}

// Message 消息表
type Message struct {
	ID        string  `gorm:"type:char(36);primaryKey" json:"id"`
	SessionID string  `gorm:"type:varchar(64);not null;index" json:"session_id"` // 所属会话ID
	ParentID  *string `gorm:"index:idx_parent" json:"parent_id"`                 //父消息id
	Role      string  `gorm:"type:varchar(20);not null" json:"role"`             // "user" 或 "assistant"
	Content   string  `gorm:"type:longtext;not null" json:"content"`
	Status    string  `gorm:"type:varchar(20);not null" json:"status"` // 消息状态，如"FINISHED"、"PROCESSING"、"INTERRUPTED"

	Steps StepList `gorm:"type:longtext" json:"steps"` // ← 关键：用自定义类型 + longtext
	Files FileList `gorm:"type:longtext" json:"files"` // 建议用 json 而非 text（PostgreSQL）或 longtext（MySQL）

	TokenCount int       `gorm:"default:0"` // 建议添加：消息token数统计
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`
	Deleted    bool      `gorm:"not null;default:false" json:"deleted"`      //是否删除
	Extension  JSONMap   `gorm:"type:text;serializer:json" json:"extension"` // 扩展字段（存 JSON 字符串）
	Metadata   JSONMap   `gorm:"type:text;serializer:json" json:"metadata"`  //其他信息
}

// StepNode 步骤节点，表示助手的思考、工具调用等
type StepNode struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"` // plan、tool_call、tool_return、thought 等
	Name     string  `json:"name"`
	Text     string  `json:"text"`
	ToolName string  `json:"tool_name"`
	ToolType string  `json:"tool_type"`
	Target   string  `json:"target"`
	Input    JSONMap `json:"input"`
	Output   string  `json:"output"`
	Metadata JSONMap `gorm:"type:text;serializer:json" json:"metadata"` //其他信息
}

type StepList []StepNode
type FileList []File

func (s FileList) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *FileList) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan FileList from unsupported type")
	}
}
func (s StepList) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *StepList) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan StepList from unsupported type")
	}
}

type File struct {
	ID                 string  `json:"id"`
	Uid                string  `json:"uid"`
	Name               string  `json:"name"`
	FileName           string  `json:"fileName"`
	URL                string  `json:"url"`
	Status             string  `json:"status"`
	Percent            uint8   `json:"percent"`
	Type               string  `json:"type"`               // 文件类型（例如："image"、"document"）
	Size               uint64  `json:"size"`               // 文件大小（字节）
	WebkitRelativePath string  `json:"webkitRelativePath"` // 文件在浏览器中的相对路径
	MergeData          JSONMap `json:"mergeData"`          // 合并数据（例如：合并后的图片路径）
	Extension          JSONMap `gorm:"type:text;serializer:json" json:"extension"`
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
		s.Extension = JSONMap{}
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
		m.Extension = JSONMap{}
	}
	// 初始化 Metadata 字段
	if m.Metadata == nil {
		m.Metadata = JSONMap{}
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
		p.Extension = JSONMap{}
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
