package posts

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	PostStateEnabled  = 1 // 已启用
	PostStateDisabled = 0 // 已禁用
)

type PostDAO dbs.DAO

func NewPostDAO() *PostDAO {
	return dbs.NewDAO(&PostDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgePosts",
			Model:  new(Post),
			PkName: "id",
		},
	}).(*PostDAO)
}

var SharedPostDAO *PostDAO

func init() {
	dbs.OnReady(func() {
		SharedPostDAO = NewPostDAO()
	})
}

// EnablePost 启用条目
func (this *PostDAO) EnablePost(tx *dbs.Tx, postId int64) error {
	_, err := this.Query(tx).
		Pk(postId).
		Set("state", PostStateEnabled).
		Update()
	return err
}

// DisablePost 禁用条目
func (this *PostDAO) DisablePost(tx *dbs.Tx, postId int64) error {
	_, err := this.Query(tx).
		Pk(postId).
		Set("state", PostStateDisabled).
		Update()
	return err
}

// FindEnabledPost 查找启用中的条目
func (this *PostDAO) FindEnabledPost(tx *dbs.Tx, postId int64) (*Post, error) {
	result, err := this.Query(tx).
		Pk(postId).
		State(PostStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Post), err
}
