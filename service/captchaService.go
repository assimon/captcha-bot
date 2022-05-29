package service

import (
	tb "gopkg.in/telebot.v3"
	"sync"
)

type CaptchaPending struct {
	UserId         int64       `json:"user_id"`
	GroupId        int64       `json:"group_id"`
	JoinAt         int64       `json:"join_at"`
	PendingMessage *tb.Message `json:"pending_message"`
}

type CaptchaPendingTable struct {
	UserCaptchaPending map[string]*CaptchaPending
	sync.RWMutex
}

func NewCaptchaPendingTable() *CaptchaPendingTable {
	return &CaptchaPendingTable{
		UserCaptchaPending: make(map[string]*CaptchaPending),
	}
}

func (t *CaptchaPendingTable) Set(key string, val *CaptchaPending) {
	t.Lock()
	defer t.Unlock()
	t.UserCaptchaPending[key] = val
}

func (t *CaptchaPendingTable) Get(key string) *CaptchaPending {
	t.RLock()
	defer t.RUnlock()
	val := t.UserCaptchaPending[key]
	return val
}

func (t *CaptchaPendingTable) Del(key string) {
	t.Lock()
	defer t.Unlock()
	delete(t.UserCaptchaPending, key)
}

// CaptchaCode 验证码
type CaptchaCode struct {
	UserId         int64       `json:"user_id"`
	GroupId        int64       `json:"group_id"`
	Code           string      `json:"code"`
	CaptchaMessage *tb.Message `json:"message_id"`
	PendingMessage *tb.Message `json:"pending_message"`
	GroupTitle     string      `json:"group_title"`
	CreatedAt      int64       `json:"created_at"`
}

type CaptchaCodeTable struct {
	UserCaptchaCode map[string]*CaptchaCode
	sync.RWMutex
}

func NewCaptchaCodeTable() *CaptchaCodeTable {
	return &CaptchaCodeTable{
		UserCaptchaCode: make(map[string]*CaptchaCode),
	}
}

func (t *CaptchaCodeTable) Set(key string, val *CaptchaCode) {
	t.Lock()
	defer t.Unlock()
	t.UserCaptchaCode[key] = val
}

func (t *CaptchaCodeTable) Get(key string) *CaptchaCode {
	t.RLock()
	defer t.RUnlock()
	val := t.UserCaptchaCode[key]
	return val
}

func (t *CaptchaCodeTable) Del(key string) {
	t.Lock()
	defer t.Unlock()
	delete(t.UserCaptchaCode, key)
}
