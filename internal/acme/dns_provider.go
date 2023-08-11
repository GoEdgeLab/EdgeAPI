package acme

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/iwind/TeaGo/lists"
	"os"
	"strings"
	"sync"
)

type DNSProvider struct {
	raw       dnsclients.ProviderInterface
	dnsDomain string

	locker             sync.Mutex
	deletedRecordNames []string
}

func NewDNSProvider(raw dnsclients.ProviderInterface, dnsDomain string) *DNSProvider {
	return &DNSProvider{
		raw:       raw,
		dnsDomain: dnsDomain,
	}
}

func (this *DNSProvider) Present(domain, token, keyAuth string) error {
	_ = os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "true")
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// 设置记录
	var index = strings.Index(fqdn, "."+this.dnsDomain)
	if index < 0 {
		return errors.New("invalid fqdn value")
	}
	var recordName = fqdn[:index]

	// 先删除老的
	this.locker.Lock()
	var wasDeleted = lists.ContainsString(this.deletedRecordNames, recordName)
	this.locker.Unlock()

	if !wasDeleted {
		records, err := this.raw.QueryRecords(this.dnsDomain, recordName, dnstypes.RecordTypeTXT)
		if err != nil {
			return fmt.Errorf("query DNS record failed: %w", err)
		}
		for _, record := range records {
			err = this.raw.DeleteRecord(this.dnsDomain, record)
			if err != nil {
				return err
			}
		}
		this.locker.Lock()
		this.deletedRecordNames = append(this.deletedRecordNames, recordName)
		this.locker.Unlock()
	}

	// 添加新的
	err := this.raw.AddRecord(this.dnsDomain, &dnstypes.Record{
		Id:    "",
		Name:  recordName,
		Type:  dnstypes.RecordTypeTXT,
		Value: value,
		Route: this.raw.DefaultRoute(),
	})
	if err != nil {
		return fmt.Errorf("create DNS record failed: %w", err)
	}

	return nil
}

func (this *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return nil
}
