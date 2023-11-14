package service

import (
	"github.com/assimon/captcha-bot/model"
	"github.com/assimon/captcha-bot/util/orm"
	"github.com/golang-module/carbon/v2"
)

// CreateCaptchaRecord 创建验证记录
func CreateCaptchaRecord(record *model.UserCaptchaRecord) error {
	err := orm.Gdb.Model(&model.UserCaptchaRecord{}).Create(record).Error
	return err
}

// GetRecordByCaptchaId 通过载荷唯一标识获取记录
func GetRecordByCaptchaId(cId string) (record *model.UserCaptchaRecord, err error) {
	err = orm.Gdb.Model(&model.UserCaptchaRecord{}).Where("captcha_id = ?", cId).Find(&record).Error
	return
}

// TimeoutRecordByCaptchaId 设置某条验证消息已经超时
func TimeoutRecordByCaptchaId(cId string) error {
	err := orm.Gdb.Model(&model.UserCaptchaRecord{}).Where("captcha_id = ?", cId).
		UpdateColumns(map[string]interface{}{
			"captcha_status":       model.CaptchaStatusTimeout,
			"captcha_timeout_time": carbon.Now().ToDateTimeString(),
		}).Error
	return err
}

// SuccessRecordByCaptchaId 设置某条消息已经验证成功
func SuccessRecordByCaptchaId(cId string) error {
	err := orm.Gdb.Model(&model.UserCaptchaRecord{}).Where("captcha_id = ?", cId).
		UpdateColumns(map[string]interface{}{
			"captcha_status":       model.CaptchaStatusSuccess,
			"captcha_success_time": carbon.Now().ToDateTimeString(),
		}).Error
	return err
}

// SetCaptchaCodeByCaptchaId 设置或刷新一个验证消息的code
func SetCaptchaCodeByCaptchaId(cId, code string) error {
	err := orm.Gdb.Model(&model.UserCaptchaRecord{}).Where("captcha_id = ?", cId).
		Update("captcha_code", code).Error
	return err
}

// SetCaptchaCodeMessageIdByCaptchaId 设置私聊验证消息id
func SetCaptchaCodeMessageIdByCaptchaId(cId string, msgId int) error {
	err := orm.Gdb.Model(&model.UserCaptchaRecord{}).Where("captcha_id = ?", cId).
		Update("captcha_code_message_id", msgId).Error
	return err
}

// GetTimeoutCaptchaRecords 获取已经超时的待验证记录
func GetTimeoutCaptchaRecords() (records []model.UserCaptchaRecord, err error) {
	err = orm.Gdb.Model(&model.UserCaptchaRecord{}).Where("captcha_status = ?", model.CaptchaStatusPending).
		Where("captcha_timeout_end_time < ?", carbon.Now().ToDateTimeString()).Find(&records).Error
	return
}
