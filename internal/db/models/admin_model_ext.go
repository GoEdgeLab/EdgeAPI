package models

import stringutil "github.com/iwind/TeaGo/utils/string"

// 弱密码集合
var weakPasswords = []string{}

func init() {
	// 初始化弱密码集合
	for _, password := range []string{
		"123",
		"1234",
		"12345",
		"123456",
		"12345678",
		"123456789",
		"000000",
		"111111",
		"666666",
		"888888",
		"654321",
		"123456789",
		"password",
		"qwerty",
		"admin",
	} {
		weakPasswords = append(weakPasswords, stringutil.Md5(password))
	}
}

func (this *Admin) HasWeakPassword() bool {
	if len(this.Password) == 0 {
		return false
	}

	for _, weakPassword := range weakPasswords {
		if weakPassword == this.Password {
			return true
		}
	}
	return false
}
