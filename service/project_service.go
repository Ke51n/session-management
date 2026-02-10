package service

import (
	"log"
	my_models "session-demo/models"

	"github.com/google/uuid"
)

// 创建一个项目
func CreateOrEditProject(title string, customInstruction string, files []my_models.File, toolsConfig, modelSvcsConfig my_models.JSONMap, userID string) (*my_models.Project, error) {
	log.Println("Creating project with title:", title, "for userID:", userID)
	project := &my_models.Project{
		ID:                uuid.New().String(),
		Title:             title,
		UserID:            userID,
		CustomInstruction: customInstruction,
		Files:             files,
		ToolsConfig:       toolsConfig,
		ModelSvcsConfig:   modelSvcsConfig,
	}
	if err := My_dbservice.DB.Create(project).Error; err != nil {
		return nil, err
	}
	log.Println("Created project:", project)
	return project, nil
}

// 更新项目标题和其他字段
func UpdateProjectTitle(projectID string, newTitle string, customInstruction string, files []my_models.File, toolsConfig, modelSvcsConfig my_models.JSONMap, userID string) (*my_models.Project, error) {
	var project my_models.Project
	//权限
	if err := My_dbservice.DB.First(&project, "id = ? and user_id = ?", projectID, userID).Error; err != nil {
		return nil, err
	}
	project.Title = newTitle
	project.CustomInstruction = customInstruction
	project.Files = files
	project.ToolsConfig = toolsConfig
	project.ModelSvcsConfig = modelSvcsConfig
	if err := My_dbservice.DB.Save(&project).Error; err != nil {
		return nil, err
	}
	log.Println("Updated project", project)
	return &project, nil
}

// ListProjects 列出某个用户的所有项目
func ListProjects(userID string) ([]my_models.Project, error) {
	// 查询数据库
	var projects []my_models.Project
	log.Println("Listing projects for userID:", userID)
	err := My_dbservice.DB.Where("user_id = ?", userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	log.Println("Found projects:", projects)
	return projects, nil
}

// 删除一个项目
func DeleteProject(projectID string, userID string) error {
	if err := My_dbservice.DB.Delete(&my_models.Project{}, "id = ? and user_id = ?", projectID, userID).Error; err != nil {
		return err
	}
	log.Println("Deleted project with ID:", projectID)
	return nil
}

// GetProjectById 获取项目详情
func GetProjectById(projectID string) (*my_models.Project, error) {
	var project my_models.Project
	if err := My_dbservice.DB.First(&project, "id = ?", projectID).Error; err != nil {
		return nil, err
	}
	log.Println("Found project:", project)
	return &project, nil
}
