package authority

import "github.com/iwind/TeaGo/dbs"

const (
	AuthorityKeyField_Id           dbs.FieldName = "id"           // ID
	AuthorityKeyField_Value        dbs.FieldName = "value"        // Key值
	AuthorityKeyField_DayFrom      dbs.FieldName = "dayFrom"      // 开始日期
	AuthorityKeyField_DayTo        dbs.FieldName = "dayTo"        // 结束日期
	AuthorityKeyField_Hostname     dbs.FieldName = "hostname"     // Hostname
	AuthorityKeyField_MacAddresses dbs.FieldName = "macAddresses" // MAC地址
	AuthorityKeyField_UpdatedAt    dbs.FieldName = "updatedAt"    // 创建/修改时间
	AuthorityKeyField_Company      dbs.FieldName = "company"      // 公司组织
	AuthorityKeyField_RequestCode  dbs.FieldName = "requestCode"  // 申请码
)

// AuthorityKey 企业版认证信息
type AuthorityKey struct {
	Id           uint32   `field:"id"`           // ID
	Value        string   `field:"value"`        // Key值
	DayFrom      string   `field:"dayFrom"`      // 开始日期
	DayTo        string   `field:"dayTo"`        // 结束日期
	Hostname     string   `field:"hostname"`     // Hostname
	MacAddresses dbs.JSON `field:"macAddresses"` // MAC地址
	UpdatedAt    uint64   `field:"updatedAt"`    // 创建/修改时间
	Company      string   `field:"company"`      // 公司组织
	RequestCode  string   `field:"requestCode"`  // 申请码
}

type AuthorityKeyOperator struct {
	Id           any // ID
	Value        any // Key值
	DayFrom      any // 开始日期
	DayTo        any // 结束日期
	Hostname     any // Hostname
	MacAddresses any // MAC地址
	UpdatedAt    any // 创建/修改时间
	Company      any // 公司组织
	RequestCode  any // 申请码
}

func NewAuthorityKeyOperator() *AuthorityKeyOperator {
	return &AuthorityKeyOperator{}
}
