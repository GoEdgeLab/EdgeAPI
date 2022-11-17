// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"sync"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		go func() {
			var ticker = time.NewTicker(1 * time.Hour)
			for range ticker.C {
				sharedDomainRecordsCache.Clean()
			}
		}()
	})
}

type recordList struct {
	version   int64
	updatedAt int64
	records   []*dnstypes.Record
}

var sharedDomainRecordsCache = NewDomainRecordsCache()

// DomainRecordsCache 域名记录缓存
type DomainRecordsCache struct {
	domainRecordsMap map[string]*recordList // domain@providerId => record
	locker           sync.Mutex
}

func NewDomainRecordsCache() *DomainRecordsCache {
	return &DomainRecordsCache{
		domainRecordsMap: map[string]*recordList{},
	}
}

// WriteDomainRecords 写入域名记录缓存
func (this *DomainRecordsCache) WriteDomainRecords(providerId int64, domain string, records []*dnstypes.Record) {
	if providerId <= 0 || len(domain) == 0 {
		return
	}
	domain = types.String(providerId) + "@" + domain

	this.locker.Lock()
	defer this.locker.Unlock()

	// 版本号
	var key = "DomainRecordsCache" + "@" + types.String(providerId) + "@" + domain
	version, err := models.SharedSysLockerDAO.Increase(nil, key, 1)
	if err != nil {
		remotelogs.Error("dnsclients.BaseProvider", "WriteDomainRecordsCache: "+err.Error())
		return
	}

	var clonedRecords = []*dnstypes.Record{}
	for _, record := range records {
		clonedRecords = append(clonedRecords, record)
	}
	this.domainRecordsMap[domain] = &recordList{
		version:   version,
		updatedAt: time.Now().Unix(),
		records:   clonedRecords,
	}
}

// QueryDomainRecord 从缓存中读取单条域名记录
func (this *DomainRecordsCache) QueryDomainRecord(providerId int64, domain string, recordName string, recordType string) (record *dnstypes.Record, hasRecords bool, ok bool) {
	if providerId <= 0 || len(domain) == 0 {
		return
	}

	domain = types.String(providerId) + "@" + domain

	this.locker.Lock()
	defer this.locker.Unlock()

	// check version
	var key = "DomainRecordsCache" + "@" + types.String(providerId) + "@" + domain
	version, err := models.SharedSysLockerDAO.Read(nil, key)
	if err != nil {
		remotelogs.Error("dnsclients.BaseProvider", "ReadDomainRecordsCache: "+err.Error())
		return
	}

	// find list
	list, recordsOk := this.domainRecordsMap[domain]
	if !recordsOk {
		return
	}
	if version != list.version {
		delete(this.domainRecordsMap, domain)
		return
	}

	// check timestamp
	if list.updatedAt < time.Now().Unix()-86400 /** 缓存有效期为一天 **/ {
		delete(this.domainRecordsMap, domain)
		return
	}

	hasRecords = true
	for _, r := range list.records {
		if r.Name == recordName && r.Type == recordType {
			return r, true, true
		}
	}

	return
}

// QueryDomainRecords 从缓存中读取多条域名记录
func (this *DomainRecordsCache) QueryDomainRecords(providerId int64, domain string, recordName string, recordType string) (records []*dnstypes.Record, hasRecords bool, ok bool) {
	if providerId <= 0 || len(domain) == 0 {
		return
	}

	domain = types.String(providerId) + "@" + domain

	this.locker.Lock()
	defer this.locker.Unlock()

	// check version
	var key = "DomainRecordsCache" + "@" + types.String(providerId) + "@" + domain
	version, err := models.SharedSysLockerDAO.Read(nil, key)
	if err != nil {
		remotelogs.Error("dnsclients.BaseProvider", "ReadDomainRecordsCache: "+err.Error())
		return
	}

	// find list
	list, recordsOk := this.domainRecordsMap[domain]
	if !recordsOk {
		return
	}
	if version != list.version {
		delete(this.domainRecordsMap, domain)
		return
	}

	// check timestamp
	if list.updatedAt < time.Now().Unix()-86400 /** 缓存有效期为一天 **/ {
		delete(this.domainRecordsMap, domain)
		return
	}

	hasRecords = true
	for _, r := range list.records {
		if r.Name == recordName && r.Type == recordType {
			records = append(records, r)
			ok = true
		}
	}

	return
}

// DeleteDomainRecord 删除域名记录缓存
func (this *DomainRecordsCache) DeleteDomainRecord(providerId int64, domain string, recordId string) {
	if providerId <= 0 || len(domain) == 0 || len(recordId) == 0 {
		return
	}

	domain = types.String(providerId) + "@" + domain

	this.locker.Lock()
	defer this.locker.Unlock()

	list, ok := this.domainRecordsMap[domain]
	if !ok {
		return
	}
	var found = false
	var newRecords = []*dnstypes.Record{}
	for _, record := range list.records {
		if record.Id == recordId {
			found = true
			continue
		}
		newRecords = append(newRecords, record)
	}
	if found {
		list.records = newRecords
	}
}

// AddDomainRecord 添加域名记录缓存
func (this *DomainRecordsCache) AddDomainRecord(providerId int64, domain string, record *dnstypes.Record) {
	if providerId <= 0 || len(domain) == 0 || record == nil || len(record.Id) == 0 {
		return
	}

	domain = types.String(providerId) + "@" + domain

	this.locker.Lock()
	defer this.locker.Unlock()

	list, ok := this.domainRecordsMap[domain]
	if ok {
		list.records = append(list.records, record.Clone())
	}

	// 如果完全没有记录，则不保存
}

// UpdateDomainRecord 修改域名记录缓存
func (this *DomainRecordsCache) UpdateDomainRecord(providerId int64, domain string, record *dnstypes.Record) {
	if providerId <= 0 || len(domain) == 0 || record == nil || len(record.Id) == 0 {
		return
	}

	domain = types.String(providerId) + "@" + domain

	this.locker.Lock()
	defer this.locker.Unlock()

	list, ok := this.domainRecordsMap[domain]
	if !ok {
		return
	}
	for _, r := range list.records {
		if r.Id == record.Id {
			r.Copy(record)
			break
		}
	}
}

// Clean 清除过期缓存
func (this *DomainRecordsCache) Clean() {
	this.locker.Lock()
	defer this.locker.Unlock()

	for domain, list := range this.domainRecordsMap {
		if list.updatedAt < time.Now().Unix()-86400 {
			delete(this.domainRecordsMap, domain)
		}
	}
}
