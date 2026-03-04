package dao

import (
	"log"
	"session-management/models"
	"time"
)

// CreateProject 保存项目到数据库
func (d UniDAO) CreateProject(project *models.Project) (string, error) {
	//插入数据库
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	err := d.db.Create(project).Error
	if err != nil {
		return "", err
	}
	return project.ID, nil

}

// UpdateProject 更新项目到数据库
func (d UniDAO) UpdateProject(project *models.Project) error {
	// 只更新非零字段，跳过创建时间
	err := d.db.Model(project).Omit("created_at").Updates(project).Error
	return err
}

// FindProject 查询用户的一个项目
func (d UniDAO) FindProject(userID string, projectID string) (*models.Project, error) {
	var project models.Project
	err := d.db.Where("id = ? AND user_id = ? AND deleted = ?", projectID, userID, false).First(&project).Error
	if err != nil {
		log.Printf("[DB_ERROR] Failed to find project: %v", err)
		return nil, err
	}
	return &project, nil
}

// ListProjects 查询用户的所有项目
func (d UniDAO) ListProjects(userID string) ([]models.Project, error) {
	var projects []models.Project
	err := d.db.Where("user_id = ? AND deleted = ?", userID, false).Find(&projects).Error
	if err != nil {
		log.Printf("[DB_ERROR] Failed to list projects: %v", err)
		return nil, err
	}
	return projects, nil
}
