package bootstrap

import (
	"github.com/assimon/captcha-bot/telegram"
	"github.com/assimon/captcha-bot/util/config"
	"github.com/assimon/captcha-bot/util/log"
	"github.com/assimon/captcha-bot/util/orm"
	"os"
	"os/signal"
	"syscall"
)

// Start 服务启动
func Start() {
	config.InitConfig()
	log.InitLog()
	orm.InitDb()
	// 机器人启动
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Sugar.Error("server bot err:", err)
			}
		}()
		telegram.BotStart()
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}
