package mysql

import (
	"github.com/spf13/viper"
	"redisData/pkg/mysql"
	"time"
)

func InitMysql() {
	// 建立数据库连接池
	db := mysql.ConnectDB()
	// 命令行打印数据库请求的信息
	sqlDB, _ := db.DB()
	// 设置最大连接数
	sqlDB.SetMaxOpenConns(viper.GetInt("mysql.max_open_connections"))
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(viper.GetInt("mysql.max_idle_connections"))
	// 设置每个链接的过期时间
	sqlDB.SetConnMaxLifetime(time.Duration(viper.GetInt("mysql.max_life_seconds")) * time.Second)
}
