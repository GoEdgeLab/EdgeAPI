package models

import (
	"bytes"
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/zero"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type HTTPAccessLogDAO dbs.DAO

var SharedHTTPAccessLogDAO *HTTPAccessLogDAO

// 队列
var (
	oldAccessLogQueue       = make(chan *pb.HTTPAccessLog)
	accessLogQueue          = make(chan *pb.HTTPAccessLog, 10_000)
	accessLogQueueMaxLength = 100_000
	accessLogQueuePercent   = 100    // 0-100
	accessLogCountPerSecond = 10_000 // 0 表示不限制
	accessLogConfigJSON     = []byte{}
	accessLogQueueChanged   = make(chan zero.Zero, 1)

	accessLogEnableAutoPartial       = true    // 是否启用自动分表
	accessLogRowsPerTable      int64 = 500_000 // 自动分表的单表最大值
)

type accessLogTableQuery struct {
	daoWrapper         *HTTPAccessLogDAOWrapper
	name               string
	hasRemoteAddrField bool
	hasDomainField     bool
}

func init() {
	dbs.OnReady(func() {
		SharedHTTPAccessLogDAO = NewHTTPAccessLogDAO()
	})

	// 队列相关
	dbs.OnReadyDone(func() {
		// 检查队列变化
		goman.New(func() {
			var ticker = time.NewTicker(60 * time.Second)

			// 先执行一次初始化
			SharedHTTPAccessLogDAO.SetupQueue()

			// 循环执行
			for {
				select {
				case <-ticker.C:
					SharedHTTPAccessLogDAO.SetupQueue()
				case <-accessLogQueueChanged:
					SharedHTTPAccessLogDAO.SetupQueue()
				}
			}
		})

		// 导出队列内容
		goman.New(func() {
			var ticker = time.NewTicker(1 * time.Second)
			for range ticker.C {
				var tx *dbs.Tx
				err := SharedHTTPAccessLogDAO.DumpAccessLogsFromQueue(tx, accessLogCountPerSecond)
				if err != nil {
					remotelogs.Error("HTTP_ACCESS_LOG_QUEUE", "dump access logs failed: "+err.Error())
				}
			}
		})
	})

}

func NewHTTPAccessLogDAO() *HTTPAccessLogDAO {
	return dbs.NewDAO(&HTTPAccessLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPAccessLogs",
			Model:  new(HTTPAccessLog),
			PkName: "id",
		},
	}).(*HTTPAccessLogDAO)
}

// CreateHTTPAccessLogs 创建访问日志
func (this *HTTPAccessLogDAO) CreateHTTPAccessLogs(tx *dbs.Tx, accessLogs []*pb.HTTPAccessLog) error {
	// 写入队列
	var queue = accessLogQueue // 这样写非常重要，防止在写入过程中队列有切换
	for _, accessLog := range accessLogs {
		if accessLog.FirewallPolicyId == 0 { // 如果是WAF记录，则采取采样率
			// 采样率
			if accessLogQueuePercent <= 0 {
				return nil
			}
			if accessLogQueuePercent < 100 && rands.Int(1, 100) > accessLogQueuePercent {
				return nil
			}
		}

		select {
		case queue <- accessLog:
		default:
			// 超出的丢弃
		}
	}

	return nil
}

// DumpAccessLogsFromQueue 从队列导入访问日志
func (this *HTTPAccessLogDAO) DumpAccessLogsFromQueue(tx *dbs.Tx, size int) error {
	var dao = randomHTTPAccessLogDAO()
	if dao == nil {
		dao = &HTTPAccessLogDAOWrapper{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}
	}

	if size <= 0 {
		size = 1_000_000
	}

	// 复制变量，防止中途改变
	var oldQueue = oldAccessLogQueue
	var newQueue = accessLogQueue

Loop:
	for i := 0; i < size; i++ {
		// old
		select {
		case accessLog := <-oldQueue:
			err := this.CreateHTTPAccessLog(tx, dao.DAO, accessLog)
			if err != nil {
				return err
			}
			continue Loop
		default:

		}

		// new
		select {
		case accessLog := <-newQueue:
			err := this.CreateHTTPAccessLog(tx, dao.DAO, accessLog)
			if err != nil {
				return err
			}
			continue Loop
		default:
			break Loop
		}
	}

	return nil
}

