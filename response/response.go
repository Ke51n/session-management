package response

import "session-demo/models"

// 响应结构
type ListSessionsResponse struct {
	Data  []models.Session `json:"data"`
	Total int              `json:"total"`
}

type CreateProjectResponse struct {
	models.Project
}
type ListProjectsResponse struct {
	Data  []models.Project `json:"projects"`
	Total int              `json:"total"`
}
