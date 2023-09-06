package models

import timeutil "github.com/iwind/TeaGo/utils/time"

// IsExpired 判断套餐是否过期
func (this *UserPlan) IsExpired() bool {
	return len(this.DayTo) == 0 || this.DayTo < timeutil.Format("Y-m-d")
}
