package dao

import (
	"log"
	"session-management/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// UniDAO 统一数据访问对象，封装数据库操作
type UniDAO struct {
	db *gorm.DB
}

var AiAppUniDAO *UniDAO

// 初始化
func init() {
	initDB()
	log.Println("AiAppUniDAO initialized")
}

// 初始化数据库
func initDB() {

	dsn := "gormuser:gorm123@tcp(127.0.0.1:3306)/gorm_test?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	// dsn := "gormuser:gorm123@tcp(127.0.0.1:3306)/gorm_test?charset=utf8mb4&parseTime=True&loc=Local"

	// 使用MySQL数据库（便于演示，生产环境请用MySQL/PostgreSQL）
	// db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&models.Session{}, &models.Message{}, &models.Project{})
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 创建一些测试数据
	// createTestData()

	log.Println("dao初始化完成1")

	AiAppUniDAO = &UniDAO{db: db}
}
