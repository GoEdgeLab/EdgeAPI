package models

import "github.com/iwind/TeaGo/dbs"

// LoginSession 登录Session
type LoginSession struct {
	Id        uint64   `field:"id"`        // ID
	AdminId   uint64   `field:"adminId"`   // 管理员ID
	UserId    uint64   `field:"userId"`    // 用户ID
	Sid       string   `field:"sid"`       // 令牌
	Values    dbs.JSON `field:"values"`    // 数据
	Ip        string   `field:"ip"`        // 登录IP
	CreatedAt uint64   `field:"createdAt"` // 创建时间
	ExpiresAt uint64   `field:"expiresAt"` // 过期时间
}

type LoginSessionOperator struct {
	Id        any // ID
	AdminId   any // 管理员ID
	UserId    any // 用户ID
	Sid       any // 令牌
	Values    any // 数据
	Ip        any // 登录IP
	CreatedAt any // 创建时间
	ExpiresAt any // 过期时间
}

func NewLoginSessionOperator() *LoginSessionOperator {
	return &LoginSessionOperator{}
}
