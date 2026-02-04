package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"session-demo/models"
	my_models "session-demo/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"encoding/json"

	"gorm.io/gorm"

	my_handler "session-demo/handler"
	my_response "session-demo/response"

	"github.com/emicklei/go-restful/v3"
)

// 全局数据库实例
var GlobalDB *gorm.DB

// 模拟LLM响应函数
func mockLLMCall(prompt string) string {
	// 这里模拟LLM调用，实际项目中替换为真实的API调用
	log.Printf("调用LLM，Prompt长度: %d 字符", len(prompt))

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	// 简单的模拟回复
	responses := []string{
		"这是一个基于您历史对话生成的模拟回复。",
		"根据您的上下文，我认为可以这样继续讨论。",
		"从之前的对话来看，这个问题的解决方案可能是...",
		"结合我们之前的交流，我的建议是...",
		"基于历史上下文，我理解您想了解的是...",
	}

	// // 根据prompt内容返回不同的回复
	// if strings.Contains(prompt, "你好") || strings.Contains(prompt, "hello") {
	// 	return "你好！很高兴继续我们的对话。有什么我可以帮助您的吗？"
	// }
	// if strings.Contains(prompt, "天气") {
	// 	return "根据我们之前的对话，您似乎对天气比较关心。今天天气不错，适合外出。"
	// }

	// 默认回复
	return responses[int(time.Now().Unix())%len(responses)]
}

// // 初始化数据库
// func initDB() {

// 	dsn := "gormuser:gorm123@tcp(127.0.0.1:3306)/gorm_test?charset=utf8mb4&parseTime=True&loc=Local"
// 	var err error
// 	GlobalDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

// 	// dsn := "gormuser:gorm123@tcp(127.0.0.1:3306)/gorm_test?charset=utf8mb4&parseTime=True&loc=Local"

// 	// var err error
// 	// 使用MySQL数据库（便于演示，生产环境请用MySQL/PostgreSQL）
// 	// db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("数据库连接失败:", err)
// 	}

// 	// 自动迁移表结构
// 	err = GlobalDB.AutoMigrate(&models.Session{}, &models.Message{}, &models.Project{})
// 	if err != nil {
// 		log.Fatal("数据库迁移失败:", err)
// 	}

// 	// 创建一些测试数据
// 	// createTestData()

// 	log.Println("数据库初始化完成")
// }

// 创建测试数据
// func createTestData() {
// 	sessionID := uuid.NewString()

// 	// 创建测试会话
// 	session := models.Session{
// 		ID:        sessionID,
// 		ProjectID: "demo-project",
// 		UserID:    "test-user",
// 		Title:     "测试对话",
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 		Source:    "user_create",
// 	}
// 	GlobalDB.Create(&session)
// 	log.Printf("创建会话: %s", sessionID)

// 	// 创建历史消息
// 	messages := []models.Message{
// 		{
// 			ID:        uuid.NewString(),
// 			SessionID: sessionID,
// 			Role:      "user",
// 			Content:   "你好，我想了解天气情况",
// 			CreatedAt: time.Now().Add(-2 * time.Hour),
// 			UpdatedAt: time.Now().Add(-2 * time.Hour),
// 		},
// 		{
// 			ID:        uuid.NewString(),
// 			SessionID: sessionID,
// 			Role:      "assistant",
// 			Content:   "今天天气晴朗，气温25度，适合外出活动。",
// 			CreatedAt: time.Now().Add(-1 * time.Hour),
// 			UpdatedAt: time.Now().Add(-1 * time.Hour),
// 		},
// 		{
// 			ID:        uuid.NewString(),
// 			SessionID: sessionID,
// 			Role:      "user",
// 			Content:   "那明天呢？",
// 			CreatedAt: time.Now().Add(-30 * time.Minute),
// 			UpdatedAt: time.Now().Add(-30 * time.Minute),
// 		},
// 	}

// 	// for _, msg := range messages {
// 	// 	db.Create(&msg)
// 	// }
// 	// ⚡ 批量插入（一条 SQL）
// 	result := GlobalDB.Create(&messages)
// 	if result.Error != nil {
// 		panic(result.Error)
// 	}
// 	log.Printf("成功插入 %d 条消息记录\n", result.RowsAffected)
// }

// 新增 SSE 事件类型
// 删除原来的 DialogResponse，因为 SSE 会分块发送
type SSEMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// 请求和响应结构体
type DialogRequest struct {
	SessionID string           `json:"session_id" binding:"required"`
	Query     string           `json:"query" binding:"required"`
	UserID    string           `json:"user_id" binding:"required"`
	MessageOP string           `json:"message_op"` //TODO:
	MessageID string           `json:"message_id"`
	Files     []my_models.File `json:"file"`
}

type DialogResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id"`
	Reply     string `json:"reply"`
	SessionID string `json:"session_id"`
	Timestamp int64  `json:"timestamp"`
}

// 用户会话详情响应结构
type UserSessionsResponse struct {
	Success  bool                  `json:"success"`
	UserID   string                `json:"user_id"`
	Count    int                   `json:"count"`
	Sessions []SessionWithMessages `json:"sessions"`
}

// 会话及其消息的结构
type SessionWithMessages struct {
	Session         models.Session   `json:"session"`
	Messages        []models.Message `json:"messages"`
	MessageCount    int              `json:"message_count"`
	LastMessageTime *time.Time       `json:"last_message_time,omitempty"`
}

// 模拟流式 LLM 调用（返回多个chunk）
func mockLLMStream(prompt string) []string {
	log.Printf("调用LLM，Prompt=%s", prompt)
	// 模拟 LLM 流式生成回复的过程
	fullResponse := `这是一个基于您历史对话生成的模拟回复。我正在流式返回数据...
	SSE 限制：
	生产环境建议：对于生产环境，建议：
	使用 WebSocket 替代 SSE（支持双向通信）
	添加超时机制（如 30 秒后自动关闭连接）
	添加心跳包保持连接
	使用 Nginx 代理时需要配置 proxy_buffering off;
	Gin 中间件：如果需要支持大量并发 SSE 连接，可以添加专门的中间件：
	SSE 只支持服务器到客户端的单向通信，且对并发连接数有限制（浏览器通常限制为 6 个）。错误处理：需要处理连接中断、超时等情况。
	`

	return splitByRune(fullResponse)

}

// 方法1：按字符（rune）安全拆分 - 确保不切断中文字符
func splitByRune(text string) []string {
	// 将字符串转换为rune数组（一个rune代表一个Unicode字符）
	runes := []rune(text)
	chunks := []string{}

	// 每2-11个字符一个chunk（随机大小，但保证是完整字符）
	i := 0
	for i < len(runes) {
		// 随机chunk大小（1-10个字符）
		chunkSize := rand.Intn(10) + 1
		if chunkSize > len(runes)-i {
			chunkSize = len(runes) - i
		}

		chunk := string(runes[i : i+chunkSize])
		chunks = append(chunks, chunk)
		i += chunkSize
	}

	// 添加思考前缀
	if len(chunks) > 10 {
		prefixes := []string{"让我思考一下...", "嗯...", "根据您的提问，", "基于历史对话，", "我理解您想问的是，"}
		chunks = append(prefixes[:min(3, len(prefixes))], chunks...)
	}

	return chunks
}

// 辅助函数：将数据转为JSON字符串
func toJSON(data interface{}) string {
	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}

// 对话主处理函数
func handleDialog(c *gin.Context) {
	var req DialogRequest

	// 1. 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 2. 验证会话是否存在且属于该用户
	var session models.Session
	result := GlobalDB.Where("id = ? AND user_id = ? AND deleted = ?",
		req.SessionID, req.UserID, false).First(&session)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "会话不存在或无权限访问",
		})
		return
	}

	// 3. 查找历史消息（按时间排序）
	var historyMessages []models.Message
	GlobalDB.Where("session_id = ? AND deleted = ?", req.SessionID, false).
		Order("created_at ASC").
		Find(&historyMessages)

	// 4. 构建包含上下文的Prompt
	prompt := buildPrompt(historyMessages, req.Query)
	log.Printf("构建的Prompt长度: %d 字符", len(prompt))

	// 5. 调用模拟LLM获取回复
	reply := mockLLMCall(prompt)

	// 6. 保存用户消息到数据库
	userMsg := models.Message{
		ID:        uuid.NewString(),
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Query,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  map[string]interface{}{"source": "web_dialog"},
	}
	GlobalDB.Create(&userMsg)

	// 7. 保存助手回复到数据库
	assistantMsgID := uuid.NewString()
	assistantMsg := models.Message{
		ID:        assistantMsgID,
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   reply,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"model":  "mock_llm",
			"tokens": len(reply),
		},
	}
	GlobalDB.Create(&assistantMsg)

	// 8. 更新会话的更新时间
	GlobalDB.Model(&session).Update("updated_at", time.Now())

	// 9. 返回响应
	resp := DialogResponse{
		Success:   true,
		MessageID: assistantMsgID,
		Reply:     reply,
		SessionID: req.SessionID,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, resp)
}

