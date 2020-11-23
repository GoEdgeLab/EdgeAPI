package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	acmelog "github.com/go-acme/lego/v4/log"
	"io/ioutil"
	"log"
	"testing"

	"github.com/go-acme/lego/v4/registration"
)

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type MyProvider struct {
	t *testing.T
}

func (this *MyProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	this.t.Log("provider: domain:", domain, "fqdn:", fqdn, "value:", value)
	return nil
}

func (this *MyProvider) CleanUp(domain, token, keyAuth string) error {
	return nil
}

// 参考  https://go-acme.github.io/lego/usage/library/
func TestGenerate(t *testing.T) {
	acmelog.Logger = log.New(ioutil.Discard, "", log.LstdFlags)

	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	myUser := &MyUser{
		Email: "test1@teaos.cn",
		key:   privateKey,
	}

	config := lego.NewConfig(myUser)
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Challenge.SetDNS01Provider(&MyProvider{t: t})
	if err != nil {
		t.Fatal(err)
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		t.Fatal(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"teaos.com"},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(certificates)
}
