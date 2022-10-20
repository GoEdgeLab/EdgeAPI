package tasks

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	dnsmodels "github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"net"
	"strings"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewDNSTaskExecutor(10 * time.Second).Start()
		})
	})
}

// DNSTaskExecutor DNS任务执行器
type DNSTaskExecutor struct {
	BaseTask

	ticker *time.Ticker
}

func NewDNSTaskExecutor(duration time.Duration) *DNSTaskExecutor {
	return &DNSTaskExecutor{
		ticker: time.NewTicker(duration),
	}
}

func (this *DNSTaskExecutor) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("DNSTaskExecutor", err.Error())
		}
	}
}

func (this *DNSTaskExecutor) Loop() error {
	if !models.SharedAPINodeDAO.CheckAPINodeIsPrimaryWithoutErr() {
		return nil
	}

	return this.loop()
}

func (this *DNSTaskExecutor) loop() error {
	tasks, err := dnsmodels.SharedDNSTaskDAO.FindAllDoingTasks(nil)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		taskId := int64(task.Id)
		switch task.Type {
		case dnsmodels.DNSTaskTypeServerChange:
			err = this.doServer(taskId, int64(task.ClusterId), int64(task.ServerId))
			if err != nil {
				err = dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskError(nil, taskId, err.Error())
				if err != nil {
					return err
				}
			}
		case dnsmodels.DNSTaskTypeNodeChange:
			err = this.doNode(taskId, int64(task.NodeId))
			if err != nil {
				err = dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskError(nil, taskId, err.Error())
				if err != nil {
					return err
				}
			}
		case dnsmodels.DNSTaskTypeClusterChange:
			err = this.doCluster(taskId, int64(task.ClusterId))
			if err != nil {
				err = dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskError(nil, taskId, err.Error())
				if err != nil {
					return err
				}
			}
		case dnsmodels.DNSTaskTypeClusterRemoveDomain:
			err = this.doClusterRemove(taskId, int64(task.ClusterId), int64(task.DomainId), task.RecordName)
			if err != nil {
				err = dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskError(nil, taskId, err.Error())
				if err != nil {
					return err
				}
			}
		case dnsmodels.DNSTaskTypeDomainChange:
			err = this.doDomainWithTask(taskId, int64(task.DomainId))
			if err != nil {
				err = dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskError(nil, taskId, err.Error())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// 修改服务相关记录
func (this *DNSTaskExecutor) doServer(taskId int64, oldClusterId int64, serverId int64) error {
	var tx *dbs.Tx

	isOk := false
	defer func() {
		if isOk {
			err := dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskDone(tx, taskId)
			if err != nil {
				this.logErr("DNSTaskExecutor", err.Error())
			}
		}
	}()

	// 检查是否已通过审核
	serverDNS, err := models.SharedServerDAO.FindStatelessServerDNS(tx, serverId)
	if err != nil {
		return err
	}
	if serverDNS == nil {
		isOk = true
		return nil
	}
	if len(serverDNS.DnsName) == 0 {
		isOk = true
		return nil
	}

	var recordName = serverDNS.DnsName
	var recordType = dnstypes.RecordTypeCNAME

	// 新的DNS设置
	manager, newDomainId, domain, clusterDNSName, dnsConfig, err := this.findDNSManagerWithClusterId(tx, int64(serverDNS.ClusterId))
	if err != nil {
		return err
	}

	// 如果集群发生了变化，则从老的集群中删除
	if oldClusterId > 0 && int64(serverDNS.ClusterId) != oldClusterId {
		oldManager, oldDomainId, oldDomain, _, _, err := this.findDNSManagerWithClusterId(tx, oldClusterId)
		if err != nil {
			return err
		}

		// 如果域名发生了变化
		if oldDomainId != newDomainId {
			if oldManager != nil {
				oldRecord, err := oldManager.QueryRecord(oldDomain, recordName, recordType)
				if err != nil {
					return err
				}
				if oldRecord != nil {
					// 删除记录
					err = oldManager.DeleteRecord(oldDomain, oldRecord)
					if err != nil {
						return err
					}

					// 更新域名中记录缓存
					// 这里不创建域名更新任务，而是直接更新，避免影响其他任务的执行
					err = this.doDomain(oldDomainId)
					if err != nil {
						return err
					}
				}
			}
		}

		isOk = true
		return nil
	}

	// 处理新的集群
	if manager == nil {
		isOk = true
		return nil
	}
	var ttl int32 = 0
	if dnsConfig != nil {
		ttl = dnsConfig.TTL
	}

	recordValue := clusterDNSName + "." + domain + "."
	recordRoute := manager.DefaultRoute()
	if serverDNS.State == models.ServerStateDisabled || !serverDNS.IsOn {
		// 检查记录是否已经存在
		record, err := manager.QueryRecord(domain, recordName, recordType)
		if err != nil {
			return err
		}
		if record != nil {
			// 删除
			err = manager.DeleteRecord(domain, record)
			if err != nil {
				return err
			}
			err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, newDomainId, dnsmodels.DNSTaskTypeDomainChange)
			if err != nil {
				return err
			}
		}

		isOk = true
	} else {
		// 是否已存在
		exist, err := dnsmodels.SharedDNSDomainDAO.ExistDomainRecord(tx, newDomainId, recordName, recordType, recordRoute, recordValue)
		if err != nil {
			return err
		}
		if exist {
			isOk = true
			return nil
		}

		// 检查记录是否已经存在
		record, err := manager.QueryRecord(domain, recordName, recordType)
		if err != nil {
			return err
		}
		if record != nil {
			if record.Value == recordValue || record.Value == strings.TrimRight(recordValue, ".") {
				isOk = true
				return nil
			}

			// 删除
			err = manager.DeleteRecord(domain, record)
			if err != nil {
				return err
			}
			err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, newDomainId, dnsmodels.DNSTaskTypeDomainChange)
			if err != nil {
				return err
			}
		}

		err = manager.AddRecord(domain, &dnstypes.Record{
			Id:    "",
			Name:  recordName,
			Type:  recordType,
			Value: recordValue,
			Route: recordRoute,
			TTL:   ttl,
		})
		if err != nil {
			return err
		}

		err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, newDomainId, dnsmodels.DNSTaskTypeDomainChange)
		if err != nil {
			return err
		}

		isOk = true
	}

	return nil
}

