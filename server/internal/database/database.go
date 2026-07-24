package database

import (
	"log"
	"time"

	"munch/server/internal/config"
	"munch/server/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New 建立 GORM 连接并自动迁移表结构。
func New(cfg *config.Config) (*gorm.DB, error) {
	// 先把解析到的连接目标打出来（不含密码），连不上时一眼能看出是不是环境变量没配。
	log.Printf("[db] 连接目标 %s:%s 库=%s 用户=%s", cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser)
	if cfg.DBHost == "127.0.0.1" || cfg.DBHost == "localhost" {
		log.Printf("[db] ⚠️ 连接地址是本机，云端部署时通常意味着 MYSQL_ADDRESS 或 DB_HOST 环境变量没配")
	}

	var db *gorm.DB
	var err error
	// 数据库可能因冷启动/自动启停短暂不可用，重试几次再放弃，
	// 避免容器起不来直接进入 Back-off 重启循环。
	const attempts = 10
	for i := 1; i <= attempts; i++ {
		db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			break
		}
		log.Printf("[db] 第 %d/%d 次连接失败：%v", i, attempts, err)
		if i < attempts {
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		return nil, err
	}
	log.Println("[db] connected & migrated")
	return db, nil
}
