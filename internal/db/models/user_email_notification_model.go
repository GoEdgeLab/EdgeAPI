package models

// UserEmailNotification 邮件通知队列
type UserEmailNotification struct {
	Id        uint64 `field:"id"`        // ID
	Email     string `field:"email"`     // 邮箱地址
	Subject   string `field:"subject"`   // 标题
	Body      string `field:"body"`      // 内容
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	Day       string `field:"day"`       // YYYYMMDD
}

type UserEmailNotificationOperator struct {
	Id        any // ID
	Email     any // 邮箱地址
	Subject   any // 标题
	Body      any // 内容
	CreatedAt any // 创建时间
	Day       any // YYYYMMDD
}

func NewUserEmailNotificationOperator() *UserEmailNotificationOperator {
	return &UserEmailNotificationOperator{}
}
