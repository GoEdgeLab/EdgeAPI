package posts

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	PostCategoryStateEnabled  = 1 // 已启用
	PostCategoryStateDisabled = 0 // 已禁用
)

type PostCategoryDAO dbs.DAO

func NewPostCategoryDAO() *PostCategoryDAO {
	return dbs.NewDAO(&PostCategoryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgePostCategories",
			Model:  new(PostCategory),
			PkName: "id",
		},
	}).(*PostCategoryDAO)
}

var SharedPostCategoryDAO *PostCategoryDAO

func init() {
	dbs.OnReady(func() {
		SharedPostCategoryDAO = NewPostCategoryDAO()
	})
}

// EnablePostCategory 启用条目
func (this *PostCategoryDAO) EnablePostCategory(tx *dbs.Tx, categoryId int64) error {
	_, err := this.Query(tx).
		Pk(categoryId).
		Set("state", PostCategoryStateEnabled).
		Update()
	return err
}

// DisablePostCategory 禁用条目
func (this *PostCategoryDAO) DisablePostCategory(tx *dbs.Tx, categoryId int64) error {
	_, err := this.Query(tx).
		Pk(categoryId).
		Set("state", PostCategoryStateDisabled).
		Update()
	return err
}

// FindEnabledPostCategory 查找启用中的条目
func (this *PostCategoryDAO) FindEnabledPostCategory(tx *dbs.Tx, categoryId int64) (*PostCategory, error) {
	result, err := this.Query(tx).
		Pk(categoryId).
		State(PostCategoryStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*PostCategory), err
}

// FindPostCategoryName 根据主键查找名称
func (this *PostCategoryDAO) FindPostCategoryName(tx *dbs.Tx, categoryId int64) (string, error) {
	return this.Query(tx).
		Pk(categoryId).
		Result("name").
		FindStringCol("")
}
