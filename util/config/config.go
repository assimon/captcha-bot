package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

var AppPath string

type System struct {
	JoinHintAfterDelTime int    `mapstructure:"join_hint_after_del_time"`
	CaptchaTimeout       int    `mapstructure:"captcha_timeout"`
	RuntimePath          string `mapstructure:"runtime_path"`
}

var SystemC System

type Telegram struct {
	BotToken    string  `mapstructure:"bot_token"`
	ApiProxy    string  `mapstructure:"api_proxy"`
	ManageUsers []int64 `mapstructure:"manage_users"`
}

var TelegramC Telegram

type Log struct {
	MaxSize    int `mapstructure:"max_size"`
	MaxAge     int `mapstructure:"max_age"`
	MaxBackups int `mapstructure:"max_backups"`
}

var LogC Log

type Message struct {
	JoinHint             string `mapstructure:"join_hint"`
	CaptchaImage         string `mapstructure:"captcha_image"`
	VerificationComplete string `mapstructure:"verification_complete"`
	BlockHint            string `mapstructure:"block_hint"`
}

var MessageC Message

type AdBlock struct {
	NumberOfForbiddenWords int   `mapstructure:"number_of_forbidden_words"`
	BlockTime              int64 `mapstructure:"block_time"`
}

var AdBlockC AdBlock

// InitConfig 配置加载
func InitConfig() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	AppPath = path
	viper.SetConfigFile(path + "/config/config.toml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal("load config file err:", err)
	}
	err = viper.UnmarshalKey("system", &SystemC)
	if err != nil {
		log.Fatal("load config system err:", err)
	}
	err = viper.UnmarshalKey("telegram", &TelegramC)
	if err != nil {
		log.Fatal("load config telegram err:", err)
	}
	err = viper.UnmarshalKey("log", &LogC)
	if err != nil {
		log.Fatal("load config log err:", err)
	}
	err = viper.UnmarshalKey("message", &MessageC)
	if err != nil {
		log.Fatal("load config message err:", err)
	}
	err = viper.UnmarshalKey("adblock", &AdBlockC)
	if err != nil {
		log.Fatal("load config adblock err:", err)
	}
}
