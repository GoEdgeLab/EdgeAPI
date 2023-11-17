// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils

import "regexp"

var mobileRegex = regexp.MustCompile(`^1\d{10}$`)

// IsValidMobile validate mobile number
func IsValidMobile(mobile string) bool {
	return mobileRegex.MatchString(mobile)
}
