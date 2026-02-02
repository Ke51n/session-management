package handler

import (
	"net/http"

	"session-demo/requests"
	"session-demo/response"
	"session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

// 创建一个项目，指定标题
func CreateProjectHandler(service *service.SessionService) restful.RouteFunction {
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

// Handler 函数
func ListSessionsHandler(service *service.SessionService) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		projectID := req.PathParameter("projectId")
		userID := req.Attribute("user_id").(string)

		// 简单参数校验
		if projectID == "" {
			resp.WriteErrorString(http.StatusBadRequest, "projectId is required")
			return
		}

		// 调用服务层
		sessions, err := service.ListByProject(userID, projectID)
		if err != nil {
			// 根据错误类型返回不同状态码（示例简化）
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
