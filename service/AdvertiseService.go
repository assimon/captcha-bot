package service

import (
	"github.com/assimon/captcha-bot/model"
	"github.com/assimon/captcha-bot/util/orm"
	"github.com/golang-module/carbon/v2"
)

// AddAdvertiseService 新增广告
func AddAdvertiseService(advertise model.Advertise) (err error) {
	return orm.Gdb.Model(&advertise).Create(&advertise).Error
}

// AllAdvertiseService 加载所有广告
func AllAdvertiseService() (advertises []model.Advertise, err error) {
	err = orm.Gdb.Model(&advertises).Find(&advertises).Error
	return
}

// GetEfficientAdvertiseService 加载正在生效的广告
func GetEfficientAdvertiseService() (advertises []model.Advertise, err error) {
	err = orm.Gdb.Model(&advertises).Where("validity_period > ?", carbon.Now().Timestamp()).Order("sort desc").Find(&advertises).Error
	return
}

// DeleteAdvertiseService 删除一条广告
func DeleteAdvertiseService(id int64) (err error) {
	return orm.Gdb.Where("id = ?", id).Delete(&model.Advertise{}).Error
}
