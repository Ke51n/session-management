package handler

import (
	"net/http"

	"session-demo/pkg/auth"
	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

// 创建一个项目，指定标题
func CreateProjectHandler(req *restful.Request, resp *restful.Response) {
	// 1. 解析参数
	reqBody, err := service.BindRequestBody[requests.CreateAndUpdateProjectReq](req)
	// 2. 统一处理解析错误 (Handler 负责 HTTP 响应)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	// 调用服务层
	project, err := service.CreateProject(reqBody, auth.GetUserID(req))
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	response.WriteSuccess(resp, http.StatusCreated, response.CreateProjectResponse{ProjectID: project.ID})
}

// 更新项目
func UpdateProjectHandler(req *restful.Request, resp *restful.Response) {
	// 1. 解析参数
	reqBody, err := service.BindRequestBody[requests.CreateAndUpdateProjectReq](req)
	// 2. 统一处理解析错误 (Handler 负责 HTTP 响应)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	projectID := req.PathParameter("projectId")
	userID := auth.GetUserID(req)

	// 调用服务层
	if _, err := service.UpdateProject(reqBody, projectID, userID); err != nil {
		response.WriteBizError(resp, err)
		return
	}

	response.WriteSuccess(resp, http.StatusOK, response.CreateOrEditProjectResponse{ProjectID: projectID})
}

// 删除一个项目
func DeleteProjectHandler(req *restful.Request, resp *restful.Response) {
	projectID := req.PathParameter("projectId")
	userID := auth.GetUserID(req)

	// 调用服务层
	if err := service.DeleteProject(projectID, userID); err != nil {
		response.WriteBizError(resp, err)
		return
	}
	response.WriteSuccess(resp, http.StatusOK, nil)
}

// 查询所有项目
func ListProjectsHandler(req *restful.Request, resp *restful.Response) {
	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)

	// 调用服务层
	projects, err := service.ListProjects(userID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	// 构造响应
	response.WriteSuccess(resp, http.StatusOK, projects)
}

// 查询某个项目下的所有会话
func ListProjectSessionsHandler(req *restful.Request, resp *restful.Response) {

	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)
	projectID := req.PathParameter("projectId")

	// 调用服务层
	sessions, err := service.ListSessionsInProject(userID, projectID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	// 构造响应
	response.WriteSuccess(resp, http.StatusOK, sessions)
}
