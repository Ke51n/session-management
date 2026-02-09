package service

import (
	"log"
	my_models "session-demo/models"

	"github.com/google/uuid"
)

// 创建一个项目
func CreateProject(title string, userID string) (*my_models.Project, error) {
	log.Println("Creating project with title:", title, "for userID:", userID)
	project := &my_models.Project{
		ID:     uuid.New().String(),
		Title:  title,
		UserID: userID,
	}
	if err := My_dbservice.DB.Create(project).Error; err != nil {
		return nil, err
	}
	log.Println("Created project:", project)
	return project, nil
}

// 更新项目标题
func UpdateProjectTitle(projectID string, newTitle string, userID string) (*my_models.Project, error) {
	var project my_models.Project
	if err := My_dbservice.DB.First(&project, "id = ? and user_id = ?", projectID, userID).Error; err != nil {
		return nil, err
	}
	project.Title = newTitle
	if err := My_dbservice.DB.Save(&project).Error; err != nil {
		return nil, err
	}
	log.Println("Updated project title:", project)
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
func GetProjectById(userID string, projectID string) (*my_models.Project, error) {
	var project my_models.Project
	if err := My_dbservice.DB.First(&project, "id = ? and user_id = ?", projectID, userID).Error; err != nil {
		return nil, err
	}
	log.Println("Found project:", project)
	return &project, nil
}
