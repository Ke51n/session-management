package service

import (
	"log"
	my_models "session-demo/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 会话服务
type DBService struct {
	DB *gorm.DB
}

var My_dbservice *DBService

// 初始化
func init() {
	initDB()
	log.Println("SessionService initialized")
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
	err = db.AutoMigrate(&my_models.Session{}, &my_models.Message{}, &my_models.Project{})
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 创建一些测试数据
	// createTestData()

	log.Println("数据库初始化完成1")

	My_dbservice = &DBService{DB: db}
}