// CreateHTTPAccessLog 写入单条访问日志
func (this *HTTPAccessLogDAO) CreateHTTPAccessLog(tx *dbs.Tx, dao *HTTPAccessLogDAO, accessLog *pb.HTTPAccessLog) error {
	var day = timeutil.FormatTime("Ymd", accessLog.Timestamp)
	tableDef, err := SharedHTTPAccessLogManager.FindLastTable(dao.Instance, day, true)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{}
	fields["serverId"] = accessLog.ServerId
	fields["nodeId"] = accessLog.NodeId
	fields["status"] = accessLog.Status
	fields["createdAt"] = accessLog.Timestamp
	fields["requestId"] = accessLog.RequestId
	fields["firewallPolicyId"] = accessLog.FirewallPolicyId
	fields["firewallRuleGroupId"] = accessLog.FirewallRuleGroupId
	fields["firewallRuleSetId"] = accessLog.FirewallRuleSetId
	fields["firewallRuleId"] = accessLog.FirewallRuleId

	if len(accessLog.RequestBody) > 0 {
		fields["requestBody"] = accessLog.RequestBody
		accessLog.RequestBody = nil
	}

	if tableDef.HasRemoteAddr {
		fields["remoteAddr"] = accessLog.RemoteAddr
	}
	if tableDef.HasDomain {
		fields["domain"] = accessLog.Host
	}

	content, err := json.Marshal(accessLog)
	if err != nil {
		return err
	}
	fields["content"] = content

	var lastId int64
	lastId, err = dao.Query(tx).
		Table(tableDef.Name).
		Sets(fields).
		Insert()
	if err != nil {
		// 错误重试
		if CheckSQLErrCode(err, 1146) { // Error 1146: Table 'xxx' doesn't exist
			err = SharedHTTPAccessLogManager.CreateTable(dao.Instance, tableDef.Name)
			if err != nil {
				return err
			}

			// 重新尝试
			lastId, err = dao.Query(tx).
				Table(tableDef.Name).
				Sets(fields).
				Insert()
		}

		if err != nil {
			return err
		}
	}

	if accessLogEnableAutoPartial && accessLogRowsPerTable > 0 && lastId >= accessLogRowsPerTable {
		SharedHTTPAccessLogManager.ResetTable(dao.Instance, day)
	}

	return nil
}

// ListAccessLogs 读取往前的 单页访问日志
func (this *HTTPAccessLogDAO) ListAccessLogs(tx *dbs.Tx,
	partition int32,
	lastRequestId string,
	size int64,
	day string,
	hourFrom string,
	hourTo string,
	clusterId int64,
	nodeId int64,
	serverId int64,
	reverse bool,
	hasError bool,
	firewallPolicyId int64,
	firewallRuleGroupId int64,
	firewallRuleSetId int64,
	hasFirewallPolicy bool,
	userId int64,
	keyword string,
	ip string,
	domain string) (result []*HTTPAccessLog, nextLastRequestId string, hasMore bool, err error) {
	if len(day) != 8 {
		return
	}

	// 限制能查询的最大条数，防止占用内存过多
	if size > 1000 {
		size = 1000
	}

	result, nextLastRequestId, err = this.listAccessLogs(tx, partition, lastRequestId, size, day, hourFrom, hourTo, clusterId, nodeId, serverId, reverse, hasError, firewallPolicyId, firewallRuleGroupId, firewallRuleSetId, hasFirewallPolicy, userId, keyword, ip, domain)
	if err != nil || int64(len(result)) < size {
		return
	}

	moreResult, _, _ := this.listAccessLogs(tx, partition, nextLastRequestId, 1, day, hourFrom, hourTo, clusterId, nodeId, serverId, reverse, hasError, firewallPolicyId, firewallRuleGroupId, firewallRuleSetId, hasFirewallPolicy, userId, keyword, ip, domain)
	hasMore = len(moreResult) > 0
	return
}

