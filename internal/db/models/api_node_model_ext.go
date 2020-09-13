package models

import "strconv"

// 地址
func (this *APINode) Address() string {
	return this.Host + ":" + strconv.Itoa(int(this.Port))
}
