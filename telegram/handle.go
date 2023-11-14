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
)

var (
	captchaMessageMenu = &tb.ReplyMarkup{ResizeKeyboard: true}
	manslaughterMenu   = &tb.ReplyMarkup{ResizeKeyboard: true}
)

var (
	TgUserIdMapToCaptchaSession sync.Map
)

// StartCaptcha 开始验证
func StartCaptcha(c tb.Context) error {
	captchaId := c.Message().Payload
	// 不是私聊或者载荷为空
	if !c.Message().Private() || captchaId == "" {
		return nil
	}
	captchaRecord, err := service.GetRecordByCaptchaId(captchaId)
	if err != nil {
		log.Sugar.Error("[StartCaptcha] GetRecordByCaptchaId err:", err)
		return c.Send("服务器异常~，请稍后再试")
	}
	if captchaRecord.ID <= 0 || captchaRecord.TelegramUserId != c.Sender().ID || captchaRecord.CaptchaStatus != model.CaptchaStatusPending {
		return c.Send("您在该群没有待验证记录，或已超时，请重新加入后验证")
	}

	// 临时会话对应的验证消息，用于后面用户输入验证码后知道是哪条消息
	TgUserIdMapToCaptchaSession.Store(c.Sender().ID, captchaId)

	// 获得一个验证码
	captchaCode, imgUrl, err := captcha.GetCaptcha()
	if err != nil {
		log.Sugar.Error("[StartCaptcha] get image captcha err:", err)
		return c.Send("服务器异常~，请稍后再试")
	}
	captchaMessage := fmt.Sprintf(config.MessageC.CaptchaImage,
		captchaRecord.TelegramChatName,
		config.SystemC.CaptchaTimeout,
	)
	sendMessage := &tb.Photo{
		File:    tb.FromDisk(imgUrl),
		Caption: captchaMessage,
	}
	refreshCaptchaImageBtn := captchaMessageMenu.Data("🔁刷新验证码", "refreshCaptchaImageBtn", captchaId)
	Bot.Handle(&refreshCaptchaImageBtn, refreshCaptcha())
	captchaMessageMenu.Inline(
		captchaMessageMenu.Row(refreshCaptchaImageBtn),
	)
	botMsg, err := Bot.Send(c.Chat(), sendMessage, captchaMessageMenu)
	if err != nil {
		log.Sugar.Error("[StartCaptcha] send image captcha err:", err)
		return c.Send("服务器异常~，请稍后再试")
	}
	err = service.SetCaptchaCodeMessageIdByCaptchaId(captchaId, botMsg.ID)
	if err != nil {
		log.Sugar.Error("[StartCaptcha] SetCaptchaCodeMessageIdByCaptchaId err:", err)
	}
	_ = os.Remove(imgUrl)
	time.AfterFunc(time.Duration(config.SystemC.CaptchaTimeout)*time.Second, func() {
		err = Bot.Delete(botMsg)
		if err != nil {
			log.Sugar.Error("[StartCaptcha] delete captcha err:", err)
		}
	})
	return service.SetCaptchaCodeByCaptchaId(captchaId, captchaCode)
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
	if err = c.Reply(blockMessage, manslaughterMenu, tb.ModeMarkdownV2); err != nil {
		log.Sugar.Error("[AdBlock] reply message err:", err)
		return err
	}
	return c.Delete()
}

// VerificationProcess 验证处理
func VerificationProcess(c tb.Context) error {
	captchaIdObj, ok := TgUserIdMapToCaptchaSession.Load(c.Sender().ID)
	if !ok {
		return nil
	}
	captchaId, ok := captchaIdObj.(string)
	if !ok {
		log.Sugar.Error("Value is not a string")
		return c.Send("服务器异常~，请稍后再试")
	}
	captchaRecord, err := service.GetRecordByCaptchaId(captchaId)
	if err != nil {
		log.Sugar.Error("[VerificationProcess] GetRecordByCaptchaId err:", err)
		return c.Send("服务器异常~，请稍后再试")
	}
	if captchaRecord.ID <= 0 || captchaRecord.TelegramUserId != c.Sender().ID || captchaRecord.CaptchaStatus != model.CaptchaStatusPending {
		return c.Send("您在该群没有待验证记录，或已超时，请重新加入后验证")
	}
	// 验证
	replyCode := c.Message().Text
	if !captcha.VerifyCaptcha(captchaRecord.CaptchaCode, replyCode) {
		return c.Send("验证码错误，请重新输入！")
	}
	// 解禁用户
	err = Bot.Restrict(&tb.Chat{ID: captchaRecord.TelegramChatId}, &tb.ChatMember{
		User:   &tb.User{ID: captchaRecord.TelegramUserId},
		Rights: tb.NoRestrictions(),
	})
	if err != nil {
		log.Sugar.Error("[OnTextMessage] unban err:", err)
		return c.Send("服务器异常~，请稍后重试~")
	}
	err = service.SuccessRecordByCaptchaId(captchaId)
	if err != nil {
		log.Sugar.Error("[OnTextMessage] SuccessRecordByCaptchaId err:", err)
	}

	// 删除群内的验证消息
	Bot.Delete(&tb.StoredMessage{MessageID: strconv.Itoa(captchaRecord.CaptchaMessageId), ChatID: captchaRecord.TelegramChatId})
	// 删除验证码消息
	Bot.Delete(&tb.StoredMessage{MessageID: strconv.Itoa(captchaRecord.CaptchaCodeMessageId), ChatID: c.Message().Chat.ID})
	return c.Send(config.MessageC.VerificationComplete)
}

