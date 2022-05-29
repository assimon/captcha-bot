package captcha

import (
	"encoding/base64"
	"fmt"
	"github.com/assimon/captcha-bot/util/config"
	"github.com/mojocn/base64Captcha"
	"io/ioutil"
	"os"
)

var (
	store = base64Captcha.DefaultMemStore
)

// GetCaptcha 写入验证码，并且返回验证码id
func GetCaptcha() (string, string, error) {
	var err error
	imagePath := fmt.Sprintf("%s%s%s", config.AppPath, config.SystemC.RuntimePath, "/captcha_images/")
	if _, err = os.Stat(imagePath); os.IsNotExist(err) {
		mkdirErr := os.MkdirAll(imagePath, os.ModePerm)
		if mkdirErr != nil {
			return "", "", mkdirErr
		}
	}
	// 生成默认数字
	driver := base64Captcha.NewDriverDigit(100, 320, 6, 0.7, 80)
	// 生成base64图片
	c := base64Captcha.NewCaptcha(driver, store)
	// 获取
	code, b64s, err := c.Generate()
	if err != nil {
		return "", "", err
	}
	imageUrl := fmt.Sprintf("%s%s.png", imagePath, code)
	b64s = b64s[22:]
	b64img, err := base64.StdEncoding.DecodeString(b64s) //成图片文件并把文件写入到buffer
	if err != nil {
		return "", "", err
	}
	err = ioutil.WriteFile(imageUrl, b64img, 0666)
	return code, imageUrl, err
}

// VerifyCaptcha 验证验证码是否正确
func VerifyCaptcha(id, digits string) bool {
	if id == "" || digits == "" {
		return false
	}
	verifyRes := store.Verify(id, digits, false)
	if verifyRes {
		store.Verify(id, digits, true)
		return verifyRes
	} else {
		return false
	}
}
