package response

import "session-demo/models"

// 查询某个项目下的所有会话响应结构
type ListSessionsResponse struct {
	Data    []models.Session `json:"data"`
	Success bool             `json:"success"`
}

// 创建项目响应结构
type CreateOrEditProjectResponse struct {
	Success   bool   `json:"success"`
	ProjectID string `json:"project_id"`
}

// 查询所有项目响应结构
type ListProjectsResponse struct {
	Data    []models.Project `json:"data"`
	Success bool             `json:"success"`
}

// 更新项目标题响应结构
type UpdateProjectResponse struct {
	Success bool `json:"success"`
}

// 删除项目响应结构
type DeleteProjectResponse struct {
	Success bool `json:"success"`
}
type ListMessagesResponse struct {
	Data    []models.Message `json:"data"`
	Success bool             `json:"success"`
}

// MoveSessionToProjectResponse 移动会话到项目响应结构
type MoveSessionToProjectResponse struct {
	Success bool `json:"success"`
}

type UpdateSessionTitleResponse struct {
	Success bool `json:"success"`
}

// BreakStreamChatResponse 中断流式对话响应结构
type BreakStreamChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
