package models

import "github.com/iwind/TeaGo/dbs"

// UserIdentity 用户实名认证信息
type UserIdentity struct {
	Id           uint64   `field:"id"`           // ID
	UserId       uint64   `field:"userId"`       // 用户ID
	OrgType      string   `field:"orgType"`      // 组织类型
	Type         string   `field:"type"`         // 证件类型
	RealName     string   `field:"realName"`     // 真实姓名
	Number       string   `field:"number"`       // 编号
	FileIds      dbs.JSON `field:"fileIds"`      // 文件ID
	Status       string   `field:"status"`       // 状态：none,submitted,verified,rejected
	State        uint8    `field:"state"`        // 状态
	CreatedAt    uint64   `field:"createdAt"`    // 创建时间
	UpdatedAt    uint64   `field:"updatedAt"`    // 修改时间
	SubmittedAt  uint64   `field:"submittedAt"`  // 提交时间
	RejectedAt   uint64   `field:"rejectedAt"`   // 拒绝时间
	VerifiedAt   uint64   `field:"verifiedAt"`   // 认证时间
	RejectReason string   `field:"rejectReason"` // 拒绝原因
}

type UserIdentityOperator struct {
	Id           interface{} // ID
	UserId       interface{} // 用户ID
	OrgType      interface{} // 组织类型
	Type         interface{} // 证件类型
	RealName     interface{} // 真实姓名
	Number       interface{} // 编号
	FileIds      interface{} // 文件ID
	Status       interface{} // 状态：none,submitted,verified,rejected
	State        interface{} // 状态
	CreatedAt    interface{} // 创建时间
	UpdatedAt    interface{} // 修改时间
	SubmittedAt  interface{} // 提交时间
	RejectedAt   interface{} // 拒绝时间
	VerifiedAt   interface{} // 认证时间
	RejectReason interface{} // 拒绝原因
}

func NewUserIdentityOperator() *UserIdentityOperator {
	return &UserIdentityOperator{}
}
