package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NSAccessLogDAO dbs.DAO

func NewNSAccessLogDAO() *NSAccessLogDAO {
	return dbs.NewDAO(&NSAccessLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSAccessLogs",
			Model:  new(NSAccessLog),
			PkName: "id",
		},
	}).(*NSAccessLogDAO)
}

var SharedNSAccessLogDAO *NSAccessLogDAO

func init() {
	dbs.OnReady(func() {
		SharedNSAccessLogDAO = NewNSAccessLogDAO()
	})
}

// CreateNSAccessLogs 创建访问日志
func (this *NSAccessLogDAO) CreateNSAccessLogs(tx *dbs.Tx, accessLogs []*pb.NSAccessLog) error {
	dao := randomNSAccessLogDAO()
	if dao == nil {
		dao = &NSAccessLogDAOWrapper{
			DAO:    SharedNSAccessLogDAO,
			NodeId: 0,
		}
	}
	return this.CreateNSAccessLogsWithDAO(tx, dao, accessLogs)
}

// CreateNSAccessLogsWithDAO 使用特定的DAO创建访问日志
func (this *NSAccessLogDAO) CreateNSAccessLogsWithDAO(tx *dbs.Tx, daoWrapper *NSAccessLogDAOWrapper, accessLogs []*pb.NSAccessLog) error {
	if daoWrapper == nil {
		return errors.New("dao should not be nil")
	}
	if len(accessLogs) == 0 {
		return nil
	}

	dao := daoWrapper.DAO

	// TODO 改成事务批量提交，以加快速度

	for _, accessLog := range accessLogs {
		day := timeutil.Format("Ymd", time.Unix(accessLog.Timestamp, 0))
		table, err := findNSAccessLogTable(dao.Instance, day, false)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{}
		fields["nodeId"] = accessLog.NsNodeId
		fields["domainId"] = accessLog.NsDomainId
		fields["recordId"] = accessLog.NsRecordId
		fields["createdAt"] = accessLog.Timestamp
		fields["requestId"] = accessLog.RequestId + strconv.FormatInt(time.Now().UnixNano(), 10) + configs.PaddingId

		content, err := json.Marshal(accessLog)
		if err != nil {
			return err
		}
		fields["content"] = content

		_, err = dao.Query(tx).
			Table(table).
			Sets(fields).
			Insert()
		if err != nil {
			// 是否为 Error 1146: Table 'xxx.xxx' doesn't exist  如果是，则创建表之后重试
			if strings.Contains(err.Error(), "1146") {
				table, err = findNSAccessLogTable(dao.Instance, day, true)
				if err != nil {
					return err
				}
				_, err = dao.Query(tx).
					Table(table).
					Sets(fields).
					Insert()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ListAccessLogs 读取往前的 单页访问日志
func (this *NSAccessLogDAO) ListAccessLogs(tx *dbs.Tx, lastRequestId string, size int64, day string, nodeId int64, domainId int64, recordId int64, keyword string, reverse bool) (result []*NSAccessLog, nextLastRequestId string, hasMore bool, err error) {
	if len(day) != 8 {
		return
	}

	// 限制能查询的最大条数，防止占用内存过多
	if size > 1000 {
		size = 1000
	}

	result, nextLastRequestId, err = this.listAccessLogs(tx, lastRequestId, size, day, nodeId, domainId, recordId, keyword, reverse)
	if err != nil || int64(len(result)) < size {
		return
	}

	moreResult, _, _ := this.listAccessLogs(tx, nextLastRequestId, 1, day, nodeId, domainId, recordId, keyword, reverse)
	hasMore = len(moreResult) > 0
	return
}

// 读取往前的单页访问日志
func (this *NSAccessLogDAO) listAccessLogs(tx *dbs.Tx, lastRequestId string, size int64, day string, nodeId int64, domainId int64, recordId int64, keyword string, reverse bool) (result []*NSAccessLog, nextLastRequestId string, err error) {
	if size <= 0 {
		return nil, lastRequestId, nil
	}

	accessLogLocker.RLock()
	daoList := []*NSAccessLogDAOWrapper{}
	for _, daoWrapper := range nsAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	if len(daoList) == 0 {
		daoList = []*NSAccessLogDAOWrapper{{
			DAO:    SharedNSAccessLogDAO,
			NodeId: 0,
		}}
	}

	locker := sync.Mutex{}

	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	for _, daoWrapper := range daoList {
		go func(daoWrapper *NSAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO

			tableName, exists, err := findNSAccessLogTableName(dao.Instance, day)
			if !exists {
				// 表格不存在则跳过
				return
			}
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}

			query := dao.Query(tx)

			// 条件
			if nodeId > 0 {
				query.Attr("nodeId", nodeId)
			}
			if domainId > 0 {
				query.Attr("domainId", domainId)
			}
			if recordId > 0 {
				query.Attr("recordId", recordId)
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

			// keyword
			if len(keyword) > 0 {
				query.Where("(JSON_EXTRACT(content, '$.remoteAddr') LIKE :keyword OR JSON_EXTRACT(content, '$.questionName') LIKE :keyword OR JSON_EXTRACT(content, '$.recordValue') LIKE :keyword)").
					Param("keyword", "%"+keyword+"%")
			}

			if !reverse {
				query.Desc("requestId")
			} else {
				query.Asc("requestId")
			}

			// 开始查询
			ones, err := query.
				Table(tableName).
				Limit(size).
				FindAll()
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			locker.Lock()
			for _, one := range ones {
				accessLog := one.(*NSAccessLog)
				result = append(result, accessLog)
			}
			locker.Unlock()
		}(daoWrapper)
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

	requestId := result[len(result)-1].RequestId
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
func (this *NSAccessLogDAO) FindAccessLogWithRequestId(tx *dbs.Tx, requestId string) (*NSAccessLog, error) {
	if !regexp.MustCompile(`^\d{30,}`).MatchString(requestId) {
		return nil, errors.New("invalid requestId")
	}

	accessLogLocker.RLock()
	daoList := []*NSAccessLogDAOWrapper{}
	for _, daoWrapper := range nsAccessLogDAOMapping {
		daoList = append(daoList, daoWrapper)
	}
	accessLogLocker.RUnlock()

	if len(daoList) == 0 {
		daoList = []*NSAccessLogDAOWrapper{{
			DAO:    SharedNSAccessLogDAO,
			NodeId: 0,
		}}
	}

	count := len(daoList)
	wg := &sync.WaitGroup{}
	wg.Add(count)
	var result *NSAccessLog = nil
	day := timeutil.FormatTime("Ymd", types.Int64(requestId[:10]))
	for _, daoWrapper := range daoList {
		go func(daoWrapper *NSAccessLogDAOWrapper) {
			defer wg.Done()

			dao := daoWrapper.DAO

			tableName, exists, err := findNSAccessLogTableName(dao.Instance, day)
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			if !exists {
				return
			}

			one, err := dao.Query(tx).
				Table(tableName).
				Attr("requestId", requestId).
				Find()
			if err != nil {
				logs.Println("[DB_NODE]" + err.Error())
				return
			}
			if one != nil {
				result = one.(*NSAccessLog)
			}
		}(daoWrapper)
	}
	wg.Wait()
	return result, nil
}
