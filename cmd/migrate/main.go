package main

import (
	"fmt"
	"log"
	"os"

	"app/config"
	"app/internal/model"
	"app/pkg/database"
)

func main() {
	// 初始化配置
	err := config.Init()
	if err != nil {
		fmt.Printf("配置初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化数据库连接
	err = database.InitGormDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer database.CloseGormDB()

	// 获取数据库连接
	db := database.GetGormDB()
	if db == nil {
		log.Fatal("获取数据库连接失败")
	}

	log.Println("开始迁移数据库表结构...")

	// 自动迁移数据库表结构
	models := []interface{}{
		&model.User{},
		&model.SMSRecord{},
		&model.Relation{},
		&model.Post{},
		&model.PostComment{},
		// 在此处添加其他模型
	}

	for _, m := range models {
		modelName := fmt.Sprintf("%T", m)
		log.Printf("正在迁移模型: %s", modelName)

		err := db.AutoMigrate(m)
		if err != nil {
			log.Printf("迁移模型 %s 失败: %v", modelName, err)
			os.Exit(1)
		}
	}

	log.Println("数据库表结构迁移完成")
}