// UserJoinGroup 用户加群事件
func UserJoinGroup(c tb.Context) error {
	var err error
	// 如果是管理员邀请的，直接通过
	if isManage(c.Message().Chat, c.Sender().ID) {
		return nil
	}

	// ban user
	err = Bot.Restrict(c.Message().Chat, &tb.ChatMember{
		Rights:          tb.NoRights(),
		User:            c.Message().UserJoined,
		RestrictedUntil: tb.Forever(),
	})
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] ban user err:", err)
		return err
	}

	userLink := fmt.Sprintf("tg://user?id=%d", c.Message().UserJoined.ID)
	joinMessage := fmt.Sprintf(config.MessageC.JoinHint,
		c.Message().UserJoined.LastName+c.Message().UserJoined.FirstName,
		userLink,
		c.Message().Chat.Title,
		config.SystemC.JoinHintAfterDelTime)
	captchaId := uuid.NewV4().String()
	doCaptchaBtn := joinMessageMenu.URL("👉🏻点我开始人机验证🤖", fmt.Sprintf("https://t.me/%s?start=%s", Bot.Me.Username, captchaId))
	var (
		manageBanBtn  = joinMessageMenu.Data("👮‍管理员禁止🈲", "manageBanBtn", captchaId)
		managePassBtn = joinMessageMenu.Data("👮‍管理员通过✅", "managePassBtn", captchaId)
	)
	// 按钮点击事件
	Bot.Handle(&manageBanBtn, ManageBan(), isManageMiddleware)
	Bot.Handle(&managePassBtn, ManagePass(), isManageMiddleware)
	joinMessageMenu.Inline(
		joinMessageMenu.Row(doCaptchaBtn),
		joinMessageMenu.Row(manageBanBtn, managePassBtn),
	)
	LoadAdMenuBtn(joinMessageMenu)
	captchaMessage, err := Bot.Send(c.Message().Chat, joinMessage, joinMessageMenu, tb.ModeMarkdownV2)
	if err != nil {
		log.Sugar.Error("[UserJoinGroup] send join hint message err:", err)
		return err
	}
	defer func() {
		time.AfterFunc(time.Duration(config.SystemC.JoinHintAfterDelTime)*time.Second, func() {
			if err = Bot.Delete(captchaMessage); err != nil {
				log.Sugar.Warn("[UserJoinGroup] delete join hint message err:", err)
			}
		})
	}()

	record := &model.UserCaptchaRecord{
		CaptchaId:             captchaId,
		TelegramChatName:      c.Message().Chat.Title,
		TelegramUserLastName:  c.Message().UserJoined.LastName,
		TelegramUserFirstName: c.Message().UserJoined.FirstName,
		TelegramUserId:        c.Message().UserJoined.ID,
		TelegramChatId:        c.Message().Chat.ID,
		CaptchaMessageId:      captchaMessage.ID,
		CaptchaStatus:         model.CaptchaStatusPending,
		CaptchaTimeoutEndTime: carbon.DateTime{Carbon: carbon.Now().AddSeconds(config.SystemC.CaptchaTimeout)},
	}
	err = service.CreateCaptchaRecord(record)
	return err
}

// ManageBan 管理员手动禁止
func ManageBan() func(c tb.Context) error {
	return func(c tb.Context) error {
		defer func() {
			c.Delete()
		}()
		captchaId := c.Data()
		return service.TimeoutRecordByCaptchaId(captchaId)
	}
}

// ManagePass 管理员手动通过
func ManagePass() func(c tb.Context) error {
	return func(c tb.Context) error {
		defer func() {
			c.Delete()
		}()
		captchaId := c.Data()
		return service.SuccessRecordByCaptchaId(captchaId)
	}
}

// refreshCaptcha 刷新验证码
func refreshCaptcha() func(c tb.Context) error {
	return func(c tb.Context) error {
		captchaId := c.Data()
		captchaRecord, err := service.GetRecordByCaptchaId(captchaId)
		if err != nil {
			log.Sugar.Error("[refreshCaptcha] GetRecordByCaptchaId err:", err)
			return c.Respond(&tb.CallbackResponse{
				Text: "服务器繁忙~",
			})
		}
		if captchaRecord.ID <= 0 || captchaRecord.TelegramUserId != c.Sender().ID || captchaRecord.CaptchaStatus != model.CaptchaStatusPending {
			return c.Respond(&tb.CallbackResponse{
				Text: "您在该群没有待验证记录，或已超时，请重新加入后验证",
			})
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
				captchaRecord.TelegramChatName,
				config.SystemC.CaptchaTimeout,
			),
		}
		_, err = Bot.Edit(c.Message(), editMessage, &tb.ReplyMarkup{InlineKeyboard: c.Message().ReplyMarkup.InlineKeyboard})
		if err != nil {
			log.Sugar.Error("[refreshCaptcha] send refreshCaptcha err:", err)
			return nil
		}
		_ = os.Remove(imgUrl)
		err = service.SetCaptchaCodeByCaptchaId(captchaId, code)
		if err != nil {
			log.Sugar.Error(err)
			return c.Respond(&tb.CallbackResponse{
				Text: "服务器繁忙~",
			})
		}
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
