package nameservers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSKeyStateEnabled  = 1 // 已启用
	NSKeyStateDisabled = 0 // 已禁用
)

type NSKeyDAO dbs.DAO

func NewNSKeyDAO() *NSKeyDAO {
	return dbs.NewDAO(&NSKeyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSKeys",
			Model:  new(NSKey),
			PkName: "id",
		},
	}).(*NSKeyDAO)
}

var SharedNSKeyDAO *NSKeyDAO

func init() {
	dbs.OnReady(func() {
		SharedNSKeyDAO = NewNSKeyDAO()
	})
}

// EnableNSKey 启用条目
func (this *NSKeyDAO) EnableNSKey(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSKeyStateEnabled).
		Update()
	return err
}

// DisableNSKey 禁用条目
func (this *NSKeyDAO) DisableNSKey(tx *dbs.Tx, keyId int64) error {
	_, err := this.Query(tx).
		Pk(keyId).
		Set("state", NSKeyStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, keyId)
}

// FindEnabledNSKey 查找启用中的条目
func (this *NSKeyDAO) FindEnabledNSKey(tx *dbs.Tx, id int64) (*NSKey, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSKeyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSKey), err
}

// FindNSKeyName 根据主键查找名称
func (this *NSKeyDAO) FindNSKeyName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateKey 创建Key
func (this *NSKeyDAO) CreateKey(tx *dbs.Tx, domainId int64, zoneId int64, name string, algo dnsconfigs.KeyAlgorithmType, secret string, secretType string) (int64, error) {
	op := NewNSKeyOperator()
	op.DomainId = domainId
	op.ZoneId = zoneId
	op.Name = name
	op.Algo = algo
	op.Secret = secret
	op.SecretType = secretType
	op.State = NSKeyStateEnabled
	keyId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = this.NotifyUpdate(tx, keyId)
	if err != nil {
		return keyId, err
	}

	return keyId, nil
}

// UpdateKey 修改Key
func (this *NSKeyDAO) UpdateKey(tx *dbs.Tx, keyId int64, name string, algo dnsconfigs.KeyAlgorithmType, secret string, secretType string, isOn bool) error {
	if keyId <= 0 {
		return errors.New("invalid keyId")
	}
	op := NewNSKeyOperator()
	op.Id = keyId
	op.Name = name
	op.Algo = algo
	op.Secret = secret
	op.SecretType = secretType
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, keyId)
}

// CountEnabledKeys 计算Key的数量
func (this *NSKeyDAO) CountEnabledKeys(tx *dbs.Tx, domainId int64, zoneId int64) (int64, error) {
	var query = this.Query(tx).
		State(NSKeyStateEnabled)
	if domainId > 0 {
		query.Attr("domainId", domainId)
	}
	if zoneId > 0 {
		query.Attr("zoneId", zoneId)
	}
	return query.Count()
}

// ListEnabledKeys 列出单页Key
func (this *NSKeyDAO) ListEnabledKeys(tx *dbs.Tx, domainId int64, zoneId int64, offset int64, size int64) (result []*NSKey, err error) {
	var query = this.Query(tx).
		State(NSKeyStateEnabled)
	if domainId > 0 {
		query.Attr("domainId", domainId)
	}
	if zoneId > 0 {
		query.Attr("zoneId", zoneId)
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// IncreaseVersion 增加版本
func (this *NSKeyDAO) IncreaseVersion(tx *dbs.Tx) (int64, error) {
	return models.SharedSysLockerDAO.Increase(tx, "NS_KEY_VERSION", 1)
}

// ListKeysAfterVersion 列出某个版本后的密钥
func (this *NSKeyDAO) ListKeysAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*NSKey, err error) {
	if size <= 0 {
		size = 10000
	}

	_, err = this.Query(tx).
		Gte("version", version).
		Limit(size).
		Asc("version").
		Slice(&result).
		FindAll()
	return
}

// NotifyUpdate 通知更新
func (this *NSKeyDAO) NotifyUpdate(tx *dbs.Tx, keyId int64) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(keyId).
		Set("version", version).
		UpdateQuickly()
	if err != nil {
		return err
	}

	// 通知集群
	domainId, err := this.Query(tx).
		Pk(keyId).
		Result("domainId").
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if domainId > 0 {
		clusterId, err := SharedNSDomainDAO.FindEnabledDomainClusterId(tx, domainId)
		if err != nil {
			return err
		}
		if clusterId > 0 {
			err = models.SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleDNS, clusterId, models.NSNodeTaskTypeKeyChanged)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
