package utils

import (
	"github.com/iwind/TeaGo/logs"
)

// 打印错误
func PrintError(err error) {
	// TODO 记录调用的文件名、行数
	logs.Println("[ERROR]" + err.Error())
}
