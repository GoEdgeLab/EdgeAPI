package acme

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/go-acme/lego/v4/registration"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestRequest_Run_DNS(t *testing.T) {
	privateKey, err := ParsePrivateKeyFromBase64("MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgD3xxDXP4YVqHCfub21Yi3QL1Kvgow23J8CKJ7vU3L4+hRANCAARRl5ZKAlgGRc5RETSMYFCTXvjnePDgjALWgtgfClQGLB2rGyRecJvlesAM6Q7LQrDxVxvxdSQQmPGRqJGiBtjd")
	if err != nil {
		t.Fatal(err)
	}

	user := NewUser("19644627@qq.com", privateKey, func(resource *registration.Resource) error {
		resourceJSON, err := json.Marshal(resource)
		if err != nil {
			return err
		}
		t.Log(string(resourceJSON))
		return nil
	})

	regResource := []byte(`{"body":{"status":"valid","contact":["mailto:19644627@qq.com"]},"uri":"https://acme-v02.api.letsencrypt.org/acme/acct/103672877"}`)
	err = user.SetRegistration(regResource)
	if err != nil {
		t.Fatal(err)
	}

	dnsProvider, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}

	req := NewRequest(&Task{
		User:        user,
		Type:        TaskTypeDNS,
		DNSProvider: dnsProvider,
		DNSDomain:   "yun4s.cn",
		Domains:     []string{"yun4s.cn"},
	})
	certData, keyData, err := req.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cert:", string(certData))
	t.Log("key:", string(keyData))
}

func TestRequest_Run_HTTP(t *testing.T) {
	privateKey, err := ParsePrivateKeyFromBase64("MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgD3xxDXP4YVqHCfub21Yi3QL1Kvgow23J8CKJ7vU3L4+hRANCAARRl5ZKAlgGRc5RETSMYFCTXvjnePDgjALWgtgfClQGLB2rGyRecJvlesAM6Q7LQrDxVxvxdSQQmPGRqJGiBtjd")
	if err != nil {
		t.Fatal(err)
	}

	user := NewUser("19644627@qq.com", privateKey, func(resource *registration.Resource) error {
		resourceJSON, err := json.Marshal(resource)
		if err != nil {
			return err
		}
		t.Log(string(resourceJSON))
		return nil
	})

	regResource := []byte(`{"body":{"status":"valid","contact":["mailto:19644627@qq.com"]},"uri":"https://acme-v02.api.letsencrypt.org/acme/acct/103672877"}`)
	err = user.SetRegistration(regResource)
	if err != nil {
		t.Fatal(err)
	}

	req := NewRequest(&Task{
		User:    user,
		Type:    TaskTypeHTTP,
		Domains: []string{"teaos.cn", "www.teaos.cn", "meloy.cn"},
	})
	certData, keyData, err := req.runHTTP()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(certData))
	t.Log(string(keyData))
}

func testDNSPodProvider() (dnsclients.ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='dnspod' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &dnsclients.DNSPodProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
