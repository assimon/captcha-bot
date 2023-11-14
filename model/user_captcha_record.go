package model

import "github.com/golang-module/carbon/v2"

const (
	CaptchaStatusPending = -1
	CaptchaStatusSuccess = 1
	CaptchaStatusTimeout = 2
)

type UserCaptchaRecord struct {
	BaseModel
	CaptchaId               string           `gorm:"column:captcha_id;unique_index:captcha_id_index" json:"captcha_id"` //验证id
	TelegramChatName        string           `gorm:"column:telegram_chat_name" json:"telegram_chat_name"`
	TelegramUserLastName    string           `gorm:"column:telegram_user_last_name" json:"telegram_user_last_name"`
	TelegramUserFirstName   string           `gorm:"column:telegram_user_first_name" json:"telegram_user_first_name"`
	TelegramUserId          int64            `gorm:"column:telegram_user_id" json:"telegram_user_id"`
	TelegramChatId          int64            `gorm:"column:telegram_chat_id" json:"telegram_chat_id"`
	CaptchaMessageId        int              `gorm:"column:captcha_message_id" json:"captcha_message_id"`                 // 群-待验证消息id
	CaptchaStatus           int              `gorm:"column:captcha_status" json:"captcha_status"`                         // 验证状态 -1待验证 1已验证 2已超时
	CaptchaTimeoutMessageId int              `gorm:"column:captcha_timeout_message_id" json:"captcha_timeout_message_id"` // 群-验证超时消息id
	CaptchaTimeoutTime      carbon.Timestamp `gorm:"column:captcha_timeout_time"  json:"captcha_timeout_time"`            // 验证超时时间
	CaptchaSuccessMessageId int              `gorm:"column:captcha_success_message_id" json:"captcha_success_message_id"` // 群-验证成功的消息id
	CaptchaSuccessTime      carbon.Timestamp `gorm:"column:captcha_success_time" json:"captcha_success_time"`             // 验证成功时间

	CaptchaCodeMessageId int    `gorm:"column:captcha_code_message_id"  json:"captcha_code_message_id"` // 验证码消息id
	CaptchaCode          string `gorm:"column:captcha_code" json:"captcha_code"`                        // 验证码内容

}

func (UserCaptchaRecord) TableName() string {
	return "user_captcha_record"
}
