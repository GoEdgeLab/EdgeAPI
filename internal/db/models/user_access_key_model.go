package models

// UserAccessKey AccessKey
type UserAccessKey struct {
	Id          uint32 `field:"id"`          // ID
	UserId      uint32 `field:"userId"`      // 用户ID
	SubUserId   uint32 `field:"subUserId"`   // 子用户ID
	IsOn        uint8  `field:"isOn"`        // 是否启用
	UniqueId    string `field:"uniqueId"`    // 唯一的Key
	Secret      string `field:"secret"`      // 密钥
	Description string `field:"description"` // 备注
	AccessedAt  uint64 `field:"accessedAt"`  // 最近一次访问时间
	State       uint8  `field:"state"`       // 状态
}

type UserAccessKeyOperator struct {
	Id          interface{} // ID
	UserId      interface{} // 用户ID
	SubUserId   interface{} // 子用户ID
	IsOn        interface{} // 是否启用
	UniqueId    interface{} // 唯一的Key
	Secret      interface{} // 密钥
	Description interface{} // 备注
	AccessedAt  interface{} // 最近一次访问时间
	State       interface{} // 状态
}

func NewUserAccessKeyOperator() *UserAccessKeyOperator {
	return &UserAccessKeyOperator{}
}
