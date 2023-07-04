package models

import "github.com/iwind/TeaGo/dbs"

const (
	UserField_Id                dbs.FieldName = "id"                // ID
	UserField_IsOn              dbs.FieldName = "isOn"              // 是否启用
	UserField_Username          dbs.FieldName = "username"          // 用户名
	UserField_Password          dbs.FieldName = "password"          // 密码
	UserField_Fullname          dbs.FieldName = "fullname"          // 真实姓名
	UserField_Mobile            dbs.FieldName = "mobile"            // 手机号
	UserField_VerifiedMobile    dbs.FieldName = "verifiedMobile"    // 已验证手机号
	UserField_Tel               dbs.FieldName = "tel"               // 联系电话
	UserField_Remark            dbs.FieldName = "remark"            // 备注
	UserField_Email             dbs.FieldName = "email"             // 邮箱地址
	UserField_VerifiedEmail     dbs.FieldName = "verifiedEmail"     // 激活后的邮箱
	UserField_EmailIsVerified   dbs.FieldName = "emailIsVerified"   // 邮箱是否已验证
	UserField_AvatarFileId      dbs.FieldName = "avatarFileId"      // 头像文件ID
	UserField_CreatedAt         dbs.FieldName = "createdAt"         // 创建时间
	UserField_Day               dbs.FieldName = "day"               // YYYYMMDD
	UserField_UpdatedAt         dbs.FieldName = "updatedAt"         // 修改时间
	UserField_State             dbs.FieldName = "state"             // 状态
	UserField_Source            dbs.FieldName = "source"            // 来源
	UserField_ClusterId         dbs.FieldName = "clusterId"         // 集群ID
	UserField_Features          dbs.FieldName = "features"          // 允许操作的特征
	UserField_RegisteredIP      dbs.FieldName = "registeredIP"      // 注册使用的IP
	UserField_IsRejected        dbs.FieldName = "isRejected"        // 是否已拒绝
	UserField_RejectReason      dbs.FieldName = "rejectReason"      // 拒绝理由
	UserField_IsVerified        dbs.FieldName = "isVerified"        // 是否验证通过
	UserField_RequirePlans      dbs.FieldName = "requirePlans"      // 是否需要购买套餐
	UserField_Modules           dbs.FieldName = "modules"           // 用户模块
	UserField_PriceType         dbs.FieldName = "priceType"         // 计费类型：traffic|bandwidth
	UserField_PricePeriod       dbs.FieldName = "pricePeriod"       // 结算周期
	UserField_ServersEnabled    dbs.FieldName = "serversEnabled"    // 是否禁用所有服务
	UserField_Notification      dbs.FieldName = "notification"      // 通知设置
	UserField_BandwidthAlgo     dbs.FieldName = "bandwidthAlgo"     // 带宽算法
	UserField_BandwidthModifier dbs.FieldName = "bandwidthModifier" // 带宽修正值
	UserField_Lang              dbs.FieldName = "lang"              // 语言代号
)

// User 用户
type User struct {
	Id                uint32   `field:"id"`                // ID
	IsOn              bool     `field:"isOn"`              // 是否启用
	Username          string   `field:"username"`          // 用户名
	Password          string   `field:"password"`          // 密码
	Fullname          string   `field:"fullname"`          // 真实姓名
	Mobile            string   `field:"mobile"`            // 手机号
	VerifiedMobile    string   `field:"verifiedMobile"`    // 已验证手机号
	Tel               string   `field:"tel"`               // 联系电话
	Remark            string   `field:"remark"`            // 备注
	Email             string   `field:"email"`             // 邮箱地址
	VerifiedEmail     string   `field:"verifiedEmail"`     // 激活后的邮箱
	EmailIsVerified   uint8    `field:"emailIsVerified"`   // 邮箱是否已验证
	AvatarFileId      uint64   `field:"avatarFileId"`      // 头像文件ID
	CreatedAt         uint64   `field:"createdAt"`         // 创建时间
	Day               string   `field:"day"`               // YYYYMMDD
	UpdatedAt         uint64   `field:"updatedAt"`         // 修改时间
	State             uint8    `field:"state"`             // 状态
	Source            string   `field:"source"`            // 来源
	ClusterId         uint32   `field:"clusterId"`         // 集群ID
	Features          dbs.JSON `field:"features"`          // 允许操作的特征
	RegisteredIP      string   `field:"registeredIP"`      // 注册使用的IP
	IsRejected        bool     `field:"isRejected"`        // 是否已拒绝
	RejectReason      string   `field:"rejectReason"`      // 拒绝理由
	IsVerified        bool     `field:"isVerified"`        // 是否验证通过
	RequirePlans      uint8    `field:"requirePlans"`      // 是否需要购买套餐
	Modules           dbs.JSON `field:"modules"`           // 用户模块
	PriceType         string   `field:"priceType"`         // 计费类型：traffic|bandwidth
	PricePeriod       string   `field:"pricePeriod"`       // 结算周期
	ServersEnabled    uint8    `field:"serversEnabled"`    // 是否禁用所有服务
	Notification      dbs.JSON `field:"notification"`      // 通知设置
	BandwidthAlgo     string   `field:"bandwidthAlgo"`     // 带宽算法
	BandwidthModifier float64  `field:"bandwidthModifier"` // 带宽修正值
	Lang              string   `field:"lang"`              // 语言代号
}

type UserOperator struct {
	Id                any // ID
	IsOn              any // 是否启用
	Username          any // 用户名
	Password          any // 密码
	Fullname          any // 真实姓名
	Mobile            any // 手机号
	VerifiedMobile    any // 已验证手机号
	Tel               any // 联系电话
	Remark            any // 备注
	Email             any // 邮箱地址
	VerifiedEmail     any // 激活后的邮箱
	EmailIsVerified   any // 邮箱是否已验证
	AvatarFileId      any // 头像文件ID
	CreatedAt         any // 创建时间
	Day               any // YYYYMMDD
	UpdatedAt         any // 修改时间
	State             any // 状态
	Source            any // 来源
	ClusterId         any // 集群ID
	Features          any // 允许操作的特征
	RegisteredIP      any // 注册使用的IP
	IsRejected        any // 是否已拒绝
	RejectReason      any // 拒绝理由
	IsVerified        any // 是否验证通过
	RequirePlans      any // 是否需要购买套餐
	Modules           any // 用户模块
	PriceType         any // 计费类型：traffic|bandwidth
	PricePeriod       any // 结算周期
	ServersEnabled    any // 是否禁用所有服务
	Notification      any // 通知设置
	BandwidthAlgo     any // 带宽算法
	BandwidthModifier any // 带宽修正值
	Lang              any // 语言代号
}

func NewUserOperator() *UserOperator {
	return &UserOperator{}
}
