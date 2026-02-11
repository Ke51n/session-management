package utils

import (
	"errors"
	"log"
	"net/http"

	my_models "session-demo/models"
	my_service "session-demo/service"

	"github.com/emicklei/go-restful/v3"
)

func IsAuth(sessionID string, req *restful.Request, resp *restful.Response) bool {
	if sessionID == "" {
		resp.WriteError(http.StatusBadRequest, errors.New("需要sessionId"))
		return false
	}
	// 从请求头中获取TOKEN
	userID := GetUserIdFromHeader(req, resp)
	if userID == "" {
		resp.WriteError(http.StatusUnauthorized, errors.New("未授权"))
		return false
	}

	// 验证会话归属
	var session *my_models.Session
	var err error
	if session, err = my_service.GetSessionById(sessionID); err != nil {
		resp.WriteError(http.StatusForbidden, err)
		return false
	} else if session.UserID != userID {
		resp.WriteError(http.StatusForbidden, errors.New("会话归属错误"))
		return false
	}
	return true
}

// 从请求头中获取用户ID
func GetUserIdFromHeader(req *restful.Request, resp *restful.Response) string {
	// 示例：从请求头中获取用户ID（假设有认证中间件设置）
	token := req.HeaderParameter("TOKEN")
	log.Println("extracted user_id from header:", token)
	req.SetAttribute("user_id", token)

	v := req.Attribute("user_id")
	userID, ok := v.(string)
	if !ok || userID == "" {
		log.Println("user_id is missing or invalid, userID:", userID)
		resp.WriteErrorString(400, "uid is required")
		return ""
	}
	return userID
}
