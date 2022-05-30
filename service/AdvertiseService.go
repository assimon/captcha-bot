package service

import (
	"github.com/assimon/captcha-bot/model"
	"github.com/assimon/captcha-bot/util/orm"
	"github.com/golang-module/carbon/v2"
)

func AddAdvertiseService(advertise model.Advertise) (err error) {
	return orm.Gdb.Model(&advertise).Create(&advertise).Error
}

func AllAdvertiseService() (advertises []model.Advertise, err error) {
	err = orm.Gdb.Model(&advertises).Find(&advertises).Error
	return
}

func GetEfficientAdvertiseService() (advertises []model.Advertise, err error) {
	nowTime := carbon.Now().Timestamp()
	err = orm.Gdb.Model(&advertises).Where("validity_period > ?", nowTime).Order("sort desc").Find(&advertises).Error
	return
}

func DeleteAdvertiseService(id int64) (err error) {
	return orm.Gdb.Where("id = ?", id).Delete(&model.Advertise{}).Error
}
