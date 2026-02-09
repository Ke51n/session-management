package handler

import (
	"log"
	"net/http"
	my_models "session-demo/models"
	my_requests "session-demo/requests"
	my_response "session-demo/response"
	my_service "session-demo/service"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
)

// 创建一个会话并对话，sse流式响应CreateSessionStreamChatHandler
func CreateSessionStreamChatHandler(req *restful.Request, resp *restful.Response) {

	// 1. 设置 SSE Header
	writer := resp.ResponseWriter

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	// 从header解析TOKEN（可选）
	// 示例：从请求头中获取用户ID（假设有认证中间件设置）
	token := req.HeaderParameter("TOKEN")
	req.SetAttribute("user_id", token)

	v := req.Attribute("user_id")
	userID, ok := v.(string)
	if !ok || userID == "" {
		log.Println("user_id is missing or invalid, userID:", userID)
		resp.WriteErrorString(400, "uid is required")
		return
	}

	// 2. 解析请求体
	var reqData my_requests.CreateSessionAndChatReq
	if err := req.ReadEntity(&reqData); err != nil {
		http.Error(writer, err.Error(), 400)
		return
	}
	log.Printf("Received CreateSessionAndChatReq: %+v, userID: %s\n", reqData, userID)

	// 2.0 可选：如果请求中包含 project_id，验证项目是否存在且属于用户
	if reqData.ProjectID != "" {
		var project my_models.Project
		if my_service.My_dbservice.DB.Where("id = ? AND user_id = ? AND deleted = ?",
			reqData.ProjectID, userID, false).First(&project).Error != nil {
			http.Error(writer, "project not found", 404)
			return
		}
	}

	// 2.1 获取标题
	title := getTitleFromQuery(reqData.Query)
	// 3. 创建 Session（存数据库）
	session, err := my_service.CreateSession(userID, reqData.ProjectID, title)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	// 用户消息写数据库
	userMsgID := uuid.NewString()
	my_service.CreateAndSaveMessage(userMsgID, session.ID, nil, "user", nil, reqData.Files, reqData.Query,
		len(reqData.Query), false, nil, nil)
	log.Println("Saved user message id:", userMsgID)

	// 4. SSE 发送 session_id
	my_service.SendSSE(writer, flusher, "session", session.ID)

	// 5. 模拟 AI 流式回复
	reply := "今天上天气多云，15°C，适合外出游玩！记得带上防晒用品哦！如果你有其他问题，随时告诉我。"

	for _, ch := range reply {

		my_service.SendSSE(writer, flusher, "message", string(ch))

		time.Sleep(50 * time.Millisecond)
	}

	// 6. 结束标记
	my_service.SendSSE(writer, flusher, "done", "true")

	// 7. 关闭连接
	log.Println("Completed CreateSessionStreamChatHandler for userID:", userID)

	//模型消息写数据库
	assistantMsgID := uuid.NewString()
	steps := []my_models.StepNode{
		{
			Type: "thought",
			Text: "我需要先思考一下。",
		},
		{
			Type: "plan",
			Text: "我计划先检索相关信息，然后生成回答。",
		},
		{
			Type: "tool_call",
			Name: "get_weather",
			Text: "{'id':'tool_call_1','tool':'get_weather','args':{'location':'Beijing'}}",
		},
		{
			Type: "tool_return",
			Name: "get_weather",
			Text: "天气信息：上海当前温度为15度，多云。",
		},
		{
			Type: "thought",
			Text: "我已经有了足够的信息，可以生成回答了。",
		},
	}
	my_service.CreateAndSaveMessage(assistantMsgID, session.ID, &userMsgID, "assistant", steps, nil, reply, len(reply), false, nil, nil)
	log.Println("Saved assistant message id:", assistantMsgID)

}

// 从查询内容生成标题（简单截取前20字符）可以用更复杂的逻辑，比如llm总结，go协程异步，拿到后sse广播给所有客户端
func getTitleFromQuery(query string) string {
	runes := []rune(query) // 转为 rune 切片（每个元素是一个 Unicode 字符）
	if len(runes) > 20 {
		return string(runes[:20])
	}
	return query
}

