package telegram

import (
	"fmt"
	"github.com/assimon/captcha-bot/model"
	"github.com/assimon/captcha-bot/service"
	"github.com/assimon/captcha-bot/util/captcha"
	"github.com/assimon/captcha-bot/util/config"
	"github.com/assimon/captcha-bot/util/log"
	"github.com/assimon/captcha-bot/util/sensitiveword"
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
	manslaughterMenu   = &tb.ReplyMarkup{ResizeKeyboard: true}
)

var (
	gUserCaptchaCodeTable    = service.NewCaptchaCodeTable()
	gUserCaptchaPendingTable = service.NewCaptchaPendingTable()
)

var (
	gMessageTokenMap                 sync.Map
	gUserIdToJoinCaptchaMessageIdMap sync.Map
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
		_ = os.Remove(imgUrl)
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
	// 私聊走入群验证操作
	if c.Message().Private() {
		return VerificationProcess(c)
	}
	// 否则走广告阻止监听
	return AdBlock(c)
}

// AdBlock 广告阻止
func AdBlock(c tb.Context) error {
	userId := c.Message().Sender.ID
	userLink := fmt.Sprintf("tg://user?id=%d", c.Message().Sender.ID)
	userNickname := c.Message().Sender.LastName + c.Message().Sender.FirstName
	messageText := c.Message().Text
	// 管理员 放行任何操作
	if isManage(c.Chat(), userId) {
		return nil
	}
	dict := sensitiveword.Filter.FindAll(messageText)
	if len(dict) <= 0 || len(dict) < config.AdBlockC.NumberOfForbiddenWords {
		return nil
	}
	// ban user
	restrictedUntil := config.AdBlockC.BlockTime
	if restrictedUntil <= 0 {
		restrictedUntil = tb.Forever()
	}
	err := Bot.Restrict(c.Chat(), &tb.ChatMember{
		Rights:          tb.NoRights(),
		User:            c.Message().Sender,
		RestrictedUntil: restrictedUntil,
	})
	if err != nil {
		log.Sugar.Error("[AdBlock] ban user err:", err)
		return err
	}
	blockMessage := fmt.Sprintf(config.MessageC.BlockHint,
		userNickname,
		userLink,
		strings.Join(dict, ","))
	manslaughterBtn := manslaughterMenu.Data("👮🏻管理员解封", strconv.FormatInt(userId, 10))
	manslaughterMenu.Inline(manslaughterMenu.Row(manslaughterBtn))
	LoadAdMenuBtn(manslaughterMenu)
	Bot.Handle(&manslaughterBtn, func(c tb.Context) error {
		if err = Bot.Delete(c.Message()); err != nil {
			log.Sugar.Error("[AdBlock] delete adblock message err:", err)
			return err
		}
		// 解禁用户
		err = Bot.Restrict(c.Chat(), &tb.ChatMember{
			User:   &tb.User{ID: userId},
			Rights: tb.NoRestrictions(),
		})
		if err != nil {
			log.Sugar.Error("[AdBlock] unban user err:", err)
			return err
		}
		return c.Send(fmt.Sprintf("管理员已解除对用户：[%s](%s) 的封禁", userNickname, userLink), tb.ModeMarkdownV2)
	}, isManageMiddleware)
	//删除验证消息
	go func() {
		msgObj, ok := gUserIdToJoinCaptchaMessageIdMap.Load(userId)
		if !ok {
			return
		}
		delPendingMsg, ok := msgObj.(*tb.Message)
		if !ok {
			return
		}
		if err = Bot.Delete(delPendingMsg); err != nil {
			log.Sugar.Error("[AdBlock] delete captcha message err:", err)
		}
	}()
	if err = c.Reply(blockMessage, manslaughterMenu, tb.ModeMarkdownV2); err != nil {
		log.Sugar.Error("[AdBlock] reply message err:", err)
		return err
	}
	return c.Delete()
}

