package posts

import "github.com/iwind/TeaGo/dbs"

const (
	PostField_Id          dbs.FieldName = "id"          // ID
	PostField_CategoryId  dbs.FieldName = "categoryId"  // 文章分类
	PostField_Type        dbs.FieldName = "type"        // 类型：normal, url
	PostField_Url         dbs.FieldName = "url"         // URL
	PostField_Subject     dbs.FieldName = "subject"     // 标题
	PostField_Body        dbs.FieldName = "body"        // 内容
	PostField_CreatedAt   dbs.FieldName = "createdAt"   // 创建时间
	PostField_IsPublished dbs.FieldName = "isPublished" // 是否已发布
	PostField_PublishedAt dbs.FieldName = "publishedAt" // 发布时间
	PostField_ProductCode dbs.FieldName = "productCode" // 产品代号
	PostField_State       dbs.FieldName = "state"       // 状态
)

// Post 文章管理
type Post struct {
	Id          uint32 `field:"id"`          // ID
	CategoryId  uint32 `field:"categoryId"`  // 文章分类
	Type        string `field:"type"`        // 类型：normal, url
	Url         string `field:"url"`         // URL
	Subject     string `field:"subject"`     // 标题
	Body        string `field:"body"`        // 内容
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	IsPublished bool   `field:"isPublished"` // 是否已发布
	PublishedAt uint64 `field:"publishedAt"` // 发布时间
	ProductCode string `field:"productCode"` // 产品代号
	State       uint8  `field:"state"`       // 状态
}

type PostOperator struct {
	Id          any // ID
	CategoryId  any // 文章分类
	Type        any // 类型：normal, url
	Url         any // URL
	Subject     any // 标题
	Body        any // 内容
	CreatedAt   any // 创建时间
	IsPublished any // 是否已发布
	PublishedAt any // 发布时间
	ProductCode any // 产品代号
	State       any // 状态
}

func NewPostOperator() *PostOperator {
	return &PostOperator{}
}
