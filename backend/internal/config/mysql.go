package config

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectMysql() *gorm.DB {
	cfg := GlobalConfig
	if cfg == nil {
		LoadConfig()
		cfg = GlobalConfig
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("数据库连接失败", "err", err)
		panic(err)
	}
	slog.Info("数据库连接成功", "host", cfg.MySQL.Host, "db", cfg.MySQL.DBName)
	return db
}