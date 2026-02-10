package requests

import (
	my_models "session-demo/models"
)

// ========== 请求结构 ==========
//
// 创建项目请求结构
type CreateAndEditProjectReq struct {
	Title string `json:"title"`
	// 自定义指令
	CustomInstruction string           `json:"custom_instruction"`
	Files             []my_models.File `json:"files"`
	// 工具配置
	ToolConfig my_models.JSONMap `gorm:"type:text;serializer:json" json:"tool_config"`
	// 模型服务配置
	ModelServiceConfig my_models.JSONMap `gorm:"type:text;serializer:json" json:"model_service_config"`
}

// 创建会话请求结构
type CreateSessionReq struct {
	ProjectID string `json:"project_id"`
	Prompt    string `json:"prompt"`
}

// 聊天请求结构
type ChatReq struct {
	Prompt string `json:"prompt"`
}

// 删除项目请求结构
type DeleteProjectReq struct {
	ProjectID string `json:"project_id"`
}

// 创建会话并对话
type CreateSessionAndChatReq struct {
	ProjectID string           `json:"project_id"`
	Query     string           `json:"query"`
	Files     []my_models.File `json:"files"`
}

// 移动会话到项目请求结构
type MoveSessionToProjectReq struct {
	ProjectID string `json:"project_id"`
}

// 更新会话标题请求结构
type UpdateSessionReq struct {
	Title string `json:"title"`
}

// 流式对话请求结构
type StreamChatReq struct {
	LastMsgID string         `json:"last_message_id"`
	QueryInfo QueryInfoModel `json:"query_info"`
}

// 查询信息模型
// 包含用户查询和相关文件
type QueryInfoModel struct {
	Query string           `json:"query"`
	Files []my_models.File `json:"files"`
}
