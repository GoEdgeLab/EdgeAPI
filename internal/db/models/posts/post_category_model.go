package posts

import "github.com/iwind/TeaGo/dbs"

const (
	PostCategoryField_Id    dbs.FieldName = "id"    // ID
	PostCategoryField_Name  dbs.FieldName = "name"  // 分类名称
	PostCategoryField_IsOn  dbs.FieldName = "isOn"  // 是否启用
	PostCategoryField_Code  dbs.FieldName = "code"  // 代号
	PostCategoryField_Order dbs.FieldName = "order" // 排序
	PostCategoryField_State dbs.FieldName = "state" // 分类状态
)

// PostCategory 文章分类
type PostCategory struct {
	Id    uint32 `field:"id"`    // ID
	Name  string `field:"name"`  // 分类名称
	IsOn  bool   `field:"isOn"`  // 是否启用
	Code  string `field:"code"`  // 代号
	Order uint32 `field:"order"` // 排序
	State uint8  `field:"state"` // 分类状态
}

type PostCategoryOperator struct {
	Id    any // ID
	Name  any // 分类名称
	IsOn  any // 是否启用
	Code  any // 代号
	Order any // 排序
	State any // 分类状态
}

func NewPostCategoryOperator() *PostCategoryOperator {
	return &PostCategoryOperator{}
}
