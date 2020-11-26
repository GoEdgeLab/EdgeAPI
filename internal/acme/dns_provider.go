package acme

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"strings"
)

type DNSProvider struct {
	raw dnsclients.ProviderInterface
}

func NewDNSProvider(raw dnsclients.ProviderInterface) *DNSProvider {
	return &DNSProvider{raw: raw}
}

func (this *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// 设置记录
	index := strings.Index(fqdn, "."+domain)
	if index < 0 {
		return errors.New("invalid fqdn value")
	}
	recordName := fqdn[:index]
	record, err := this.raw.QueryRecord(domain, recordName, dnsclients.RecordTypeTXT)
	if err != nil {
		return errors.New("query DNS record failed: " + err.Error())
	}
	if record == nil {
		err = this.raw.AddRecord(domain, &dnsclients.Record{
			Id:    "",
			Name:  recordName,
			Type:  dnsclients.RecordTypeTXT,
			Value: value,
			Route: this.raw.DefaultRoute(),
		})
		if err != nil {
			return errors.New("create DNS record failed: " + err.Error())
		}
	} else {
		err = this.raw.UpdateRecord(domain, record, &dnsclients.Record{
			Name:  recordName,
			Type:  dnsclients.RecordTypeTXT,
			Value: value,
			Route: this.raw.DefaultRoute(),
		})
		if err != nil {
			return errors.New("update DNS record failed: " + err.Error())
		}
	}

	return nil
}

func (this *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return nil
}
