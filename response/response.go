package response

import "session-demo/models"

// 查询某个项目下的所有会话响应结构
type ListSessionsResponse struct {
	Data  []models.Session `json:"data"`
	Total int              `json:"total"`
}

// 创建项目响应结构
type CreateProjectResponse struct {
	Success   bool   `json:"success"`
	ProjectID string `json:"project_id"`
}

// 查询所有项目响应结构
type ListProjectsResponse struct {
	Data  []models.Project `json:"projects"`
	Total int              `json:"total"`
}

// 更新项目标题响应结构
type UpdateProjectResponse struct {
	Success   bool   `json:"success"`
	ProjectID string `json:"project_id"`
}

// 删除项目响应结构
type DeleteProjectResponse struct {
	Success bool `json:"success"`
}
type ListMessagesResponse struct {
	Data  []models.Message `json:"data"`
	Total int              `json:"total"`
}

// MoveSessionToProjectResponse 移动会话到项目响应结构
type MoveSessionToProjectResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ProjectID string `json:"project_id"`
	SessionID string `json:"session_id"`
}

type UpdateSessionResponse struct {
	Success   bool   `json:"success"`
	SessionId string `json:"session_id"`
}
