package acme

import (
	"crypto"
	"encoding/json"
	"github.com/go-acme/lego/v4/registration"
)

type User struct {
	email        string
	resource     *registration.Resource
	key          crypto.PrivateKey
	registerFunc func(resource *registration.Resource) error
}

func NewUser(email string, key crypto.PrivateKey, registerFunc func(resource *registration.Resource) error) *User {
	return &User{
		email:        email,
		key:          key,
		registerFunc: registerFunc,
	}
}

func (this *User) GetEmail() string {
	return this.email
}

func (this *User) GetRegistration() *registration.Resource {
	return this.resource
}

func (this *User) SetRegistration(resourceData []byte) error {
	resource := &registration.Resource{}
	err := json.Unmarshal(resourceData, resource)
	if err != nil {
		return err
	}
	this.resource = resource
	return nil
}

func (this *User) GetPrivateKey() crypto.PrivateKey {
	return this.key
}

func (this *User) Register(resource *registration.Resource) error {
	this.resource = resource
	return this.registerFunc(resource)
}
