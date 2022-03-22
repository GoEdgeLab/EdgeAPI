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
	IsRejected      uint8    `field:"isRejected"`      // 是否已拒绝
	RejectReason    string   `field:"rejectReason"`    // 拒绝理由
	IsVerified      uint8    `field:"isVerified"`      // 是否验证通过
	RequirePlans    uint8    `field:"requirePlans"`    // 是否需要购买套餐
}

type UserOperator struct {
	Id              interface{} // ID
	IsOn            interface{} // 是否启用
	Username        interface{} // 用户名
	Password        interface{} // 密码
	Fullname        interface{} // 真实姓名
	Mobile          interface{} // 手机号
	Tel             interface{} // 联系电话
	Remark          interface{} // 备注
	Email           interface{} // 邮箱地址
	EmailIsVerified interface{} // 邮箱是否已验证
	AvatarFileId    interface{} // 头像文件ID
	CreatedAt       interface{} // 创建时间
	Day             interface{} // YYYYMMDD
	UpdatedAt       interface{} // 修改时间
	State           interface{} // 状态
	Source          interface{} // 来源
	ClusterId       interface{} // 集群ID
	Features        interface{} // 允许操作的特征
	RegisteredIP    interface{} // 注册使用的IP
	IsRejected      interface{} // 是否已拒绝
	RejectReason    interface{} // 拒绝理由
	IsVerified      interface{} // 是否验证通过
	RequirePlans    interface{} // 是否需要购买套餐
}

func NewUserOperator() *UserOperator {
	return &UserOperator{}
}
