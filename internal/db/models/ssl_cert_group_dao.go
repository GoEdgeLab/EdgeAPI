package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	SSLCertGroupStateEnabled  = 1 // 已启用
	SSLCertGroupStateDisabled = 0 // 已禁用
)

type SSLCertGroupDAO dbs.DAO

func NewSSLCertGroupDAO() *SSLCertGroupDAO {
	return dbs.NewDAO(&SSLCertGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeSSLCertGroups",
			Model:  new(SSLCertGroup),
			PkName: "id",
		},
	}).(*SSLCertGroupDAO)
}

var SharedSSLCertGroupDAO *SSLCertGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedSSLCertGroupDAO = NewSSLCertGroupDAO()
	})
}

// 启用条目
func (this *SSLCertGroupDAO) EnableSSLCertGroup(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", SSLCertGroupStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *SSLCertGroupDAO) DisableSSLCertGroup(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", SSLCertGroupStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *SSLCertGroupDAO) FindEnabledSSLCertGroup(id uint32) (*SSLCertGroup, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", SSLCertGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*SSLCertGroup), err
}

// 根据主键查找名称
func (this *SSLCertGroupDAO) FindSSLCertGroupName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
