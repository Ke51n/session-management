package requests

//
// ========== 请求结构 ==========
//

type CreateProjectReq struct {
	Title  string `json:"title"`
	UserID string `json:"user_id"`
}

type CreateSessionReq struct {
	ProjectID string `json:"project_id"`
	Prompt    string `json:"prompt"`
}

type ChatReq struct {
	Prompt string `json:"prompt"`
}
