package service

import (
	"errors"
	"log"
	"net/http"
	constant "session-demo/const"
	"session-demo/models"
	"session-demo/requests"
	"session-demo/response"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultProjectTitle = "新项目"
)

// 创建一个项目
func CreateProject(req *requests.CreateAndEditProjectReq, userID string) (*models.Project, error) {
	//业务层校验参数
	if req.Title == "" {
		req.Title = defaultProjectTitle
	}

	project := &models.Project{
		ID:                uuid.New().String(),
		Title:             req.Title,
		UserID:            userID,
		CustomInstruction: req.CustomInstruction,
		Files:             req.Files,
		ToolsConfig:       req.ToolConfig,
		ModelSvcsConfig:   req.ModelServiceConfig,
	}
	if err := Dbservice.DB.Create(project).Error; err != nil {
		return nil, &response.BizError{HttpStatus: http.StatusInternalServerError, Msg: err.Error()}
	}
	log.Println("Created project:", project)
	return project, nil
}

// 更新项目标题和其他字段
func UpdateProject(req *requests.CreateAndEditProjectReq, projectID string, userID string) (*models.Project, error) {

	if req.Title == "" {
		return nil, constant.ErrBadRequest
	}

	var project models.Project
	//权限
	if err := Dbservice.DB.Where("id = ? AND user_id = ? AND deleted = ?", projectID, userID, false).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，或者属于别人（对用户来说都是没权限或没找到）
			return nil, constant.ErrUnauthorized
		}
		// 数据库错误，记录日志，不要把 err.Error() 暴露给前端
		log.Printf("[DB_ERROR] Failed to find project: %v", err)
		return nil, &response.BizError{
			HttpStatus: http.StatusInternalServerError,
			Code:       500,
			Msg:        "Failed to update project",
		}
	}

	project.Title = req.Title
	project.CustomInstruction = req.CustomInstruction
	project.Files = req.Files
	project.ToolsConfig = req.ToolConfig
	project.ModelSvcsConfig = req.ModelServiceConfig
	if err := Dbservice.DB.Save(&project).Error; err != nil {
		return nil, &response.BizError{HttpStatus: http.StatusInternalServerError, Msg: err.Error()}
	}
	log.Printf("[INFO] User %s updated project %s", userID, projectID)
	return &project, nil
}

// ListProjects 列出某个用户的所有项目
func ListProjects(userID string) ([]models.Project, error) {
	// 查询数据库
	var projects []models.Project
	log.Println("Listing projects for userID:", userID)
	err := Dbservice.DB.Where("user_id = ?", userID).Find(&projects).Error
	if err != nil {
		return nil, &response.BizError{HttpStatus: http.StatusInternalServerError, Msg: err.Error()}
	}
	log.Println("Found projects:", projects)
	return projects, nil
}

// 删除一个项目
func DeleteProject(projectID string, userID string) error {
	if err := Dbservice.DB.Delete(&models.Project{}, "id = ? and user_id = ?", projectID, userID).Error; err != nil {
		return &response.BizError{HttpStatus: http.StatusInternalServerError, Msg: err.Error()}
	}
	log.Printf("[INFO] User %s deleted project %s", userID, projectID)
	return nil
}

// GetProjectById 获取项目详情
func GetProjectById(projectID string) (*models.Project, error) {
	var project models.Project
	if err := Dbservice.DB.First(&project, "id = ?", projectID).Error; err != nil {
		return nil, &response.BizError{HttpStatus: http.StatusInternalServerError, Msg: err.Error()}
	}
	// log.Printf("[INFO] User %s get project %s", userID, projectID)
	return &project, nil
}
