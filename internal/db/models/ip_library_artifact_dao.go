package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/iplibrary"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	stringutil "github.com/iwind/TeaGo/utils/string"
)

const (
	IPLibraryArtifactStateEnabled  = 1 // 已启用
	IPLibraryArtifactStateDisabled = 0 // 已禁用
)

type IPLibraryArtifactDAO dbs.DAO

func NewIPLibraryArtifactDAO() *IPLibraryArtifactDAO {
	return dbs.NewDAO(&IPLibraryArtifactDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPLibraryArtifacts",
			Model:  new(IPLibraryArtifact),
			PkName: "id",
		},
	}).(*IPLibraryArtifactDAO)
}

var SharedIPLibraryArtifactDAO *IPLibraryArtifactDAO

func init() {
	dbs.OnReady(func() {
		SharedIPLibraryArtifactDAO = NewIPLibraryArtifactDAO()
	})
}

// EnableIPLibraryArtifact 启用条目
func (this *IPLibraryArtifactDAO) EnableIPLibraryArtifact(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPLibraryArtifactStateEnabled).
		Update()
	return err
}

// DisableIPLibraryArtifact 禁用条目
func (this *IPLibraryArtifactDAO) DisableIPLibraryArtifact(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPLibraryArtifactStateDisabled).
		Update()
	return err
}

// FindEnabledIPLibraryArtifact 查找启用中的条目
func (this *IPLibraryArtifactDAO) FindEnabledIPLibraryArtifact(tx *dbs.Tx, id int64) (*IPLibraryArtifact, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(IPLibraryArtifactStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*IPLibraryArtifact), err
}

// CreateArtifact 创建制品
func (this *IPLibraryArtifactDAO) CreateArtifact(tx *dbs.Tx, name string, fileId int64, libraryFileId int64, meta *iplibrary.Meta) (int64, error) {
	var op = NewIPLibraryArtifactOperator()
	op.Name = name
	op.FileId = fileId
	op.LibraryFileId = libraryFileId

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return 0, err
	}
	op.Meta = metaJSON
	op.State = IPLibraryArtifactStateEnabled

	var code = stringutil.Md5(utils.Sha1RandomString())[:8]
	meta.Code = code
	op.Code = code // 要比较短，方便识别

	return this.SaveInt64(tx, op)
}

// FindAllArtifacts 查找制品列表
func (this *IPLibraryArtifactDAO) FindAllArtifacts(tx *dbs.Tx) (result []*IPLibraryArtifact, err error) {
	_, err = this.Query(tx).
		State(IPLibraryArtifactStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindPublicArtifact 查找当前使用的制品
func (this *IPLibraryArtifactDAO) FindPublicArtifact(tx *dbs.Tx) (*IPLibraryArtifact, error) {
	one, err := this.Query(tx).
		State(IPLibraryArtifactStateEnabled).
		Attr("isPublic", true).
		Result("id", "fileId", "code").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*IPLibraryArtifact), nil
}

// UpdateArtifactPublic 使用某个制品
func (this *IPLibraryArtifactDAO) UpdateArtifactPublic(tx *dbs.Tx, artifactId int64, isPublic bool) error {
	// 取消使用
	if !isPublic {
		return this.Query(tx).
			Pk(artifactId).
			Set("isPublic", false).
			UpdateQuickly()
	}

	// 使用

	// 先取消别的
	err := this.Query(tx).
		Neq("id", artifactId).
		State(IPLibraryArtifactStateEnabled).
		Attr("isPublic", true).
		Set("isPublic", false).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.Query(tx).
		Pk(artifactId).
		Set("isPublic", true).
		UpdateQuickly()
}