// 读取往前的单页访问日志
func (this *HTTPAccessLogDAO) listAccessLogs(tx *dbs.Tx,
	partition int32,
	lastRequestId string,
	size int64,
	day string,
	hourFrom string,
	hourTo string,
	clusterId int64,
	nodeId int64,
	serverId int64,
	reverse bool,
	hasError bool,
	firewallPolicyId int64,
	firewallRuleGroupId int64,
	firewallRuleSetId int64,
	hasFirewallPolicy bool,
	userId int64,
	keyword string,
	ip string,
	domain string) (result []*HTTPAccessLog, nextLastRequestId string, err error) {
	if size <= 0 {
		return nil, lastRequestId, nil
	}

	var serverIds = []int64{}
	if userId > 0 {
		serverIds, err = SharedServerDAO.FindAllEnabledServerIdsWithUserId(tx, userId)
		if err != nil {
			return
		}
		if len(serverIds) == 0 {
			return
		}
	}

	accessLogLocker.RLock()
	var daoList = []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}

	// 查询某个集群下的节点
	var nodeIds = []int64{}
	if clusterId > 0 {
		nodeIds, err = SharedNodeDAO.FindAllEnabledNodeIdsWithClusterId(tx, clusterId)
		if err != nil {
			remotelogs.Error("DB_NODE", err.Error())
			return
		}
		sort.Slice(nodeIds, func(i, j int) bool {
			return nodeIds[i] < nodeIds[j]
		})
	}

	// 准备查询
	var tableQueries = []*accessLogTableQuery{}
	var maxTableName = ""
	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		def, err := SharedHTTPAccessLogManager.FindPartitionTable(instance, day, partition)
		if err != nil {
			return nil, "", err
		}
		if !def.Exists {
			continue
		}

		if len(maxTableName) == 0 || def.Name > maxTableName {
			maxTableName = def.Name
		}

		tableQueries = append(tableQueries, &accessLogTableQuery{
			daoWrapper:         daoWrapper,
			name:               def.Name,
			hasRemoteAddrField: def.HasRemoteAddr,
			hasDomainField:     def.HasDomain,
		})
	}

	// 检查各个分表是否一致
	if partition < 0 {
		var newTableQueries = []*accessLogTableQuery{}
		for _, tableQuery := range tableQueries {
			if tableQuery.name != maxTableName {
				continue
			}
			newTableQueries = append(newTableQueries, tableQuery)
		}
		tableQueries = newTableQueries
	}

	if len(tableQueries) == 0 {
		return nil, "", nil
	}

	var locker = sync.Mutex{}

	// 这里正则表达式中的括号不能轻易变更，因为后面有引用
	// TODO 支持多个查询条件的组合，比如 status:200 proto:HTTP/1.1
	var statusPrefixReg = regexp.MustCompile(`status:\s*(\d{3})\b`)
	var statusRangeReg = regexp.MustCompile(`status:\s*(\d{3})-(\d{3})\b`)
	var urlReg = regexp.MustCompile(`^(http|https)://`)
	var requestPathReg = regexp.MustCompile(`requestPath:(\S+)`)
	var protoReg = regexp.MustCompile(`proto:(\S+)`)
	var schemeReg = regexp.MustCompile(`scheme:(\S+)`)

	var count = len(tableQueries)
	var wg = &sync.WaitGroup{}
	wg.Add(count)
	for _, tableQuery := range tableQueries {
		go func(tableQuery *accessLogTableQuery, keyword string) {
			defer wg.Done()

			var dao = tableQuery.daoWrapper.DAO
			var query = dao.Query(tx)

			// 条件
			if nodeId > 0 {
				query.Attr("nodeId", nodeId)
			} else if clusterId > 0 {
				if len(nodeIds) > 0 {
					var nodeIdStrings = []string{}
					for _, subNodeId := range nodeIds {
						nodeIdStrings = append(nodeIdStrings, types.String(subNodeId))
					}

					query.Where("nodeId IN (" + strings.Join(nodeIdStrings, ",") + ")")
					query.Reuse(false)
				} else {
					// 如果没有节点，则直接返回空
					return
				}
			}
			if serverId > 0 {
				query.Attr("serverId", serverId)
			} else if userId > 0 && len(serverIds) > 0 {
				query.Attr("serverId", serverIds).
					Reuse(false)
			}
			if hasError {
				query.Where("status>=400")
			}
			if firewallPolicyId > 0 {
				query.Attr("firewallPolicyId", firewallPolicyId)
			}
			if firewallRuleGroupId > 0 {
				query.Attr("firewallRuleGroupId", firewallRuleGroupId)
			}
			if firewallRuleSetId > 0 {
				query.Attr("firewallRuleSetId", firewallRuleSetId)
			}
			if hasFirewallPolicy {
				query.Where("firewallPolicyId>0")
				query.UseIndex("firewallPolicyId")
			}

			// keyword
			if len(ip) > 0 {
				// TODO 支持IP范围
				if tableQuery.hasRemoteAddrField {
					// IP格式
					if strings.Contains(ip, ",") || strings.Contains(ip, "-") {
						rangeConfig, err := shared.ParseIPRange(ip)
						if err == nil {
							if len(rangeConfig.IPFrom) > 0 && len(rangeConfig.IPTo) > 0 {
								query.Between("INET_ATON(remoteAddr)", utils.IP2Long(rangeConfig.IPFrom), utils.IP2Long(rangeConfig.IPTo))
							}
						}
					} else {
						// 去掉IPv6的[]
						ip = strings.Trim(ip, "[]")

						query.Attr("remoteAddr", ip)
						query.UseIndex("remoteAddr")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.remoteAddr')=:ip1").
						Param("ip1", ip)
				}
			}
			if len(domain) > 0 {
				if tableQuery.hasDomainField {
					if strings.Contains(domain, "*") {
						domain = strings.ReplaceAll(domain, "*", "%")
						domain = regexp.MustCompile(`[^a-zA-Z0-9-.%]`).ReplaceAllString(domain, "")
						query.Where("domain LIKE :host2").
							Param("host2", domain)
					} else {
						query.Attr("domain", domain)
						query.UseIndex("domain")
					}
				} else {
					query.Where("JSON_EXTRACT(content, '$.host')=:host1").
						Param("host1", domain)
				}
			}

			if len(keyword) > 0 {
				var isSpecialKeyword = false

				if tableQuery.hasRemoteAddrField && net.ParseIP(keyword) != nil { // ip
					isSpecialKeyword = true
					query.Attr("remoteAddr", keyword)
				} else if tableQuery.hasRemoteAddrField && regexp.MustCompile(`^ip:.+`).MatchString(keyword) { // ip:x.x.x.x
					isSpecialKeyword = true
					keyword = keyword[3:]
					pieces := strings.SplitN(keyword, ",", 2)
					if len(pieces) == 1 || len(pieces[1]) == 0 || pieces[0] == pieces[1] {
						query.Attr("remoteAddr", pieces[0])
					} else {
						query.Between("INET_ATON(remoteAddr)", utils.IP2Long(pieces[0]), utils.IP2Long(pieces[1]))
					}
				} else if statusRangeReg.MatchString(keyword) { // status:200-400
					isSpecialKeyword = true
					var matches = statusRangeReg.FindStringSubmatch(keyword)
					query.Between("status", types.Int(matches[1]), types.Int(matches[2]))
				} else if statusPrefixReg.MatchString(keyword) { // status:200
					isSpecialKeyword = true
					var matches = statusPrefixReg.FindStringSubmatch(keyword)
					query.Attr("status", matches[1])
				} else if requestPathReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = requestPathReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.requestPath')=:keyword").
						Param("keyword", matches[1])
				} else if protoReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = protoReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.proto')=:keyword").
						Param("keyword", strings.ToUpper(matches[1]))
				} else if schemeReg.MatchString(keyword) {
					isSpecialKeyword = true
					var matches = schemeReg.FindStringSubmatch(keyword)
					query.Where("JSON_EXTRACT(content, '$.scheme')=:keyword").
						Param("keyword", strings.ToLower(matches[1]))
				} else if urlReg.MatchString(keyword) { // https://xxx/yyy
					u, err := url.Parse(keyword)
					if err == nil {
						isSpecialKeyword = true
						query.Attr("domain", u.Host)
						query.Where("JSON_EXTRACT(content, '$.requestURI') LIKE :keyword").
							Param("keyword", dbutils.QuoteLikePrefix("\""+u.RequestURI()))
					}
				}
				if !isSpecialKeyword {
					if regexp.MustCompile(`^ip:.+`).MatchString(keyword) {
						keyword = keyword[3:]
					}

					var useOriginKeyword = false

					where := "JSON_EXTRACT(content, '$.remoteAddr') LIKE :keyword OR JSON_EXTRACT(content, '$.requestURI') LIKE :keyword OR JSON_EXTRACT(content, '$.host') LIKE :keyword OR JSON_EXTRACT(content, '$.userAgent') LIKE :keyword"

					jsonKeyword, err := json.Marshal(keyword)
					if err == nil {
						where += " OR JSON_CONTAINS(content, :jsonKeyword, '$.tags')"
						query.Param("jsonKeyword", jsonKeyword)
					}

					// 请求方法
					if keyword == http.MethodGet ||
						keyword == http.MethodPost ||
						keyword == http.MethodHead ||
						keyword == http.MethodConnect ||
						keyword == http.MethodPut ||
						keyword == http.MethodTrace ||
						keyword == http.MethodOptions ||
						keyword == http.MethodDelete ||
						keyword == http.MethodPatch {
						where += " OR JSON_EXTRACT(content, '$.requestMethod')=:originKeyword"
						useOriginKeyword = true
					}

					// 响应状态码
					if regexp.MustCompile(`^\d{3}$`).MatchString(keyword) {
						where += " OR status=:intKeyword"
						query.Param("intKeyword", types.Int(keyword))
					}

					if regexp.MustCompile(`^\d{3}-\d{3}$`).MatchString(keyword) {
						pieces := strings.Split(keyword, "-")
						where += " OR status BETWEEN :intKeyword1 AND :intKeyword2"
						query.Param("intKeyword1", types.Int(pieces[0]))
						query.Param("intKeyword2", types.Int(pieces[1]))
					}

					if regexp.MustCompile(`^\d{20,}\s*\.?$`).MatchString(keyword) {
						where += " OR requestId=:requestId"
						query.Param("requestId", strings.TrimRight(keyword, ". "))
					}

					query.Where("("+where+")").
						Param("keyword", dbutils.QuoteLike(keyword))
					if useOriginKeyword {
						query.Param("originKeyword", keyword)
					}
				}
			}

			// hourFrom - hourTo
			if len(hourFrom) > 0 && len(hourTo) > 0 {
				var hourFromInt = types.Int(hourFrom)
				var hourToInt = types.Int(hourTo)
				if hourFromInt >= 0 && hourFromInt <= 23 && hourToInt >= hourFromInt && hourToInt <= 23 {
					var y = types.Int(day[:4])
					var m = types.Int(day[4:6])
					var d = types.Int(day[6:])
					var timeFrom = time.Date(y, time.Month(m), d, hourFromInt, 0, 0, 0, time.Local)
					var timeTo = time.Date(y, time.Month(m), d, hourToInt, 59, 59, 0, time.Local)
					query.Between("createdAt", timeFrom.Unix(), timeTo.Unix())
				}
			}

			// offset
			if len(lastRequestId) > 0 {
				if !reverse {
					query.Where("requestId<:requestId").
						Param("requestId", lastRequestId)
				} else {
					query.Where("requestId>:requestId").
						Param("requestId", lastRequestId)
				}
			}

			if !reverse {
				query.Desc("requestId")
			} else {
				query.Asc("requestId")
			}

			// 开始查询
			ones, err := query.
				Table(tableQuery.name).
				Limit(size).
				FindAll()
			if err != nil {
				remotelogs.Println("DB_NODE", err.Error())
				return
			}

			locker.Lock()
			for _, one := range ones {
				var accessLog = one.(*HTTPAccessLog)
				result = append(result, accessLog)
			}
			locker.Unlock()
		}(tableQuery, keyword)
	}
	wg.Wait()

	if len(result) == 0 {
		return nil, lastRequestId, nil
	}

	// 按照requestId排序
	sort.Slice(result, func(i, j int) bool {
		if !reverse {
			return result[i].RequestId > result[j].RequestId
		} else {
			return result[i].RequestId < result[j].RequestId
		}
	})

	if int64(len(result)) > size {
		result = result[:size]
	}

	var requestId = result[len(result)-1].RequestId
	if reverse {
		lists.Reverse(result)
	}

	if !reverse {
		return result, requestId, nil
	} else {
		return result, requestId, nil
	}
}

