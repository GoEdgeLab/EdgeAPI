package acme

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	acmelog "github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/registration"
	"io"
	"log"
)

type Request struct {
	debug bool

	task   *Task
	onAuth AuthCallback
}

func NewRequest(task *Task) *Request {
	return &Request{
		task: task,
	}
}

func (this *Request) Debug() {
	this.debug = true
}

func (this *Request) OnAuth(onAuth AuthCallback) {
	this.onAuth = onAuth
}

func (this *Request) Run() (certData []byte, keyData []byte, err error) {
	if this.task.Provider == nil {
		err = errors.New("provider should not be nil")
		return
	}
	if this.task.Provider.RequireEAB && this.task.Account == nil {
		err = errors.New("account should not be nil when provider require EAB")
	}

	switch this.task.AuthType {
	case AuthTypeDNS:
		return this.runDNS()
	case AuthTypeHTTP:
		return this.runHTTP()
	default:
		err = errors.New("invalid task type '" + this.task.AuthType + "'")
		return
	}
}

func (this *Request) runDNS() (certData []byte, keyData []byte, err error) {
	if !this.debug {
		acmelog.Logger = log.New(io.Discard, "", log.LstdFlags)
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

	var config = lego.NewConfig(this.task.User)
	config.Certificate.KeyType = certcrypto.RSA2048
	config.CADirURL = this.task.Provider.APIURL
	config.UserAgent = teaconst.ProductName + "/" + teaconst.Version

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	// 注册用户
	var resource = this.task.User.GetRegistration()
	if resource != nil {
		resource, err = client.Registration.QueryRegistration()
		if err != nil {
			return nil, nil, err
		}
	} else {
		if this.task.Provider.RequireEAB {
			resource, err := client.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
				TermsOfServiceAgreed: true,
				Kid:                  this.task.Account.EABKid,
				HmacEncoded:          this.task.Account.EABKey,
			})
			if err != nil {
				return nil, nil, errors.New("register user failed: " + err.Error())
			}
			err = this.task.User.Register(resource)
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
	}

	err = client.Challenge.SetDNS01Provider(NewDNSProvider(this.task.DNSProvider, this.task.DNSDomain))
	if err != nil {
		return nil, nil, err
	}

	// 申请证书
	var request = certificate.ObtainRequest{
		Domains: this.task.Domains,
		Bundle:  true,
	}
	certResource, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, nil, errors.New("obtain cert failed: " + err.Error())
	}

	return certResource.Certificate, certResource.PrivateKey, nil
}

func (this *Request) runHTTP() (certData []byte, keyData []byte, err error) {
	if !this.debug {
		acmelog.Logger = log.New(io.Discard, "", log.LstdFlags)
	}

	if this.task.User == nil {
		err = errors.New("'user' must not be nil")
		return
	}

	var config = lego.NewConfig(this.task.User)
	config.Certificate.KeyType = certcrypto.RSA2048
	config.CADirURL = this.task.Provider.APIURL
	config.UserAgent = teaconst.ProductName + "/" + teaconst.Version

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	// 注册用户
	var resource = this.task.User.GetRegistration()
	if resource != nil {
		resource, err = client.Registration.QueryRegistration()
		if err != nil {
			return nil, nil, err
		}
	} else {
		if this.task.Provider.RequireEAB {
			resource, err := client.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
				TermsOfServiceAgreed: true,
				Kid:                  this.task.Account.EABKid,
				HmacEncoded:          this.task.Account.EABKey,
			})
			if err != nil {
				return nil, nil, errors.New("register user failed: " + err.Error())
			}
			err = this.task.User.Register(resource)
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
	}

	err = client.Challenge.SetHTTP01Provider(NewHTTPProvider(this.onAuth))
	if err != nil {
		return nil, nil, err
	}

	// 申请证书
	var request = certificate.ObtainRequest{
		Domains: this.task.Domains,
		Bundle:  true,
	}
	certResource, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, nil, err
	}

	return certResource.Certificate, certResource.PrivateKey, nil
}
