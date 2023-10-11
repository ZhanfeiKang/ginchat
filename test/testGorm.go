package main

import (
	"ginchat/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:abc123@tcp(127.0.0.1:3306)/ginchat?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移 schema (表示没有这个表，就会帮我创建)
	// db.AutoMigrate(&models.UserBasic{})
	// db.AutoMigrate(&models.Message{})
	db.AutoMigrate(&models.GroupBasic{})
	db.AutoMigrate(&models.Contact{})

	// Create
	// user := &models.UserBasic{}
	// user.Name = "东契奇"
	// db.Create(user)

	// // Read
	// fmt.Println(db.First(user, 1)) // 根据整型主键查找
	// // db.First(user, "code = ?", "D42") // 查找 code 字段值为 D42 的记录

	// // Update
	// db.Model(user).Update("PassWord", "abc123")
}
