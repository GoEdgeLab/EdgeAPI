package errors

import (
	"errors"
	"github.com/iwind/TeaGo/Tea"
	"path/filepath"
	"runtime"
	"strconv"
)

type errorObj struct {
	err      error
	file     string
	line     int
	funcName string
}

func (this *errorObj) Error() string {
	// 在非测试环境下，我们不提示详细的行数等信息
	if !Tea.IsTesting() {
		return this.err.Error()
	}

	s := this.err.Error() + "\n  " + this.file
	if len(this.funcName) > 0 {
		s += ":" + this.funcName + "()"
	}
	s += ":" + strconv.Itoa(this.line)
	return s
}

// 新错误
func New(errText string) error {
	ptr, file, line, ok := runtime.Caller(1)
	funcName := ""
	if ok {
		frame, _ := runtime.CallersFrames([]uintptr{ptr}).Next()
		funcName = filepath.Base(frame.Function)
	}
	return &errorObj{
		err:      errors.New(errText),
		file:     file,
		line:     line,
		funcName: funcName,
	}
}

// 包装已有错误
func Wrap(err error) error {
	if err == nil {
		return nil
	}

	ptr, file, line, ok := runtime.Caller(1)
	funcName := ""
	if ok {
		frame, _ := runtime.CallersFrames([]uintptr{ptr}).Next()
		funcName = filepath.Base(frame.Function)
	}
	return &errorObj{
		err:      err,
		file:     file,
		line:     line,
		funcName: funcName,
	}
}
