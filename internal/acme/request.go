package acme

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	acmelog "github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/registration"
	"io/ioutil"
	"log"
)

type Request struct {
	debug bool

	task *Task
}

func NewRequest(task *Task) *Request {
	return &Request{
		task: task,
	}
}

func (this *Request) Debug() {
	this.debug = true
}

func (this *Request) Run() (certData []byte, keyData []byte, err error) {
	if !this.debug {
		acmelog.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}

	if this.task.User == nil {
		err = errors.New("'user' must not be nil")
		return
	}
	if this.task.DNSProvider == nil {
		err = errors.New("'dnsProvider' must not be nil")
		return
	}
	if len(this.task.DNSDomain) == 0 {
		err = errors.New("'dnsDomain' must not be empty")
		return
	}
	if len(this.task.Domains) == 0 {
		err = errors.New("'domains' must not be empty")
		return
	}

	config := lego.NewConfig(this.task.User)
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	// 注册用户
	resource := this.task.User.GetRegistration()
	if resource != nil {
		resource, err = client.Registration.QueryRegistration()
		if err != nil {
			return nil, nil, err
		}
	} else {
		resource, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, nil, err
		}
		err = this.task.User.Register(resource)
		if err != nil {
			return nil, nil, err
		}
	}

	err = client.Challenge.SetDNS01Provider(NewDNSProvider(this.task.DNSProvider))
	if err != nil {
		return nil, nil, err
	}

	// 申请证书
	request := certificate.ObtainRequest{
		Domains: this.task.Domains,
		Bundle:  true,
	}
	certResource, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, nil, err
	}

	return certResource.Certificate, certResource.PrivateKey, nil
}
