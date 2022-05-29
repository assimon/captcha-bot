package telegram

import (
	"fmt"
	"github.com/assimon/captcha-bot/model"
	"github.com/assimon/captcha-bot/service"
	"github.com/assimon/captcha-bot/util/captcha"
	"github.com/assimon/captcha-bot/util/config"
	"github.com/assimon/captcha-bot/util/log"
	"github.com/golang-module/carbon/v2"
	uuid "github.com/satori/go.uuid"
	tb "gopkg.in/telebot.v3"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	joinMessageMenu = &tb.ReplyMarkup{ResizeKeyboard: true}
	manageBanBtn    = joinMessageMenu.Data("👮‍管理员禁止🈲", "manageBanBtn")
	managePassBtn   = joinMessageMenu.Data("👮‍管理员通过✅", "managePassBtn")
)

var (
	captchaMessageMenu = &tb.ReplyMarkup{ResizeKeyboard: true}
)

var (
	gUserCaptchaCodeTable    = service.NewCaptchaCodeTable()
	gUserCaptchaPendingTable = service.NewCaptchaPendingTable()
)

var (
	gMessageTokenMap sync.Map
)

// StartCaptcha 开始验证
func StartCaptcha(c tb.Context) error {
	chatToken := c.Message().Payload
	// 不是私聊或者载荷为空
	if !c.Message().Private() || chatToken == "" {
		return nil
	}
	payload, ok := gMessageTokenMap.Load(chatToken)
	if !ok {
		return nil
	}
	// payload不能正常解析
	payloadSlice := strings.Split(payload.(string), "|")
	if len(payloadSlice) != 3 {
		return nil
	}
	pendingMessageId, err := strconv.Atoi(payloadSlice[0])
	groupId, err := strconv.ParseInt(payloadSlice[1], 10, 64)
	groupTitle := payloadSlice[2]
	if err != nil {
		log.Sugar.Error("[StartCaptcha] groupId err:", err)
		return nil
	}
	userId := c.Sender().ID
	pendingKey := fmt.Sprintf("%d|%d", pendingMessageId, groupId)
	record := gUserCaptchaPendingTable.Get(pendingKey)
	if record == nil || record.UserId != c.Sender().ID {
		return c.Send("您在该群没有待验证记录😁")
	}
	// 获得一个验证码
	captchaCode, imgUrl, err := captcha.GetCaptcha()
	if err != nil {
		log.Sugar.Error("[StartCaptcha] get image captcha err:", err)
		return c.Send("服务器异常~，请稍后再试")
	}
	captchaMessage := fmt.Sprintf(config.MessageC.CaptchaImage,
		groupTitle,
		config.SystemC.CaptchaTimeout,
	)
	sendMessage := &tb.Photo{
		File:    tb.FromDisk(imgUrl),
		Caption: captchaMessage,
	}
	refreshCaptchaImageBtn := captchaMessageMenu.Data("🔁刷新验证码", "refreshCaptchaImageBtn", strconv.FormatInt(userId, 10))
	Bot.Handle(&refreshCaptchaImageBtn, refreshCaptcha())
	captchaMessageMenu.Inline(
		captchaMessageMenu.Row(refreshCaptchaImageBtn),
	)
	botMsg, err := Bot.Send(c.Chat(), sendMessage, captchaMessageMenu)
	if err != nil {
		log.Sugar.Error("[StartCaptcha] send image captcha err:", err)
		return c.Send("服务器异常~，请稍后再试")
	}
	userCaptchaCodeVal := &service.CaptchaCode{
		UserId:         userId,
		GroupId:        groupId,
		Code:           captchaCode,
		CaptchaMessage: botMsg,
		PendingMessage: record.PendingMessage,
		GroupTitle:     groupTitle,
		CreatedAt:      carbon.Now().Timestamp(),
	}
	userCaptchaCodeKey := strconv.FormatInt(userId, 10)
	gUserCaptchaCodeTable.Set(userCaptchaCodeKey, userCaptchaCodeVal)
	time.AfterFunc(time.Duration(config.SystemC.CaptchaTimeout)*time.Second, func() {
		os.Remove(imgUrl)
		gMessageTokenMap.Delete(chatToken)
		gUserCaptchaCodeTable.Del(userCaptchaCodeKey)
		err = Bot.Delete(botMsg)
		if err != nil {
			log.Sugar.Error("[StartCaptcha] delete captcha err:", err)
		}
	})
	return nil
}