// 修改节点相关记录
func (this *DNSTaskExecutor) doNode(taskId int64, nodeId int64) error {
	isOk := false
	defer func() {
		if isOk {
			err := dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskDone(nil, taskId)
			if err != nil {
				this.logErr("DNSTaskExecutor", err.Error())
			}
		}
	}()

	var tx *dbs.Tx
	node, err := models.SharedNodeDAO.FindStatelessNodeDNS(tx, nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		isOk = true
		return nil
	}

	// 转交给cluster统一处理
	clusterIds, err := models.SharedNodeDAO.FindEnabledAndOnNodeClusterIds(tx, nodeId)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		err = dnsmodels.SharedDNSTaskDAO.CreateClusterTask(tx, clusterId, dnsmodels.DNSTaskTypeClusterChange)
		if err != nil {
			return err
		}
	}

	isOk = true

	return nil
}

// 修改集群相关记录
func (this *DNSTaskExecutor) doCluster(taskId int64, clusterId int64) error {
	isOk := false
	defer func() {
		if isOk {
			err := dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskDone(nil, taskId)
			if err != nil {
				this.logErr("DNSTaskExecutor", err.Error())
			}
		}
	}()

	var tx *dbs.Tx
	manager, domainId, domain, clusterDNSName, dnsConfig, err := this.findDNSManagerWithClusterId(tx, clusterId)
	if err != nil {
		return err
	}
	if manager == nil {
		isOk = true
		return nil
	}

	var clusterDomain = clusterDNSName + "." + domain

	var ttl int32 = 0
	if dnsConfig != nil {
		ttl = dnsConfig.TTL
	}

	// 以前的节点记录
	records, err := manager.GetRecords(domain)
	if err != nil {
		return err
	}
	var oldRecordsMap = map[string]*dnstypes.Record{}      // route@value => record
	var oldCnameRecordsMap = map[string]*dnstypes.Record{} // cname => record
	for _, record := range records {
		if (record.Type == dnstypes.RecordTypeA || record.Type == dnstypes.RecordTypeAAAA) && record.Name == clusterDNSName {
			key := record.Route + "@" + record.Value
			oldRecordsMap[key] = record
		}

		if record.Type == dnstypes.RecordTypeCNAME {
			oldCnameRecordsMap[record.Name] = record
		}
	}

	// 当前的节点记录
	var newRecordKeys = []string{}
	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(tx, clusterId, true, dnsConfig != nil && dnsConfig.IncludingLnNodes)
	if err != nil {
		return err
	}
	var isChanged = false
	for _, node := range nodes {
		routes, err := node.DNSRouteCodesForDomainId(domainId)
		if err != nil {
			return err
		}
		if len(routes) == 0 {
			routes = []string{manager.DefaultRoute()}
		}

		// 所有的IP记录
		ipAddresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(tx, int64(node.Id), nodeconfigs.NodeRoleNode)
		if err != nil {
			return err
		}
		if len(ipAddresses) == 0 {
			continue
		}
		for _, ipAddress := range ipAddresses {
			var ip = ipAddress.DNSIP()
			if len(ip) == 0 || !ipAddress.CanAccess || !ipAddress.IsUp || !ipAddress.IsOn {
				continue
			}
			if net.ParseIP(ip) == nil {
				continue
			}
			for _, route := range routes {
				var key = route + "@" + ip
				_, ok := oldRecordsMap[key]
				if ok {
					newRecordKeys = append(newRecordKeys, key)
					continue
				}

				var recordType = dnstypes.RecordTypeA
				if utils.IsIPv6(ip) {
					recordType = dnstypes.RecordTypeAAAA
				}
				err = manager.AddRecord(domain, &dnstypes.Record{
					Id:    "",
					Name:  clusterDNSName,
					Type:  recordType,
					Value: ip,
					Route: route,
					TTL:   ttl,
				})
				if err != nil {
					return err
				}
				isChanged = true
				newRecordKeys = append(newRecordKeys, key)
			}
		}
	}

	// 删除多余的节点解析记录
	for key, record := range oldRecordsMap {
		if !lists.ContainsString(newRecordKeys, key) {
			isChanged = true
			err = manager.DeleteRecord(domain, record)
			if err != nil {
				return err
			}
		}
	}

	// 服务域名
	servers, err := models.SharedServerDAO.FindAllServersDNSWithClusterId(tx, clusterId)
	if err != nil {
		return err
	}
	serverRecords := []*dnstypes.Record{}             // 之所以用数组再存一遍，是因为dnsName可能会重复
	serverRecordsMap := map[string]*dnstypes.Record{} // dnsName => *Record
	for _, record := range records {
		if record.Type == dnstypes.RecordTypeCNAME && record.Value == clusterDomain+"." {
			serverRecords = append(serverRecords, record)
			serverRecordsMap[record.Name] = record
		}
	}

	// 新增的域名
	serverDNSNames := []string{}
	for _, server := range servers {
		dnsName := server.DnsName
		if len(dnsName) == 0 {
			continue
		}
		serverDNSNames = append(serverDNSNames, dnsName)
		_, ok := serverRecordsMap[dnsName]
		if !ok {
			isChanged = true
			err = manager.AddRecord(domain, &dnstypes.Record{
				Id:    "",
				Name:  dnsName,
				Type:  dnstypes.RecordTypeCNAME,
				Value: clusterDomain + ".",
				Route: "", // 注意这里为空，需要在执行过程中获取默认值
				TTL:   ttl,
			})
			if err != nil {
				return err
			}
		}
	}

	// 自动设置的CNAME
	var cnameRecords = []string{}
	if dnsConfig != nil {
		cnameRecords = dnsConfig.CNAMERecords
	}
	for _, cnameRecord := range cnameRecords {
		serverDNSNames = append(serverDNSNames, cnameRecord)
		_, ok := serverRecordsMap[cnameRecord]
		if !ok {
			isChanged = true
			err = manager.AddRecord(domain, &dnstypes.Record{
				Id:    "",
				Name:  cnameRecord,
				Type:  dnstypes.RecordTypeCNAME,
				Value: clusterDomain + ".",
				Route: "", // 注意这里为空，需要在执行过程中获取默认值
				TTL:   ttl,
			})
			if err != nil {
				return err
			}
		}
	}

	// 多余的域名
	for _, record := range serverRecords {
		if !lists.ContainsString(serverDNSNames, record.Name) {
			isChanged = true
			err = manager.DeleteRecord(domain, record)
			if err != nil {
				return err
			}
		}
	}

	// 通知更新域名
	if isChanged {
		err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, domainId, dnsmodels.DNSTaskTypeDomainChange)
		if err != nil {
			return err
		}
	}

	isOk = true

	return nil
}