// VerificationProcess 验证处理
func VerificationProcess(c tb.Context) error {
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
	if err = Bot.Delete(captchaCode.CaptchaMessage); err != nil {
		log.Sugar.Error("[OnTextMessage] delete captcha message err:", err)
	}
	if err = Bot.Delete(captchaCode.PendingMessage); err != nil {
		log.Sugar.Error("[OnTextMessage] delete pending message err:", err)
	}
	return c.Send(config.MessageC.VerificationComplete)
}

// UserJoinGroup 用户加群事件
func UserJoinGroup(c tb.Context) error {
	var err error
	// 如果是管理员邀请的，直接通过
	if isManage(c.ChatMember().Chat, c.ChatMember().Sender.ID) {
		return nil
	}
	// ban user
	err = Bot.Restrict(c.ChatMember().Chat, &tb.ChatMember{
		Rights:          tb.NoRights(),
		User:            c.ChatMember().Sender,
		RestrictedUntil: tb.Forever(),
	})
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] ban user err:", err)
		return err
	}
	userLink := fmt.Sprintf("tg://user?id=%d", c.ChatMember().Sender.ID)
	joinMessage := fmt.Sprintf(config.MessageC.JoinHint,
		c.ChatMember().Sender.LastName+c.ChatMember().Sender.FirstName,
		userLink,
		c.ChatMember().Chat.Title,
		config.SystemC.JoinHintAfterDelTime)
	chatToken := uuid.NewV4().String()
	doCaptchaBtn := joinMessageMenu.URL("👉🏻点我开始人机验证🤖", fmt.Sprintf("https://t.me/%s?start=%s", Bot.Me.Username, chatToken))
	joinMessageMenu.Inline(
		joinMessageMenu.Row(doCaptchaBtn),
		joinMessageMenu.Row(manageBanBtn, managePassBtn),
	)
	LoadAdMenuBtn(joinMessageMenu)
	captchaMessage, err := Bot.Send(c.ChatMember().Chat, joinMessage, joinMessageMenu, tb.ModeMarkdownV2)
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] send join hint message err:", err)
		return err
	}
	// 设置token对于验证消息
	gMessageTokenMap.Store(chatToken, fmt.Sprintf("%d|%d|%s", captchaMessage.ID, c.ChatMember().Chat.ID, c.ChatMember().Chat.Title))
	gUserIdToJoinCaptchaMessageIdMap.Store(c.ChatMember().Sender.ID, captchaMessage)
	captchaDataVal := &service.CaptchaPending{
		PendingMessage: captchaMessage,
		UserId:         c.ChatMember().Sender.ID,
		GroupId:        c.ChatMember().Chat.ID,
		JoinAt:         carbon.Now().Timestamp(),
	}
	captchaDataKey := fmt.Sprintf("%d|%d", captchaMessage.ID, c.ChatMember().Chat.ID)
	gUserCaptchaPendingTable.Set(captchaDataKey, captchaDataVal)
	time.AfterFunc(time.Duration(config.SystemC.JoinHintAfterDelTime)*time.Second, func() {
		if err = Bot.Delete(captchaMessage); err != nil {
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
		_ = os.Remove(imgUrl)
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
	if err = c.Send("新增广告成功"); err != nil {
		log.Sugar.Error("[AddAd] send success message err:", err)
	}
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
	if err = service.DeleteAdvertiseService(id); err != nil {
		return c.Send(err.Error())
	}
	if err = c.Send("广告删除成功！"); err != nil {
		log.Sugar.Error("[DelAd] send success message err:", err)
	}
	return AllAd(c)
}

// LoadAdMenuBtn 加载广告
func LoadAdMenuBtn(menu *tb.ReplyMarkup) {
	advertises, err := service.GetEfficientAdvertiseService()
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] load advertise err:", err)
	} else {
		for _, advertise := range advertises {
			menu.InlineKeyboard = append(menu.InlineKeyboard, []tb.InlineButton{
				{
					Text: advertise.Title,
					URL:  advertise.Url,
				},
			})
		}
	}
}