// OnTextMessage 文本消息
func OnTextMessage(c tb.Context) error {
	// 不是私聊
	if !c.Message().Private() {
		return nil
	}
	userIdStr := strconv.FormatInt(c.Sender().ID, 10)
	captchaCode := gUserCaptchaCodeTable.Get(userIdStr)
	if captchaCode == nil || captchaCode.UserId != c.Sender().ID {
		return nil
	}
	// 验证
	replyCode := c.Message().Text
	if !captcha.VerifyCaptcha(captchaCode.Code, replyCode) {
		return nil
	}
	// 解禁用户
	err := Bot.Restrict(&tb.Chat{ID: captchaCode.GroupId}, &tb.ChatMember{
		User:   &tb.User{ID: captchaCode.UserId},
		Rights: tb.NoRestrictions(),
	})
	if err != nil {
		log.Sugar.Error("[OnTextMessage] unban err:", err)
		return c.Send("服务器异常~，请稍后重试~")
	}
	gUserCaptchaCodeTable.Del(userIdStr)
	gUserCaptchaPendingTable.Del(fmt.Sprintf("%d|%d", captchaCode.PendingMessage.ID, captchaCode.PendingMessage.Chat.ID))
	//删除验证消息
	Bot.Delete(captchaCode.CaptchaMessage)
	Bot.Delete(captchaCode.PendingMessage)
	return c.Send(config.MessageC.VerificationComplete)

}

// UserJoinGroup 用户加群事件
func UserJoinGroup(c tb.Context) error {
	var err error
	err = c.Delete()
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] delete join message err:", err)
	}
	// 如果是管理员邀请的，直接通过
	if isManage(c.Chat(), c.Sender().ID) {
		return nil
	}
	// ban user
	err = Bot.Restrict(c.Chat(), &tb.ChatMember{
		Rights:          tb.NoRights(),
		User:            c.Message().UserJoined,
		RestrictedUntil: tb.Forever(),
	})
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] ban user err:", err)
	}
	joinMessage := fmt.Sprintf(config.MessageC.JoinHint, c.Message().UserJoined.Username, c.Chat().Title, config.SystemC.JoinHintAfterDelTime)
	chatToken := uuid.NewV4().String()
	doCaptchaBtn := joinMessageMenu.URL("👉🏻点我开始人机验证🤖", fmt.Sprintf("https://t.me/%s?start=%s", Bot.Me.Username, chatToken))

	joinMessageMenu.Inline(
		joinMessageMenu.Row(doCaptchaBtn),
		joinMessageMenu.Row(manageBanBtn, managePassBtn),
	)
	// 加载广告
	advertises, err := service.GetEfficientAdvertiseService()
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] load advertise err:", err)
	} else {
		for _, advertise := range advertises {
			joinMessageMenu.InlineKeyboard = append(joinMessageMenu.InlineKeyboard, []tb.InlineButton{
				{
					Text: advertise.Title,
					URL:  advertise.Url,
				},
			})
		}
	}
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] add captcha record err:", err)
	}
	captchaMessage, err := Bot.Send(c.Chat(), joinMessage, joinMessageMenu)
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] send join hint message err:", err)
	}
	// 设置token对于验证消息
	gMessageTokenMap.Store(chatToken, fmt.Sprintf("%d|%d|%s", captchaMessage.ID, c.Chat().ID, c.Chat().Title))
	captchaDataVal := &service.CaptchaPending{
		PendingMessage: captchaMessage,
		UserId:         c.Message().UserJoined.ID,
		GroupId:        c.Chat().ID,
		JoinAt:         carbon.Now().Timestamp(),
	}
	captchaDataKey := fmt.Sprintf("%d|%d", captchaMessage.ID, c.Chat().ID)
	gUserCaptchaPendingTable.Set(captchaDataKey, captchaDataVal)
	time.AfterFunc(time.Duration(config.SystemC.JoinHintAfterDelTime)*time.Second, func() {
		err = Bot.Delete(captchaMessage)
		if err != nil {
			log.Sugar.Error("[UserJoinGroup] delete join hint message err:", err)
		}
	})
	time.AfterFunc(time.Hour, func() {
		gUserCaptchaPendingTable.Del(captchaDataKey)
	})
	return err
}

