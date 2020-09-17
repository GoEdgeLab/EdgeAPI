package models

// 处理JSON字节Slice
func JSONBytes(data []byte) []byte {
	if len(data) == 0 {
		return []byte("null")
	}
	return data
}

// 判断JSON是否为空
func IsNotNull(data string) bool {
	return len(data) > 0 && data != "null"
}
