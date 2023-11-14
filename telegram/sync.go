package telegram

import (
	"github.com/assimon/captcha-bot/service"
	E "github.com/assimon/captcha-bot/util/error"
	"github.com/assimon/captcha-bot/util/log"
	tb "gopkg.in/telebot.v3"
)

import (
	"github.com/robfig/cron"
)

func RunSyncTask() {
	c := cron.New()
	c.AddFunc("*/10 * * * * *", func() {
		go func() {
			E.MustPanicErrorFunc(TimeoutLeaveGroupToUser)
		}()
	})
	c.Start()
}

// TimeoutLeaveGroupToUser 一直没有验证就把用户踢出群
func TimeoutLeaveGroupToUser() {
	records, err := service.GetTimeoutCaptchaRecords()
	if err != nil {
		log.Sugar.Error("[TimeoutLeaveGroupToUser] GetTimeoutCaptchaRecords err:", err)
		return
	}
	for _, record := range records {
		// 设置超时
		err = service.TimeoutRecordByCaptchaId(record.CaptchaId)
		if err != nil {
			log.Sugar.Error("[TimeoutLeaveGroupToUser] TimeoutRecordByCaptchaId err:", err)
			continue
		}
		// 先ban再unban，就可以实现删除用户而不是封禁，貌似只能这样，TG没有提供单独的删除用户出群聊的方法，很扯淡
		Bot.Ban(&tb.Chat{ID: record.TelegramChatId}, &tb.ChatMember{User: &tb.User{ID: record.TelegramUserId}}, false)
		Bot.Unban(&tb.Chat{ID: record.TelegramChatId}, &tb.User{ID: record.TelegramUserId})
	}
}