// ManageBan 管理员手动禁止
func ManageBan() func(c tb.Context) error {
	return func(c tb.Context) error {
		key := fmt.Sprintf("%d|%d", c.Callback().Message.ID, c.Chat().ID)
		record := gUserCaptchaPendingTable.Get(key)
		if record.UserId > 0 {
			gUserCaptchaPendingTable.Del(key)
		}
		return c.Delete()
	}
}

// ManagePass 管理员手动通过
func ManagePass() func(c tb.Context) error {
	return func(c tb.Context) error {
		key := fmt.Sprintf("%d|%d", c.Callback().Message.ID, c.Chat().ID)
		record := gUserCaptchaPendingTable.Get(key)
		if record != nil && record.UserId > 0 {
			// 解禁用户
			err := Bot.Restrict(c.Chat(), &tb.ChatMember{
				User:   &tb.User{ID: record.UserId},
				Rights: tb.NoRestrictions(),
			})
			if err != nil {
				log.Sugar.Error("[ManagePass] unban err:", err)
			}
			gUserCaptchaPendingTable.Del(key)
		}
		return c.Delete()
	}
}

// refreshCaptcha 刷新验证码
func refreshCaptcha() func(c tb.Context) error {
	return func(c tb.Context) error {
		userIdStr := strconv.FormatInt(c.Sender().ID, 10)
		captchaCode := gUserCaptchaCodeTable.Get(userIdStr)
		if captchaCode == nil || captchaCode.UserId != c.Sender().ID {
			return nil
		}
		// 获得一个新验证码
		code, imgUrl, err := captcha.GetCaptcha()
		if err != nil {
			log.Sugar.Error(err)
			return c.Respond(&tb.CallbackResponse{
				Text: "服务器繁忙~",
			})
		}
		editMessage := &tb.Photo{
			File: tb.FromDisk(imgUrl),
			Caption: fmt.Sprintf(config.MessageC.CaptchaImage,
				captchaCode.GroupTitle,
				config.SystemC.CaptchaTimeout,
			),
		}
		_, err = Bot.Edit(c.Message(), editMessage, &tb.ReplyMarkup{InlineKeyboard: c.Message().ReplyMarkup.InlineKeyboard})
		if err != nil {
			log.Sugar.Error("[refreshCaptcha] send refreshCaptcha err:", err)
			return nil
		}
		captchaCode.Code = code
		gUserCaptchaCodeTable.Set(userIdStr, captchaCode)
		os.Remove(imgUrl)
		return c.Respond(&tb.CallbackResponse{
			Text: "验证码已刷新~",
		})
	}
}

func AddAd(c tb.Context) error {
	payload := c.Message().Payload
	payloadSlice := strings.Split(payload, "|")
	if len(payloadSlice) != 4 {
		return c.Send("消息格式错误")
	}
	title := payloadSlice[0]
	url := payloadSlice[1]
	validityPeriod := payloadSlice[2]
	sort, _ := strconv.Atoi(payloadSlice[3])
	ad := model.Advertise{
		Title:          title,
		Url:            url,
		Sort:           sort,
		ValidityPeriod: carbon.Parse(validityPeriod).Timestamp(),
		CreatedAt:      carbon.Now().Timestamp(),
	}
	err := service.AddAdvertiseService(ad)
	if err != nil {
		return c.Send("新增广告失败:" + err.Error())
	}
	c.Send("新增广告成功")
	return AllAd(c)
}

func AllAd(c tb.Context) error {
	adList, err := service.AllAdvertiseService()
	if err != nil {
		return c.Send("获取广告失败，err:" + err.Error())
	}
	table := "所有广告：\n"
	for _, advertise := range adList {
		table += fmt.Sprintf("Id:%d|Title:%s|Url:%s|Sort:%d|ValidityPeriod:%s|CreatedAt:%s \n",
			advertise.ID,
			advertise.Title,
			advertise.Url,
			advertise.Sort,
			carbon.CreateFromTimestamp(advertise.ValidityPeriod).ToDateTimeString(),
			carbon.CreateFromTimestamp(advertise.CreatedAt).ToDateTimeString(),
		)
	}
	return c.Send(table)
}

func DelAd(c tb.Context) error {
	payload := c.Message().Payload
	if payload == "" {
		return c.Send("消息格式错误")
	}
	id, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return c.Send(err.Error())
	}
	err = service.DeleteAdvertiseService(id)
	if err != nil {
		return c.Send(err.Error())
	}
	c.Send("广告删除成功！")
	return AllAd(c)
}
