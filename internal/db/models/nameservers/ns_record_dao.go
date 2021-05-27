package nameservers

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSRecordStateEnabled  = 1 // 已启用
	NSRecordStateDisabled = 0 // 已禁用
)

type NSRecordDAO dbs.DAO

func NewNSRecordDAO() *NSRecordDAO {
	return dbs.NewDAO(&NSRecordDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSRecords",
			Model:  new(NSRecord),
			PkName: "id",
		},
	}).(*NSRecordDAO)
}

var SharedNSRecordDAO *NSRecordDAO

func init() {
	dbs.OnReady(func() {
		SharedNSRecordDAO = NewNSRecordDAO()
	})
}

// EnableNSRecord 启用条目
func (this *NSRecordDAO) EnableNSRecord(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSRecordStateEnabled).
		Update()
	return err
}

// DisableNSRecord 禁用条目
func (this *NSRecordDAO) DisableNSRecord(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSRecordStateDisabled).
		Update()
	return err
}

// FindEnabledNSRecord 查找启用中的条目
func (this *NSRecordDAO) FindEnabledNSRecord(tx *dbs.Tx, id int64) (*NSRecord, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSRecordStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSRecord), err
}

// FindNSRecordName 根据主键查找名称
func (this *NSRecordDAO) FindNSRecordName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateRecord 创建记录
func (this *NSRecordDAO) CreateRecord(tx *dbs.Tx, domainId int64, description string, name string, dnsType dnsconfigs.RecordType, value string, ttl int32, routeIds []int64) (int64, error) {
	op := NewNSRecordOperator()
	op.DomainId = domainId
	op.Description = description
	op.Name = name
	op.Type = dnsType
	op.Value = value
	op.Ttl = ttl

	if len(routeIds) == 0 {
		op.RouteIds = "[]"
	} else {
		routeIds, err := json.Marshal(routeIds)
		if err != nil {
			return 0, err
		}
		op.RouteIds = routeIds
	}

	op.IsOn = true
	op.State = NSRecordStateEnabled
	return this.SaveInt64(tx, op)
}

func (this *NSRecordDAO) UpdateRecord(tx *dbs.Tx, recordId int64, description string, name string, dnsType dnsconfigs.RecordType, value string, ttl int32, routeIds []int64) error {
	if recordId <= 0 {
		return errors.New("invalid recordId")
	}

	op := NewNSRecordOperator()
	op.Id = recordId
	op.Description = description
	op.Name = name
	op.Type = dnsType
	op.Value = value
	op.Ttl = ttl

	if len(routeIds) == 0 {
		op.RouteIds = "[]"
	} else {
		routeIds, err := json.Marshal(routeIds)
		if err != nil {
			return err
		}
		op.RouteIds = routeIds
	}

	return this.Save(tx, op)
}

func (this *NSRecordDAO) CountAllEnabledRecords(tx *dbs.Tx, domainId int64, dnsType dnsconfigs.RecordType, keyword string, routeId int64) (int64, error) {
	query := this.Query(tx).
		Attr("domainId", domainId).
		State(NSRecordStateEnabled)
	if len(dnsType) > 0 {
		query.Attr("type", dnsType)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	if routeId > 0 {
		query.JSONContains("routeIds", routeId)
	}
	return query.Count()
}

func (this *NSRecordDAO) ListAllEnabledRecords(tx *dbs.Tx, domainId int64, dnsType dnsconfigs.RecordType, keyword string, routeId int64, offset int64, size int64) (result []*NSRecord, err error) {
	query := this.Query(tx).
		Attr("domainId", domainId).
		State(NSRecordStateEnabled)
	if len(dnsType) > 0 {
		query.Attr("type", dnsType)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	if routeId > 0 {
		query.JSONContains("routeIds", routeId)
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}
