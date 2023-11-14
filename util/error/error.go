package error

import "fmt"

// MustPanicErrorFunc 任何可能触发panic的方法丢到这里面执行
func MustPanicErrorFunc(fn func()) (err error) {
	defer func() {
		p := recover()
		if p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	fn()
	return err
}
