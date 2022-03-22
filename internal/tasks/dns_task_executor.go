package tasks

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	dnsmodels "github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
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
			NewDNSTaskExecutor().Start()
		})
	})
}

// DNSTaskExecutor DNS任务执行器
type DNSTaskExecutor struct {
}

func NewDNSTaskExecutor() *DNSTaskExecutor {
	return &DNSTaskExecutor{}
}

func (this *DNSTaskExecutor) Start() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		err := this.LoopWithLocker(10)
		if err != nil {
			remotelogs.Error("DNSTaskExecutor", err.Error())
		}
	}
}

func (this *DNSTaskExecutor) LoopWithLocker(seconds int64) error {
	ok, err := models.SharedSysLockerDAO.Lock(nil, "dns_task_executor", seconds-1) // 假设执行时间为1秒
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return this.Loop()
}

func (this *DNSTaskExecutor) Loop() error {
	tasks, err := dnsmodels.SharedDNSTaskDAO.FindAllDoingTasks(nil)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		taskId := int64(task.Id)
		switch task.Type {
		case dnsmodels.DNSTaskTypeServerChange:
			err = this.doServer(taskId, int64(task.ServerId))
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
		case dnsmodels.DNSTaskTypeDomainChange:
			err = this.doDomain(taskId, int64(task.DomainId))
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
func (this *DNSTaskExecutor) doServer(taskId int64, serverId int64) error {
	var tx *dbs.Tx

	isOk := false
	defer func() {
		if isOk {
			err := dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskDone(tx, taskId)
			if err != nil {
				remotelogs.Error("DNSTaskExecutor", err.Error())
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

	manager, domainId, domain, clusterDNSName, dnsConfig, err := this.findDNSManager(tx, int64(serverDNS.ClusterId))
	if err != nil {
		return err
	}
	if manager == nil {
		isOk = true
		return nil
	}
	var ttl int32 = 0
	if dnsConfig != nil {
		ttl = dnsConfig.TTL
	}

	recordName := serverDNS.DnsName
	recordValue := clusterDNSName + "." + domain + "."
	recordRoute := manager.DefaultRoute()
	recordType := dnstypes.RecordTypeCNAME
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
			err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, domainId, dnsmodels.DNSTaskTypeDomainChange)
			if err != nil {
				return err
			}
		}

		isOk = true
	} else {
		// 是否已存在
		exist, err := dnsmodels.SharedDNSDomainDAO.ExistDomainRecord(tx, domainId, recordName, recordType, recordRoute, recordValue)
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
			err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, domainId, dnsmodels.DNSTaskTypeDomainChange)
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

		err = dnsmodels.SharedDNSTaskDAO.CreateDomainTask(tx, domainId, dnsmodels.DNSTaskTypeDomainChange)
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
				remotelogs.Error("DNSTaskExecutor", err.Error())
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
				remotelogs.Error("DNSTaskExecutor", err.Error())
			}
		}
	}()

	var tx *dbs.Tx
	manager, domainId, domain, clusterDNSName, dnsConfig, err := this.findDNSManager(tx, clusterId)
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
	newRecordKeys := []string{}
	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(tx, clusterId, true)
	if err != nil {
		return err
	}
	isChanged := false
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
			ip := ipAddress.DNSIP()
			if len(ip) == 0 || !ipAddress.CanAccess || !ipAddress.IsUp || !ipAddress.IsOn {
				continue
			}
			if net.ParseIP(ip) == nil {
				continue
			}
			for _, route := range routes {
				key := route + "@" + ip
				_, ok := oldRecordsMap[key]
				if ok {
					newRecordKeys = append(newRecordKeys, key)
					continue
				}

				recordType := dnstypes.RecordTypeA
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
		cnameRecords = dnsConfig.CNameRecords
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

func (this *DNSTaskExecutor) doDomain(taskId int64, domainId int64) error {
	var tx *dbs.Tx

	isOk := false
	defer func() {
		if isOk {
			err := dnsmodels.SharedDNSTaskDAO.UpdateDNSTaskDone(tx, taskId)
			if err != nil {
				remotelogs.Error("DNSTaskExecutor", err.Error())
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

	manager := dnsclients.FindProvider(provider.Type)
	if manager == nil {
		remotelogs.Error("DNSTaskExecutor", "unsupported dns provider type '"+provider.Type+"'")
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

func (this *DNSTaskExecutor) findDNSManager(tx *dbs.Tx, clusterId int64) (manager dnsclients.ProviderInterface, domainId int64, domain string, clusterDNSName string, dnsConfig *dnsconfigs.ClusterDNSConfig, err error) {
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

	dnsDomain, err := dnsmodels.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, int64(clusterDNS.DnsDomainId), nil)
	if err != nil {
		return nil, 0, "", "", nil, err
	}
	if dnsDomain == nil {
		return nil, 0, "", "", nil, nil
	}
	providerId := int64(dnsDomain.ProviderId)
	if providerId <= 0 {
		return nil, 0, "", "", nil, nil
	}

	provider, err := dnsmodels.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, providerId)
	if err != nil {
		return nil, 0, "", "", nil, err
	}
	if provider == nil {
		return nil, 0, "", "", nil, nil
	}

	manager = dnsclients.FindProvider(provider.Type)
	if manager == nil {
		remotelogs.Error("DNSTaskExecutor", "unsupported dns provider type '"+provider.Type+"'")
		return nil, 0, "", "", nil, nil
	}
	params, err := provider.DecodeAPIParams()
	if err != nil {
		return nil, 0, "", "", nil, err
	}
	err = manager.Auth(params)
	if err != nil {
		return nil, 0, "", "", nil, err
	}

	return manager, int64(dnsDomain.Id), dnsDomain.Name, clusterDNS.DnsName, dnsConfig, nil
}
