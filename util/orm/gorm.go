package orm

import (
	"github.com/assimon/captcha-bot/model"
	"github.com/assimon/captcha-bot/util/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var Gdb *gorm.DB

func InitDb() {
	db, err := gorm.Open(sqlite.Open(config.AppPath+"/db/geecaptcha.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("open datebase err:", err)
	}
	db.AutoMigrate(model.Advertise{})
	database, _ := db.DB()
	database.SetMaxOpenConns(1)
	err = database.Ping()
	if err != nil {
		log.Fatal("database ping err:", err)
	}
	Gdb = db
}
