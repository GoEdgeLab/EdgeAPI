package models

// ComposeValue 组合原始值
func (this *IPItem) ComposeValue() string {
	if len(this.Value) > 0 {
		return this.Value
	}

	// 兼容以往版本
	if len(this.IpTo) > 0 {
		return this.IpFrom + "-" + this.IpTo
	}

	return this.IpFrom
}
