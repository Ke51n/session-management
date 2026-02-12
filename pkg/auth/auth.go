package auth

import (
	"log"
	"net/http"

	response "session-demo/response"

	"github.com/emicklei/go-restful/v3"
)

// func IsAuth(sessionID string, req *restful.Request, resp *restful.Response) bool {
// 	if sessionID == "" {
// 		resp.WriteError(http.StatusBadRequest, errors.New("需要sessionId"))
// 		return false
// 	}
// 	// 从请求头中获取TOKEN
// 	userID := GetUserIdFromHeader(req, resp)
// 	if userID == "" {
// 		resp.WriteError(http.StatusUnauthorized, errors.New("未授权"))
// 		return false
// 	}

// 	// 验证会话归属
// 	var session *models.Session
// 	var err error
// 	if session, err = service.GetSessionById(sessionID); err != nil {
// 		resp.WriteError(http.StatusForbidden, err)
// 		return false
// 	} else if session.UserID != userID {
// 		resp.WriteError(http.StatusForbidden, errors.New("会话归属错误"))
// 		return false
// 	}
// 	return true
// }

// 从请求头中获取用户ID
func getUserIdFromHeader(req *restful.Request) string {
	// 示例：从请求头中获取用户ID（假设有认证中间件设置）
	token := req.HeaderParameter("TOKEN")
	log.Println("extracted user_id from header:", token)

	v := req.Attribute("user_id")
	userID, ok := v.(string)
	if !ok || userID == "" {
		log.Println("user_id is missing or invalid, userID:", userID)
		return ""
	}
	return userID
}

// AuthFilter 拦截所有需要认证的请求
func AuthFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	userID := getUserIdFromHeader(req) // 修改 GetUserIdFromHeader 不要传入 resp，只返回 string
	if userID == "" {
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, response.CommonResponse{
			Code:    401,
			Message: "Unauthorized",
		})
		return
	}

	// 将 userID 存入 Request 的 Attribute 中，供后续 Handler 使用
	req.SetAttribute("userID", userID)

	chain.ProcessFilter(req, resp)
}

// GetUserID 辅助函数，从 Context 获取用户ID
func GetUserID(req *restful.Request) string {
	val := req.Attribute("userID")
	if val == nil {
		return ""
	}
	return val.(string)
}
