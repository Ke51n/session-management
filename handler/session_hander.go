package handler

import (
	"log"
	"net/http"

	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

// 创建一个项目，指定标题
func CreateProjectHandler(service *service.DBService) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		var reqBody requests.CreateProjectReq
		if err := req.ReadEntity(&reqBody); err != nil {
			resp.WriteErrorString(http.StatusBadRequest, "invalid request body")
			return
		}
		// 调用服务层
		project, err := service.CreateProject(reqBody.Title, reqBody.UserID)
		if err != nil {
			resp.WriteErrorString(http.StatusInternalServerError, "failed to create project")
			return
		}
		// 构造响应
		result := response.CreateProjectResponse{
			Project: *project,
		}
		resp.WriteHeaderAndEntity(http.StatusCreated, result)
	}
}

// 查询所有项目
func ListProjectsHandler(service *service.DBService) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		// 示例：从请求头中获取用户ID（假设有认证中间件设置）
		token := req.HeaderParameter("TOKEN")
		log.Println("extracted user_id from header:", token)
		req.SetAttribute("user_id", token)

		v := req.Attribute("user_id")
		userID, ok := v.(string)
		if !ok || userID == "" {
			log.Println("user_id is missing or invalid, userID:", userID)
			resp.WriteErrorString(400, "uid is required")
			return
		}

		// 调用服务层
		projects, err := service.ListProjects(userID)
		if err != nil {
			log.Println("failed to list projects:", err)
			resp.WriteErrorString(http.StatusInternalServerError, "failed to list projects")
			return
		}

		// 构造响应
		result := response.ListProjectsResponse{
			Data:  projects,
			Total: len(projects),
		}

		resp.WriteHeaderAndEntity(http.StatusOK, result)
	}
}

// 查询某个项目下的所有会话
func ListSessionsHandler(service *service.DBService) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {

		// 示例：从请求头中获取用户ID（假设有认证中间件设置）
		token := req.HeaderParameter("TOKEN")
		log.Println("extracted user_id from header:", token)
		req.SetAttribute("user_id", token)

		projectID := req.PathParameter("projectId")
		v := req.Attribute("user_id")
		userID, ok := v.(string)
		if !ok || userID == "" {
			log.Println("user_id is missing or invalid, userID:", userID)
			resp.WriteErrorString(400, "uid is required")
			return
		}

		// 简单参数校验
		if projectID == "" {
			log.Println("projectId is missing")
			resp.WriteErrorString(http.StatusBadRequest, "projectId is required")
			return
		}

		// 调用服务层
		sessions, err := service.ListByProject(userID, projectID)
		if err != nil {
			// 根据错误类型返回不同状态码（示例简化）
			log.Println("failed to list sessions:", err)
			resp.WriteErrorString(http.StatusNotFound, "project not found")
			return
		}

		// 构造响应
		result := response.ListSessionsResponse{
			Data:  sessions,
			Total: len(sessions),
		}

		resp.WriteHeaderAndEntity(http.StatusOK, result)
	}
}