// 构建包含上下文的Prompt
func buildPrompt(history []models.Message, currentQuery string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("以下是历史对话上下文：\n\n")

	// 添加历史消息（限制最近10条避免过长）
	maxHistory := 10
	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}

	for _, msg := range history {
		role := "用户"
		if msg.Role == "assistant" {
			role = "助手"
		}
		promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	promptBuilder.WriteString("\n当前用户问题: ")
	promptBuilder.WriteString(currentQuery)
	promptBuilder.WriteString("\n\n请基于以上对话历史，以助手的身份给出回复：")

	return promptBuilder.String()
}

// 创建新会话的接口-与第一个问题绑定
type CreateSessionRequest struct {
	UserID    string        `json:"user_id" binding:"required"`
	Title     string        `json:"title" binding:"required"`
	ProjectID string        `json:"project_id"`
	Query     string        `json:"query"`
	File      []models.File `json:"file"`
}

// func handleCreateSession(c *gin.Context) {
// 	var req CreateSessionRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if req.ProjectID == "" {
// 		req.ProjectID = "system"
// 	}

// 	session := models.Session{
// 		ID:        uuid.NewString(),
// 		ProjectID: req.ProjectID,
// 		UserID:    req.UserID,
// 		Title:     req.Title,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 		Source:    "user_create",
// 	}

// 	db.Create(&session)

// 	//转发到流式接口
// 	dialogReq := DialogRequest{
// 		UserID:    req.UserID,
// 		SessionID: session.ID,
// 		Query:     req.Query,
// 		Files:     req.File,
// 		MessageID: "",       //上一条消息id，第一次对话为空
// 		MessageOP: "create", //操作类型，create表示创建新对话
// 	}
// 	handleDialogStreamWithResumeInner(c, dialogReq)

// 	c.JSON(http.StatusOK, gin.H{
// 		"success":    true,
// 		"info":       "会话创建成功",
// 		"session_id": session.ID,
// 		"title":      session.Title,
// 	})
// }

// 获取会话历史
func handleGetHistory(c *gin.Context) {
	sessionID := c.Query("session_id")
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数"})
		return
	}
	if sessionID != "" {
		// 验证会话权限
		var session models.Session
		if GlobalDB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
			return
		}

		var messages []models.Message
		GlobalDB.Where("session_id = ? AND deleted = ?", sessionID, false).
			Order("created_at ASC").
			Find(&messages)

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"session":  session,
			"messages": messages,
			"count":    len(messages),
		})
	} else {
		// 验证会话权限
		var sessions []models.Session
		if GlobalDB.Where(" user_id = ? AND deleted = ?", userID, false).Order("updated_at DESC").Find(&sessions).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "查询会话失败"})
			return
		}

		// 2. 为每个会话查询对应的消息
		var responseSessions []SessionWithMessages

		for _, session := range sessions {
			var messages []models.Message

			// 查询该会话的所有消息（按时间排序）
			GlobalDB.Where("session_id = ? AND deleted = ?", session.ID, false).
				Order("created_at ASC").
				Find(&messages)

			// 获取最后一条消息的时间
			var lastMessageTime *time.Time
			if len(messages) > 0 {
				lastMessageTime = &messages[len(messages)-1].CreatedAt
			}

			sessionWithMessages := SessionWithMessages{
				Session:         session,
				Messages:        messages,
				MessageCount:    len(messages),
				LastMessageTime: lastMessageTime,
			}

			responseSessions = append(responseSessions, sessionWithMessages)
		}

		// 3. 构建响应
		response := UserSessionsResponse{
			Success:  true,
			UserID:   userID,
			Count:    len(responseSessions),
			Sessions: responseSessions,
		}

		c.JSON(http.StatusOK, response)

	}
}

