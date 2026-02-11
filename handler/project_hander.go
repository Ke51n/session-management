package handler

import (
	"log"
	"net/http"

	"session-demo/requests"
	"session-demo/response"
	my_service "session-demo/service"
	my_utils "session-demo/utils"

	"github.com/emicklei/go-restful/v3"
)

// 创建一个项目，指定标题
func CreateProjectHandler() restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		var reqBody requests.CreateAndEditProjectReq
		if err := req.ReadEntity(&reqBody); err != nil {
			resp.WriteErrorString(http.StatusBadRequest, "invalid request body")
			return
		}

		// 从请求头中获取用户ID
		userID := my_utils.GetUserIdFromHeader(req, resp)
		if userID == "" {
			return
		}

		// 调用服务层
		project, err := my_service.CreateOrEditProject(reqBody.Title, reqBody.CustomInstruction, reqBody.Files, reqBody.ToolConfig, reqBody.ModelServiceConfig, userID)
		if err != nil {
			resp.WriteErrorString(http.StatusInternalServerError, "failed to create project")
			return
		}
		// 构造响应
		result := response.CreateOrEditProjectResponse{
			Success:   true,
			ProjectID: project.ID,
		}
		resp.WriteHeaderAndEntity(http.StatusCreated, result)
	}
}

// 更新项目标题
func UpdateProjectHandler() restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		projectID := req.PathParameter("projectId")
		var reqBody requests.CreateAndEditProjectReq
		if err := req.ReadEntity(&reqBody); err != nil {
			resp.WriteErrorString(http.StatusBadRequest, "invalid request body")
			return
		}

		// 从请求头中获取用户ID
		userID := my_utils.GetUserIdFromHeader(req, resp)
		if userID == "" {
			return
		}

		// 调用服务层
		_, err := my_service.UpdateProjectTitle(projectID, reqBody.Title, reqBody.CustomInstruction, reqBody.Files, reqBody.ToolConfig, reqBody.ModelServiceConfig, userID)
		if err != nil {
			resp.WriteErrorString(http.StatusInternalServerError, "failed to update project title")
			return
		}

		// 构造响应
		result := response.UpdateProjectResponse{
			Success: true,
		}
		resp.WriteHeaderAndEntity(http.StatusOK, result)
	}
}

// 删除一个项目
func DeleteProjectHandler() restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		projectID := req.PathParameter("projectId")

		// 从请求头中获取用户ID
		userID := my_utils.GetUserIdFromHeader(req, resp)
		if userID == "" {
			return
		}

		// 调用服务层
		err := my_service.DeleteProject(projectID, userID)
		if err != nil {
			resp.WriteErrorString(http.StatusInternalServerError, "failed to delete project")
			return
		}
		// 构造响应
		result := response.DeleteProjectResponse{
			Success: true,
		}
		resp.WriteHeaderAndEntity(http.StatusOK, result)
	}
}

// 查询所有项目
func ListProjectsHandler() restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		// 从请求头中获取用户ID
		userID := my_utils.GetUserIdFromHeader(req, resp)
		if userID == "" {
			return
		}

		// 调用服务层
		projects, err := my_service.ListProjects(userID)
		if err != nil {
			log.Println("failed to list projects:", err)
			resp.WriteErrorString(http.StatusInternalServerError, "failed to list projects")
			return
		}

		// 构造响应
		result := response.ListProjectsResponse{
			Data: projects,
		}

		resp.WriteHeaderAndEntity(http.StatusOK, result)
	}
}

// 查询某个项目下的所有会话
func ListSessionsHandler() restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {

		// 从请求头中获取用户ID
		userID := my_utils.GetUserIdFromHeader(req, resp)
		if userID == "" {
			return
		}
		//  2. 将字符串转换为 uint64
		// 这里的 10 表示十进制进制，64 表示 bit 大小
		projectIDStr := req.PathParameter("projectId")
		// 调用服务层
		sessions, err := my_service.ListByProject(userID, projectIDStr)
		if err != nil {
			// 根据错误类型返回不同状态码（示例简化）
			log.Println("failed to list sessions:", err)
			resp.WriteErrorString(http.StatusNotFound, "project not found")
			return
		}

		// 构造响应
		result := response.ListSessionsResponse{
			Data:    sessions,
			Success: true,
		}

		resp.WriteHeaderAndEntity(http.StatusOK, result)
	}
}
