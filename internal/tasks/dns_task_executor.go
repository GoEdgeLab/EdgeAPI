package tasks

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	dnsmodels "github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"net"
	"strings"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		go NewDNSTaskExecutor().Start()
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

	manager, domainId, domain, clusterDNSName, err := this.findDNSManager(tx, int64(serverDNS.ClusterId))
	if err != nil {
		return err
	}
	if manager == nil {
		isOk = true
		return nil
	}

	recordName := serverDNS.DnsName
	recordValue := clusterDNSName + "." + domain + "."
	recordRoute := manager.DefaultRoute()
	recordType := dnsclients.RecordTypeCName
	if serverDNS.State == models.ServerStateDisabled || serverDNS.IsOn == 0 {
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

		err = manager.AddRecord(domain, &dnsclients.Record{
			Id:    "",
			Name:  recordName,
			Type:  recordType,
			Value: recordValue,
			Route: recordRoute,
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

	if node.ClusterId == 0 {
		isOk = true
		return nil
	}

	// 转交给cluster统一处理
	err = dnsmodels.SharedDNSTaskDAO.CreateClusterTask(tx, int64(node.ClusterId), dnsmodels.DNSTaskTypeClusterChange)
	if err != nil {
		return err
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
	manager, domainId, domain, clusterDNSName, err := this.findDNSManager(tx, clusterId)
	if err != nil {
		return err
	}
	if manager == nil {
		isOk = true
		return nil
	}

	// 以前的节点记录
	records, err := manager.GetRecords(domain)
	if err != nil {
		return err
	}
	oldRecordsMap := map[string]*dnsclients.Record{} // route@value => record
	for _, record := range records {
		if record.Type == dnsclients.RecordTypeA && record.Name == clusterDNSName {
			key := record.Route + "@" + record.Value
			oldRecordsMap[key] = record
		}
	}

	// 当前的节点记录
	newRecordKeys := []string{}
	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(tx, clusterId)
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
		ipAddresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(tx, int64(node.Id))
		if err != nil {
			return err
		}
		if len(ipAddresses) == 0 {
			continue
		}
		for _, ipAddress := range ipAddresses {
			ip := ipAddress.Ip
			if len(ip) == 0 {
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

				err = manager.AddRecord(domain, &dnsclients.Record{
					Id:    "",
					Name:  clusterDNSName,
					Type:  dnsclients.RecordTypeA,
					Value: ip,
					Route: route,
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

	dnsDomain, err := dnsmodels.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId)
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

func (this *DNSTaskExecutor) findDNSManager(tx *dbs.Tx, clusterId int64) (manager dnsclients.ProviderInterface, domainId int64, domain string, clusterDNSName string, err error) {
	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId)
	if err != nil {
		return nil, 0, "", "", err
	}
	if clusterDNS == nil || len(clusterDNS.DnsName) == 0 || clusterDNS.DnsDomainId <= 0 {
		return nil, 0, "", "", nil
	}

	dnsDomain, err := dnsmodels.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, int64(clusterDNS.DnsDomainId))
	if err != nil {
		return nil, 0, "", "", err
	}
	if dnsDomain == nil {
		return nil, 0, "", "", nil
	}
	providerId := int64(dnsDomain.ProviderId)
	if providerId <= 0 {
		return nil, 0, "", "", nil
	}

	provider, err := dnsmodels.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, providerId)
	if err != nil {
		return nil, 0, "", "", err
	}
	if provider == nil {
		return nil, 0, "", "", nil
	}

	manager = dnsclients.FindProvider(provider.Type)
	if manager == nil {
		remotelogs.Error("DNSTaskExecutor", "unsupported dns provider type '"+provider.Type+"'")
		return nil, 0, "", "", nil
	}
	params, err := provider.DecodeAPIParams()
	if err != nil {
		return nil, 0, "", "", err
	}
	err = manager.Auth(params)
	if err != nil {
		return nil, 0, "", "", err
	}
	return manager, int64(dnsDomain.Id), dnsDomain.Name, clusterDNS.DnsName, nil
}
