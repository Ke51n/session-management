package service

import (
	"errors"
	"log"
	"net/http"
	"session-management/dao"
	"session-management/models"
	"session-management/requests"
	"session-management/response"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultProjectTitle = "新项目"
)

// 创建一个项目
func CreateProject(req *requests.CreateAndUpdateProjectReq, userID string) (*models.Project, error) {
	//业务层校验参数
	if req.Title == "" {
		req.Title = defaultProjectTitle
	}

	project := &models.Project{
		ID:                uuid.New().String(),
		Title:             req.Title,
		Source:            req.Source,
		UserID:            userID,
		CustomInstruction: req.CustomInstruction,
		Files:             req.Files,
		ToolsConfig:       req.ToolConfig,
		ModelSvcsConfig:   req.ModelServiceConfig,
		Extension:         req.Extension,
	}
	log.Printf("[INFO] Creating project %v", project) // 记录创建的项目信息，注意不要记录敏感信息
	if _, err := dao.AiAppUniDAO.CreateProject(project); err != nil {
		return nil, response.WrapError(500, "创建项目失败", err)
	}
	return project, nil
}

// 更新项目标题和其他字段
func UpdateProject(req *requests.CreateAndUpdateProjectReq, projectID string, userID string) (*models.Project, error) {

	if req.Title == "" {
		return nil, &response.BizError{HttpStatus: http.StatusBadRequest, Code: 400, Msg: "项目标题不能为空"}
	}

	var project models.Project
	//权限
	if err := Dbservice.DB.Where("id = ? AND user_id = ? AND deleted = ?", projectID, userID, false).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，或者属于别人（对用户来说都是没权限或没找到）
			return nil, &response.BizError{HttpStatus: http.StatusNotFound, Code: 403, Msg: "项目不存在"}
		}
		// 数据库错误，记录日志，不要把 err.Error() 暴露给前端
		log.Printf("[DB_ERROR] Failed to find project: %v", err)
		return nil, response.WrapError(500, "更新项目失败", err)
	}

	project.Title = req.Title
	project.Source = req.Source
	project.CustomInstruction = req.CustomInstruction
	project.Files = req.Files
	project.ToolsConfig = req.ToolConfig
	project.ModelSvcsConfig = req.ModelServiceConfig
	project.Version++
	project.UpdatedAt = time.Now()
	// 更新数据库
	if err := dao.AiAppUniDAO.UpdateProject(&project); err != nil {
		return nil, response.WrapError(500, "更新项目失败", err)
	}
	log.Printf("[INFO] User %s updated project %s", userID, projectID)
	return &project, nil
}

// ListProjects 列出某个用户的所有项目
func ListProjects(userID string) ([]models.Project, error) {
	// 查询数据库
	var projects []models.Project
	log.Println("Listing projects for userID:", userID)
	projects, err := dao.AiAppUniDAO.ListProjects(userID)
	if err != nil {
		return nil, response.WrapError(500, "查询项目失败", err)
	}
	return projects, nil
}

// 删除一个项目
func DeleteProject(projectID string, userID string) error {
	//查找项目
	var project models.Project
	if foundProject, err := dao.AiAppUniDAO.FindProject(userID, projectID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.WrapError(404, "项目不存在", err)
		}
	} else {
		project = *foundProject
	}

	project.Version++
	project.UpdatedAt = time.Now()
	project.Deleted = true // 软删除
	if err := Dbservice.DB.Model(&project).Omit("created_at").Updates(project).Error; err != nil {
		return response.WrapError(500, "删除项目失败", err)
	}
	log.Printf("[INFO] User %s deleted project %s", userID, projectID)
	return nil
}

// GetProjectById 获取项目详情
func GetProjectById(userID, projectID string) (*models.Project, error) {

	var project models.Project
	if foundProject, err := dao.AiAppUniDAO.FindProject(userID, projectID); err != nil {
		return nil, response.WrapError(500, "获取项目失败", err)
	} else {
		project = *foundProject
	}
	return &project, nil
}