// FindAccessLogWithRequestId 根据请求ID获取访问日志
func (this *HTTPAccessLogDAO) FindAccessLogWithRequestId(tx *dbs.Tx, requestId string) (*HTTPAccessLog, error) {
	if !regexp.MustCompile(`^\d{11,}`).MatchString(requestId) {
		return nil, errors.New("invalid requestId")
	}

	accessLogLocker.RLock()
	daoList := []*HTTPAccessLogDAOWrapper{}
	for _, daoWrapper := range httpAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	if len(daoList) == 0 {
		daoList = []*HTTPAccessLogDAOWrapper{{
			DAO:    SharedHTTPAccessLogDAO,
			NodeId: 0,
		}}
	}

	// 准备查询
	var day = timeutil.FormatTime("Ymd", types.Int64(requestId[:10]))
	var tableQueries = []*accessLogTableQuery{}
	for _, daoWrapper := range daoList {
		var instance = daoWrapper.DAO.Instance
		tableDefs, err := SharedHTTPAccessLogManager.FindTables(instance, day)
		if err != nil {
			return nil, err
		}
		for _, def := range tableDefs {
			tableQueries = append(tableQueries, &accessLogTableQuery{
				daoWrapper:         daoWrapper,
				name:               def.Name,
				hasRemoteAddrField: def.HasRemoteAddr,
				hasDomainField:     def.HasDomain,
			})
		}
	}

	var count = len(tableQueries)
	var wg = &sync.WaitGroup{}
	wg.Add(count)
	var result *HTTPAccessLog = nil
	for _, tableQuery := range tableQueries {
		go func(tableQuery *accessLogTableQuery) {
			defer wg.Done()

			var dao = tableQuery.daoWrapper.DAO
			one, err := dao.Query(tx).
				Table(tableQuery.name).
				Attr("requestId", requestId).
				Find()
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			if one != nil {
				result = one.(*HTTPAccessLog)
			}
		}(tableQuery)
	}
	wg.Wait()
	return result, nil
}

