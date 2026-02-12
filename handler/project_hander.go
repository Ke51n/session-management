package handler

import (
	"log"
	"net/http"

	constant "session-demo/const"
	"session-demo/pkg/auth"
	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

// 创建一个项目，指定标题
func CreateProjectHandler(req *restful.Request, resp *restful.Response) {
	var reqBody requests.CreateAndEditProjectReq
	if err := req.ReadEntity(&reqBody); err != nil {
		response.WriteBizError(resp, constant.ErrBadRequest)
		return
	}

	// 调用服务层
	project, err := service.CreateProject(&reqBody, auth.GetUserID(req))
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(http.StatusCreated, response.SuccessResp(project.ID))
}

// 更新项目
func UpdateProjectHandler(req *restful.Request, resp *restful.Response) {
	projectID := req.PathParameter("projectId")

	var reqBody requests.CreateAndEditProjectReq
	if err := req.ReadEntity(&reqBody); err != nil {
		response.WriteBizError(resp, constant.ErrBadRequest)
		return
	}

	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)

	// 调用服务层
	if _, err := service.UpdateProject(&reqBody, projectID, userID); err != nil {
		response.WriteBizError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, response.SuccessResp(nil))
}

// 删除一个项目
func DeleteProjectHandler(req *restful.Request, resp *restful.Response) {
	projectID := req.PathParameter("projectId")

	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)

	// 调用服务层
	if err := service.DeleteProject(projectID, userID); err != nil {
		response.WriteBizError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, response.SuccessResp(nil))
}

// 查询所有项目
func ListProjectsHandler(req *restful.Request, resp *restful.Response) {
	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)
	if userID == "" {
		response.WriteBizError(resp, constant.ErrUnauthorized)
		return
	}

	// 调用服务层
	projects, err := service.ListProjects(userID)
	if err != nil {
		response.WriteBizError(resp, err)
		return
	}

	// 构造响应
	resp.WriteHeaderAndEntity(http.StatusOK, response.SuccessResp(projects))
}

// 查询某个项目下的所有会话
func ListSessionsHandler(req *restful.Request, resp *restful.Response) {

	// 从请求头中获取用户ID
	userID := auth.GetUserID(req)
	projectIDStr := req.PathParameter("projectId")

	// 调用服务层
	sessions, err := service.ListByProject(userID, projectIDStr)
	if err != nil {
		// 根据错误类型返回不同状态码（示例简化）
		log.Println("failed to list sessions:", err)
		response.WriteBizError(resp, err)
		return
	}

	// 构造响应
	resp.WriteHeaderAndEntity(http.StatusOK, response.SuccessResp(sessions))
}
