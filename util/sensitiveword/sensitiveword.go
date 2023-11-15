package sensitiveword

import (
	"github.com/assimon/captcha-bot/util/config"
	"github.com/importcjj/sensitive"
	"log"
	"os"
	"strings"
)

var Filter *sensitive.Filter

// InitSensitiveWord 加载敏感词库
func InitSensitiveWord() {
	sensitiveWordPath := config.AppPath + "/dict/"
	Filter = sensitive.New()
	files, err := os.ReadDir(sensitiveWordPath)
	if err != nil {
		log.Fatalln("[InitSensitiveWord] load dict err:", err)
	}
	for _, file := range files {
		// 文件名必须是已解密文件
		if !strings.Contains(file.Name(), "dec_") {
			continue
		}
		sensitiveFile := sensitiveWordPath + file.Name()
		err = Filter.LoadWordDict(sensitiveFile)
		if err != nil {
			log.Fatalln("[InitSensitiveWord] load sensitive file err:", err, ", file:", sensitiveFile)
		}
	}
}