// SetupQueue 建立队列
func (this *HTTPAccessLogDAO) SetupQueue() {
	configJSON, err := SharedSysSettingDAO.ReadSetting(nil, systemconfigs.SettingCodeAccessLogQueue)
	if err != nil {
		remotelogs.Error("HTTP_ACCESS_LOG_QUEUE", "read settings failed: "+err.Error())
		return
	}

	if len(configJSON) == 0 {
		return
	}

	if bytes.Compare(accessLogConfigJSON, configJSON) == 0 {
		return
	}
	accessLogConfigJSON = configJSON

	var config = &serverconfigs.AccessLogQueueConfig{}
	err = json.Unmarshal(configJSON, config)
	if err != nil {
		remotelogs.Error("HTTP_ACCESS_LOG_QUEUE", "decode settings failed: "+err.Error())
		return
	}

	accessLogQueuePercent = config.Percent
	accessLogCountPerSecond = config.CountPerSecond
	if config.MaxLength <= 0 {
		config.MaxLength = 100_000
	}

	accessLogEnableAutoPartial = config.EnableAutoPartial
	if config.RowsPerTable > 0 {
		accessLogRowsPerTable = config.RowsPerTable
	}

	if accessLogQueueMaxLength != config.MaxLength {
		accessLogQueueMaxLength = config.MaxLength
		oldAccessLogQueue = accessLogQueue
		accessLogQueue = make(chan *pb.HTTPAccessLog, config.MaxLength)
	}

	if Tea.IsTesting() {
		remotelogs.Println("HTTP_ACCESS_LOG_QUEUE", "change queue "+string(configJSON))
	}
}
