package service

import "session-demo/models"

//创建一个项目
func (s *SessionService) CreateProject(title string, userID string) (*models.Project, error) {
	project := &models.Project{
		Title:  title,
		UserID: userID,
	}
	if err := s.DB.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

// import (
// 	"net/http"
// 	my_models "session-demo/models"
// 	my_requests "session-demo/requests"
// 	"time"
// )

// //
// // ========== Service 注册 ==========
// //

// func NewAIService(root string) *restful.WebService {

// 	ws := new(restful.WebService)

// 	ws.Path(root).
// 		Consumes(restful.MIME_JSON).
// 		Produces(restful.MIME_JSON)

// 	tags := []string{"AI 会话管理"}

// 	//
// 	// -------- Project --------
// 	//

// 	ws.Route(ws.POST("/projects").
// 		To(CreateProject).
// 		Doc("创建项目").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Reads(my_requests.CreateProjectReq{}).
// 		Writes(my_models.Project{}))

// 	ws.Route(ws.GET("/projects").
// 		To(ListProjects).
// 		Doc("查询项目列表").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Writes([]my_models.Project{}))

// 	ws.Route(ws.DELETE("/projects/{id}").
// 		To(DeleteProject).
// 		Doc("删除项目").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Param(ws.PathParameter("id", "项目ID")))

// 	//
// 	// -------- Session --------
// 	//

// 	ws.Route(ws.POST("/sessions").
// 		To(CreateSession).
// 		Doc("创建会话并对话").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Reads(my_requests.CreateSessionReq{}).
// 		Writes(my_models.Session{}))

// 	ws.Route(ws.GET("/sessions").
// 		To(ListSessions).
// 		Doc("查询会话").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Param(ws.QueryParameter("project_id", "项目ID").Required(false)).
// 		Writes([]my_models.Session{}))

// 	ws.Route(ws.POST("/sessions/{id}/chat").
// 		To(ChatInSession).
// 		Doc("会话中对话").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Param(ws.PathParameter("id", "会话ID")).
// 		Reads(my_requests.ChatReq{}))

// 	ws.Route(ws.DELETE("/sessions/{id}").
// 		To(DeleteSession).
// 		Doc("删除会话").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Param(ws.PathParameter("id", "会话ID")))

// 	//
// 	// -------- Message --------
// 	//

// 	ws.Route(ws.GET("/sessions/{id}/messages").
// 		To(ListMessages).
// 		Doc("查询消息历史").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Param(ws.PathParameter("id", "会话ID")).
// 		Writes([]my_models.Message{}))

// 	ws.Route(ws.DELETE("/messages/{id}").
// 		To(DeleteMessage).
// 		Doc("删除消息").
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Param(ws.PathParameter("id", "消息ID")))

// 	return ws
// }

// //
// // ========== Handler（Demo 用内存模拟） ==========
// //

// var (
// 	projects = make(map[string]my_models.Project)
// 	sessions = make(map[string]my_models.Session)
// 	messages = make(map[string][]my_models.Message)
// )

// func CreateProject(req *restful.Request, resp *restful.Response) {

// 	data := new(my_requests.CreateProjectReq)
// 	req.ReadEntity(data)

// 	id := genID()

// 	p := my_models.Project{
// 		ID:        id,
// 		Title:     data.Title,
// 		CreatedAt: time.Now(),
// 	}

// 	projects[id] = p

// 	resp.WriteEntity(p)
// }

// func ListProjects(req *restful.Request, resp *restful.Response) {

// 	list := make([]my_models.Project, 0)

// 	for _, p := range projects {
// 		list = append(list, p)
// 	}

// 	resp.WriteEntity(list)
// }

// func DeleteProject(req *restful.Request, resp *restful.Response) {

// 	id := req.PathParameter("id")

// 	delete(projects, id)

// 	resp.WriteHeader(http.StatusNoContent)
// }

// func CreateSession(req *restful.Request, resp *restful.Response) {

// 	data := new(my_requests.CreateSessionReq)
// 	req.ReadEntity(data)

// 	id := genID()

// 	s := my_models.Session{
// 		ID:        id,
// 		ProjectID: data.ProjectID,
// 		Title:     "New Chat",
// 		CreatedAt: time.Now(),
// 	}

// 	sessions[id] = s

// 	// 模拟第一条消息
// 	msg := my_models.Message{
// 		ID:        genID(),
// 		SessionID: id,
// 		Role:      "user",
// 		Content:   data.Prompt,
// 		CreatedAt: time.Now(),
// 	}

// 	messages[id] = append(messages[id], msg)

// 	resp.WriteEntity(s)
// }

// func ListSessions(req *restful.Request, resp *restful.Response) {

// 	projectID := req.QueryParameter("project_id")

// 	list := make([]my_models.Session, 0)

// 	for _, s := range sessions {

// 		if projectID != "" && s.ProjectID != projectID {
// 			continue
// 		}

// 		list = append(list, s)
// 	}

// 	resp.WriteEntity(list)
// }

// func ChatInSession(req *restful.Request, resp *restful.Response) {

// 	id := req.PathParameter("id")

// 	data := new(my_requests.ChatReq)
// 	req.ReadEntity(data)

// 	// user 消息
// 	userMsg := my_models.Message{
// 		ID:        genID(),
// 		SessionID: id,
// 		Role:      "user",
// 		Content:   data.Prompt,
// 		CreatedAt: time.Now(),
// 	}

// 	// assistant 模拟回复
// 	aiMsg := my_models.Message{
// 		ID:        genID(),
// 		SessionID: id,
// 		Role:      "assistant",
// 		Content:   "AI 回复: " + data.Prompt,
// 		CreatedAt: time.Now(),
// 	}

// 	messages[id] = append(messages[id], userMsg, aiMsg)

// 	resp.WriteEntity(aiMsg)
// }

// func DeleteSession(req *restful.Request, resp *restful.Response) {

// 	id := req.PathParameter("id")

// 	delete(sessions, id)
// 	delete(messages, id)

// 	resp.WriteHeader(http.StatusNoContent)
// }

// func ListMessages(req *restful.Request, resp *restful.Response) {

// 	id := req.PathParameter("id")

// 	resp.WriteEntity(messages[id])
// }

// func DeleteMessage(req *restful.Request, resp *restful.Response) {

// 	id := req.PathParameter("id")

// 	for sid, list := range messages {

// 		newList := make([]my_models.Message, 0)

// 		for _, m := range list {
// 			if m.ID != id {
// 				newList = append(newList, m)
// 			}
// 		}

// 		messages[sid] = newList
// 	}

// 	resp.WriteHeader(http.StatusNoContent)
// }

// //
// // ========== 工具函数 ==========
// //

// func genID() string {
// 	return time.Now().Format("20060102150405.000000")
// }