func (this *DNSTaskExecutor) doClusterRemove(taskId int64, clusterId int64, domainId int64, dnsName string) error {
	var isOk = false
	defer func() {
		if isOk {
			err := dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskDone(nil, taskId)
			if err != nil {
				this.logErr("DNSTaskExecutor", err.Error())
			}
		}
	}()

	var tx *dbs.Tx

	dnsInfo, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId, nil)
	if err != nil {
		return err
	}

	if len(dnsName) == 0 {
		if dnsInfo == nil {
			isOk = true
			return nil
		}
		dnsName = dnsInfo.DnsName
		if len(dnsName) == 0 {
			isOk = true
			return nil
		}
	}

	// 再次检查是否正在使用，如果正在使用，则直接返回
	if dnsInfo != nil && dnsInfo.State == models.NodeClusterStateEnabled /** 尚未被删除 **/ && int64(dnsInfo.DnsDomainId) == domainId && dnsInfo.DnsName == dnsName {
		isOk = true
		return nil
	}

	domain, manager, err := this.findDNSManagerWithDomainId(tx, domainId)
	if err != nil {
		return err
	}
	if domain == nil {
		isOk = true
		return nil
	}
	var fullName = dnsName + "." + domain.Name

	records, err := domain.DecodeRecords()
	if err != nil {
		return err
	}

	var isChanged bool

	for _, record := range records {
		// node A
		if (record.Type == dnstypes.RecordTypeA || record.Type == dnstypes.RecordTypeAAAA) && record.Name == dnsName {
			err = manager.DeleteRecord(domain.Name, record)
			if err != nil {
				return err
			}
			isChanged = true
		}

		// server CNAME
		if record.Type == dnstypes.RecordTypeCNAME && strings.TrimRight(record.Value, ".") == fullName {
			err = manager.DeleteRecord(domain.Name, record)
			if err != nil {
				return err
			}
			isChanged = true
		}
	}

	if isChanged {
		err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, domainId, dnsmodels.DNSTaskTypeDomainChange)
		if err != nil {
			return err
		}
	}

	isOk = true

	return nil
}

