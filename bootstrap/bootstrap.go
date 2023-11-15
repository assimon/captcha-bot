package bootstrap

import (
	"github.com/assimon/captcha-bot/telegram"
	"github.com/assimon/captcha-bot/util/config"
	E "github.com/assimon/captcha-bot/util/error"
	"github.com/assimon/captcha-bot/util/log"
	"github.com/assimon/captcha-bot/util/orm"
	"github.com/assimon/captcha-bot/util/sensitiveword"
	"os"
	"os/signal"
	"syscall"
)

// Start 服务启动
func Start() {
	config.InitConfig()
	log.InitLog()
	orm.InitDb()
	sensitiveword.InitSensitiveWord()
	// 机器人启动
	go E.MustPanicErrorFunc(func() {
		telegram.BotStart()
	})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}
