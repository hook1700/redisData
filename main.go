package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"redisData/dao/mysql"
	"redisData/dao/redis"
	"redisData/logger"
	"redisData/logic"
	"redisData/routes"
	"redisData/setting"
)

func main() {

	//初始化viper
	if err := setting.Init(""); err != nil {
		zap.L().Error("viper init fail", zap.Error(err))
		return
	}

	//初始化日志
	if err := logger.InitLogger(viper.GetString("mode")); err != nil {
		zap.L().Error("init logger fail err", zap.Error(err))
		return
	}
	defer zap.L().Sync() //把缓冲区的日志添加
	zap.L().Debug("init logger success")

	//初始化MySQL
	if err := mysql.InitMysql(); err != nil {
		zap.L().Error("init mysql fail err", zap.Error(err))
		return
	}
	defer mysql.Close()

	//初始化redis
	if err := redis.InitClient(); err != nil {
		zap.L().Error("init redis fail err", zap.Error(err))
		return
	}
	defer redis.Close()


	//初始化数据 开始存redis数据
	logic.InitStartData()

	fmt.Println("success")
	//初始化routes
	r := routes.SetUp()
	r.Run(fmt.Sprintf(":%d",viper.GetInt("port")))

	//宕机处理
	defer func() {
		recover()
	}()

}
