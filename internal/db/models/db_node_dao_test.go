package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestDBNodeDAO_EncodePassword(t *testing.T) {
	dao := NewDBNodeDAO()
	for _, password := range []string{
		"123456",
		"abcdefxyz",
		"123abc$*&^%",
		"$%#@!@(*))*&^&=]{|",
		"中文",
	} {
		encoded := dao.EncodePassword(password)
		decoded := dao.DecodePassword(encoded)
		if decoded != password {
			t.Fatal(decoded, password)
		}
	}
}

func TestDBNodeDAO_EncodePassword_Encoded(t *testing.T) {
	dao := NewDBNodeDAO()
	password := DBNodePasswordEncodedPrefix + "123456"
	encoded := dao.EncodePassword(password)
	if encoded != password {
		t.Fatal()
	}
	t.Log(encoded)
}

func TestDBNodeDAO_EncodePassword_Decoded(t *testing.T) {
	dao := NewDBNodeDAO()
	password := "123456"
	decoded := dao.DecodePassword(password)
	if decoded != password {
		t.Fatal()
	}
	t.Log(decoded)
}
