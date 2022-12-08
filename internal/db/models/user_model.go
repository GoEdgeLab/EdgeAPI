package models

import "github.com/iwind/TeaGo/dbs"

// User 用户
type User struct {
	Id              uint32   `field:"id"`              // ID
	IsOn            bool     `field:"isOn"`            // 是否启用
	Username        string   `field:"username"`        // 用户名
	Password        string   `field:"password"`        // 密码
	Fullname        string   `field:"fullname"`        // 真实姓名
	Mobile          string   `field:"mobile"`          // 手机号
	Tel             string   `field:"tel"`             // 联系电话
	Remark          string   `field:"remark"`          // 备注
	Email           string   `field:"email"`           // 邮箱地址
	VerifiedEmail   string   `field:"verifiedEmail"`   // 激活后的邮箱
	EmailIsVerified uint8    `field:"emailIsVerified"` // 邮箱是否已验证
	AvatarFileId    uint64   `field:"avatarFileId"`    // 头像文件ID
	CreatedAt       uint64   `field:"createdAt"`       // 创建时间
	Day             string   `field:"day"`             // YYYYMMDD
	UpdatedAt       uint64   `field:"updatedAt"`       // 修改时间
	State           uint8    `field:"state"`           // 状态
	Source          string   `field:"source"`          // 来源
	ClusterId       uint32   `field:"clusterId"`       // 集群ID
	Features        dbs.JSON `field:"features"`        // 允许操作的特征
	RegisteredIP    string   `field:"registeredIP"`    // 注册使用的IP
	IsRejected      bool     `field:"isRejected"`      // 是否已拒绝
	RejectReason    string   `field:"rejectReason"`    // 拒绝理由
	IsVerified      bool     `field:"isVerified"`      // 是否验证通过
	RequirePlans    uint8    `field:"requirePlans"`    // 是否需要购买套餐
	Modules         dbs.JSON `field:"modules"`         // 用户模块
	PriceType       string   `field:"priceType"`       // 计费类型：traffic|bandwidth
	PricePeriod     string   `field:"pricePeriod"`     // 结算周期
	ServersEnabled  uint8    `field:"serversEnabled"`  // 是否禁用所有服务
}

type UserOperator struct {
	Id              any // ID
	IsOn            any // 是否启用
	Username        any // 用户名
	Password        any // 密码
	Fullname        any // 真实姓名
	Mobile          any // 手机号
	Tel             any // 联系电话
	Remark          any // 备注
	Email           any // 邮箱地址
	VerifiedEmail   any // 激活后的邮箱
	EmailIsVerified any // 邮箱是否已验证
	AvatarFileId    any // 头像文件ID
	CreatedAt       any // 创建时间
	Day             any // YYYYMMDD
	UpdatedAt       any // 修改时间
	State           any // 状态
	Source          any // 来源
	ClusterId       any // 集群ID
	Features        any // 允许操作的特征
	RegisteredIP    any // 注册使用的IP
	IsRejected      any // 是否已拒绝
	RejectReason    any // 拒绝理由
	IsVerified      any // 是否验证通过
	RequirePlans    any // 是否需要购买套餐
	Modules         any // 用户模块
	PriceType       any // 计费类型：traffic|bandwidth
	PricePeriod     any // 结算周期
	ServersEnabled  any // 是否禁用所有服务
}

func NewUserOperator() *UserOperator {
	return &UserOperator{}
}