// 查询某个会话的所有消息
func ListMessagesBySessionHandler(req *restful.Request, resp *restful.Response) {

	// 示例：从请求头中获取用户ID（假设有认证中间件设置）
	token := req.HeaderParameter("TOKEN")
	log.Println("extracted user_id from header:", token)
	req.SetAttribute("user_id", token)

	sessionID := req.PathParameter("sessionId")
	v := req.Attribute("user_id")
	userID, ok := v.(string)
	if !ok || userID == "" {
		log.Println("user_id is missing or invalid, userID:", userID)
		resp.WriteErrorString(400, "uid is required")
		return
	}

	// 调用服务层
	messages, err := my_service.ListMessagesBySession(userID, sessionID)
	if err != nil {
		log.Println("failed to list messages:", err)
		resp.WriteErrorString(http.StatusInternalServerError, "failed to list messages")
		return
	}

	// 构造响应
	result := my_response.ListMessagesResponse{
		Data:  messages,
		Total: len(messages),
	}

	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

func ListSessionsNotInProjectHandler(req *restful.Request, resp *restful.Response) {

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
	sessions, err := my_service.ListSessionsNotInProject(userID)
	if err != nil {
		log.Println("failed to list sessions not in project:", err)
		resp.WriteErrorString(http.StatusInternalServerError, "failed to list sessions not in project")
		return
	}

	// 构造响应
	result := my_response.ListSessionsResponse{
		Data:  sessions,
		Total: len(sessions),
	}

	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

// / MoveSessionToProjectHandler 移动一个会话到某个指定项目
func MoveSessionToProjectHandler(req *restful.Request, resp *restful.Response) {

	// 从请求头中获取用户ID
	userID := getUserIdFromHeader(req, resp)
	if userID == "" {
		return
	}

	sessionID := req.PathParameter("sessionId")
	var reqData my_requests.MoveSessionToProjectReq
	if err := req.ReadEntity(&reqData); err != nil {
		log.Println("failed to read request body:", err)
		resp.WriteErrorString(400, "invalid request body")
		return
	}

	// 判断项目和用户是否存在
	if reqData.ProjectID != "" {
		_, err := my_service.GetProjectById(userID, reqData.ProjectID)
		if err != nil {
			log.Println("project not found or user not authorized:", err)
			resp.WriteErrorString(http.StatusNotFound, "project not found or user not authorized")
			return
		}
	}

	// 调用服务层
	err := my_service.MoveSessionToProject(userID, sessionID, reqData.ProjectID)
	if err != nil {
		log.Println("failed to move session to project:", err)
		resp.WriteErrorString(http.StatusInternalServerError, "failed to move session to project")
		return
	}

	// 构造响应
	result := my_response.MoveSessionToProjectResponse{
		Success:   true,
		Message:   "session moved to project successfully",
		ProjectID: reqData.ProjectID,
		SessionID: sessionID,
	}

	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

func UpdateSessionHandler(req *restful.Request, resp *restful.Response) {

	// 从请求头中获取用户ID
	userID := getUserIdFromHeader(req, resp)
	if userID == "" {
		return
	}

	sessionID := req.PathParameter("sessionId")
	var reqData my_requests.UpdateSessionReq
	if err := req.ReadEntity(&reqData); err != nil {
		log.Println("failed to read request body:", err)
		resp.WriteErrorString(400, "invalid request body")
		return
	}

	// 调用服务层
	err := my_service.UpdateSession(userID, sessionID, reqData.Title)
	if err != nil {
		log.Println("failed to update session:", err)
		resp.WriteErrorString(http.StatusInternalServerError, "failed to update session")
		return
	}

	// 构造响应
	result := my_response.UpdateSessionResponse{
		Success:   true,
		SessionId: sessionID,
	}

	resp.WriteHeaderAndEntity(http.StatusOK, result)
}
