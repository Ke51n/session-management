package requests

// ========== 请求结构 ==========
//
// 创建项目请求结构
type CreateProjectReq struct {
	Title  string `json:"title"`
	UserID string `json:"user_id"`
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

// 更新项目标题请求结构
type UpdateProjectReq struct {
	Title string `json:"title"`
}

// 删除项目请求结构
type DeleteProjectReq struct {
	ProjectID string `json:"project_id"`
}
