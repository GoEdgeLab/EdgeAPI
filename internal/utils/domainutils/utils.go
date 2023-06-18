// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package domainutils

import (
	"regexp"
	"strings"
)

// ValidateDomainFormat 校验域名格式
func ValidateDomainFormat(domain string) bool {
	pieces := strings.Split(domain, ".")
	for _, piece := range pieces {
		if piece == "-" ||
			strings.HasPrefix(piece, "-") ||
			strings.HasSuffix(piece, "-") ||
			//strings.Contains(piece, "--") ||
			len(piece) > 63 ||
			// 支持中文、大写字母、下划线
			!regexp.MustCompile(`^[\p{Han}_a-zA-Z0-9-]+$`).MatchString(piece) {
			return false
		}
	}

	// 最后一段不能是全数字
	if regexp.MustCompile(`^(\d+)$`).MatchString(pieces[len(pieces)-1]) {
		return false
	}

	return true
}