func main() {
	// 初始化
	// initDB()

	ws := new(restful.WebService)
	ws.
		Path("/api/v1/applet/ai").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	//项目
	//创建一个项目，指定标题（可选）
	ws.Route(ws.POST("/projects").To(my_handler.CreateProjectHandler()).
		Doc("Create a new project").
		Param(ws.BodyParameter("request", "CreateProjectReq").DataType("my_requests.CreateProjectReq")).
		Returns(201, "Created", my_response.CreateProjectResponse{}).
		Returns(400, "Bad Request", nil))

	//更新项目标题
	ws.Route(ws.PATCH("/projects/{projectId}").To(my_handler.UpdateProjectHandler()).
		Doc("Update a project title").
		Param(ws.PathParameter("projectId", "Project ID").DataType("string")).
		Param(ws.BodyParameter("request", "UpdateProjectReq").DataType("my_requests.UpdateProjectReq")).
		Returns(200, "OK", my_response.UpdateProjectResponse{}).
		Returns(400, "Bad Request", nil))

	//查询所有项目
	ws.Route(ws.GET("/projects").To(my_handler.ListProjectsHandler()).
		Doc("List all projects").
		Returns(200, "OK", my_response.ListProjectsResponse{}).
		Returns(400, "Bad Request", nil))

	//删除一个项目
	ws.Route(ws.DELETE("/projects/{projectId}").To(my_handler.DeleteProjectHandler()).
		Doc("Delete a project").
		Param(ws.PathParameter("projectId", "Project ID").DataType("string")).
		Returns(204, "No Content", nil).
		Returns(400, "Bad Request", nil))

	// 查询某个项目下的所有会话
	ws.Route(ws.GET("/projects/{projectId}/sessions").To(my_handler.ListSessionsHandler()).
		Doc("List all sessions under a project").
		Param(ws.PathParameter("projectId", "Project ID").DataType("string")).
		Returns(200, "OK", my_response.ListSessionsResponse{}).
		Returns(401, "Unauthorized", nil).
		Returns(403, "Forbidden", nil).
		Returns(404, "Not Found", nil))

	//会话
	//创建一个会话并对话，sse流式响应
	ws.Route(ws.POST("/sessions/stream").
		To(my_handler.CreateSessionStreamChatHandler).
		Doc("Create session and chat (SSE)").
		Consumes(restful.MIME_JSON).
		Produces("text/event-stream").
		Param(ws.BodyParameter("request", "CreateSessionAndChatReq").
			DataType("my_requests.CreateSessionAndChatReq")).
		Returns(200, "OK", nil))

	//查询某个会话所有消息
	ws.Route(ws.GET("/sessions/{sessionId}/messages").To(my_handler.ListMessagesBySessionHandler).
		Doc("Get session history").
		Param(ws.PathParameter("sessionId", "Session ID").DataType("string")).
		Returns(200, "OK", my_response.ListMessagesResponse{}).
		Returns(400, "Bad Request", nil))

	//获取会话历史
	// ws.Route(ws.GET("/sessions/history").To(my_handler.GetHistoryHandler()).
	// 	Doc("Get session history").
	// 	Param(ws.QueryParameter("session_id", "Session ID").DataType("string")).
	// 	Param(ws.QueryParameter("user_id", "User ID").DataType("string")).
	// 	Returns(200, "OK", nil).
	// 	Returns(400, "Bad Request", nil))

	restful.Add(ws)
	restful.EnableTracing(true)

	http.ListenAndServe(":8080", nil)

	// 创建Gin引擎
	// r := gin.Default()

	// // 添加中间件
	// r.Use(gin.Logger())
	// r.Use(gin.Recovery())

	// // 添加CORS支持（前端调用需要）
	// r.Use(func(c *gin.Context) {
	// 	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	// 	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	// 	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 	if c.Request.Method == "OPTIONS" {
	// 		c.AbortWithStatus(204)
	// 		return
	// 	}

	// 	c.Next()
	// })

	// // 注册路由
	// r.GET("/health", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now()})
	// })

	// // 流式对话路由
	// r.POST("/api/dialog/stream", handleDialogStreamWithResumeApi) // 支持断点续传的流式对话接口
	// r.GET("/api/stream/status", handleStreamStatus)               // 查询某个会话状态接口

	// r.POST("/api/session/create", handleCreateSession)
	// r.GET("/api/session/query", handleGetHistory)
	// r.GET("/api/session/queryAll", handleGetHistory) // 兼容旧接口

	// // 项目相关接口
	// r.GET("/api/project/query", handleGetProject)

	// // 启动服务器
	// port := ":8080"
	// log.Printf("服务器启动在 http://localhost%s", port)
	// log.Println("可用接口:")
	// log.Println("  GET  /health              - 服务器健康检查")
	// log.Println("  POST /api/dialog          - 对话接口")
	// log.Println("  POST /api/session/create  - 创建新会话")
	// log.Println("  GET  /api/session/query   - 获取会话历史")
	// log.Println("  GET  /api/session/queryAll   - 获取所有会话历史")

	// if err := r.Run(port); err != nil {
	// 	log.Fatal("服务器启动失败:", err)
	// }
}
