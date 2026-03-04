package auth

import (
	"log"
	"net/http"
	response "session-management/response"
	"time"

	"github.com/emicklei/go-restful/v3"
)

type (
	JWTClaim      = string
	JWTTokenScope = string
)

const (
	// CV 模型开发角色
	UserRoleAdmin = "admin"
	UserRoleGuest = "guest"
	// CV 模型应用角色
	UserRoleCVAdmin = "ROLE_ADMIN"
	UserRoleCVBasic = "ROLE_BASIC"
	// 兼容 SOPHON BASE 产品线角色
	UserRoleSophonAdmin = "SOPHON_ADMIN"
	UserRoleSophonBasic = "SOPHON_BASIC"

	DefaultUserName                     = "thinger"
	DefaultIss                          = "transwarp"
	DefaultSub                          = "llmops"
	InnerAuthCtx                        = "__auth_ctx__" // InnerAuthCtx 用于完成用户验证后，将用户信息存入 HTTP 请求上下文
	JWTTokenScopeInternal JWTTokenScope = "internal"
)

type JWToken struct {
	username  string
	roles     []string
	scope     string
	expiredAt int64 // unix second
	issuedAt  int64
	iss       string
	sub       string
}

var defaultToken = &JWToken{
	username:  DefaultUserName,
	roles:     []string{UserRoleSophonAdmin, UserRoleAdmin, UserRoleCVAdmin},
	scope:     JWTTokenScopeInternal,
	expiredAt: time.Now().Add(24 * time.Hour).Unix(),
	issuedAt:  time.Now().Unix(),
	iss:       DefaultIss,
	sub:       DefaultSub,
}

// 从请求头中获取用户ID
func getUserIdFromHeader(req *restful.Request) string {
	// 示例：从请求头中获取用户ID（假设有认证中间件设置）
	token := req.HeaderParameter("TOKEN")
	req.SetAttribute("user_id", token)
	log.Println("extracted user_id from header:", token)

	v := req.Attribute("user_id")
	userID, ok := v.(string)
	if !ok || userID == "" {
		log.Println("user_id is missing or invalid, userID=", userID)
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

	chain.ProcessFilter(req, resp)
}

// GetUserID 辅助函数，从 Context 获取用户ID
func GetUserID(req *restful.Request) string {
	val := req.Attribute("user_id")
	if val == nil {
		return ""
	}
	return val.(string)
}