func (this *DNSTaskExecutor) doDomain(domainId int64) error {
	return this.doDomainWithTask(0, domainId)
}

func (this *DNSTaskExecutor) doDomainWithTask(taskId int64, domainId int64) error {
	var tx *dbs.Tx

	isOk := false
	defer func() {
		if isOk {
			if taskId > 0 {
				err := dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskDone(tx, taskId)
				if err != nil {
					this.logErr("DNSTaskExecutor", err.Error())
				}
			}
		}
	}()

	dnsDomain, err := dnsmodels.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId, nil)
	if err != nil {
		return err
	}
	if dnsDomain == nil {
		isOk = true
		return nil
	}
	providerId := int64(dnsDomain.ProviderId)
	if providerId <= 0 {
		isOk = true
		return nil
	}

	provider, err := dnsmodels.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, providerId)
	if err != nil {
		return err
	}
	if provider == nil {
		isOk = true
		return nil
	}

	manager := dnsclients.FindProvider(provider.Type, int64(provider.Id))
	if manager == nil {
		this.logErr("DNSTaskExecutor", "unsupported dns provider type '"+provider.Type+"'")
		isOk = true
		return nil
	}
	params, err := provider.DecodeAPIParams()
	if err != nil {
		return err
	}
	err = manager.Auth(params)
	if err != nil {
		return err
	}
	records, err := manager.GetRecords(dnsDomain.Name)
	if err != nil {
		return err
	}
	recordsJSON, err := json.Marshal(records)
	if err != nil {
		return err
	}
	err = dnsmodels.SharedDNSDomainDAO.UpdateDomainRecords(tx, domainId, recordsJSON)
	if err != nil {
		return err
	}
	isOk = true
	return nil
}

func (this *DNSTaskExecutor) findDNSManagerWithClusterId(tx *dbs.Tx, clusterId int64) (manager dnsclients.ProviderInterface, domainId int64, domain string, clusterDNSName string, dnsConfig *dnsconfigs.ClusterDNSConfig, err error) {
	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId, nil)
	if err != nil {
		return nil, 0, "", "", nil, err
	}
	if clusterDNS == nil || len(clusterDNS.DnsName) == 0 || clusterDNS.DnsDomainId <= 0 {
		return nil, 0, "", "", nil, nil
	}

	dnsConfig, err = clusterDNS.DecodeDNSConfig()
	if err != nil {
		return nil, 0, "", "", nil, err
	}

	dnsDomain, manager, err := this.findDNSManagerWithDomainId(tx, int64(clusterDNS.DnsDomainId))
	if err != nil {
		return nil, 0, "", "", nil, err
	}

	if dnsDomain == nil {
		return nil, 0, "", clusterDNS.DnsName, dnsConfig, nil
	}

	return manager, int64(dnsDomain.Id), dnsDomain.Name, clusterDNS.DnsName, dnsConfig, nil
}

func (this *DNSTaskExecutor) findDNSManagerWithDomainId(tx *dbs.Tx, domainId int64) (*dnsmodels.DNSDomain, dnsclients.ProviderInterface, error) {
	dnsDomain, err := dnsmodels.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId, nil)
	if err != nil {
		return nil, nil, err
	}
	if dnsDomain == nil {
		return nil, nil, nil
	}
	providerId := int64(dnsDomain.ProviderId)
	if providerId <= 0 {
		return nil, nil, nil
	}

	provider, err := dnsmodels.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, providerId)
	if err != nil {
		return nil, nil, err
	}
	if provider == nil {
		return nil, nil, nil
	}

	var manager = dnsclients.FindProvider(provider.Type, int64(provider.Id))
	if manager == nil {
		this.logErr("DNSTaskExecutor", "unsupported dns provider type '"+provider.Type+"'")
		return nil, nil, nil
	}
	params, err := provider.DecodeAPIParams()
	if err != nil {
		return nil, nil, err
	}
	err = manager.Auth(params)
	if err != nil {
		return nil, nil, err
	}
	return dnsDomain, manager, nil
}
