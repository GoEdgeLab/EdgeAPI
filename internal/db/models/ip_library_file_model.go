package models

import "github.com/iwind/TeaGo/dbs"

// IPLibraryFile IP库上传的文件
type IPLibraryFile struct {
	Id              uint64   `field:"id"`              // ID
	Name            string   `field:"name"`            // IP库名称
	FileId          uint64   `field:"fileId"`          // 原始文件ID
	Template        string   `field:"template"`        // 模板
	EmptyValues     dbs.JSON `field:"emptyValues"`     // 空值列表
	GeneratedFileId uint64   `field:"generatedFileId"` // 生成的文件ID
	GeneratedAt     uint64   `field:"generatedAt"`     // 生成时间
	IsFinished      bool     `field:"isFinished"`      // 是否已经完成
	Countries       dbs.JSON `field:"countries"`       // 国家/地区
	Provinces       dbs.JSON `field:"provinces"`       // 省份
	Cities          dbs.JSON `field:"cities"`          // 城市
	Towns           dbs.JSON `field:"towns"`           // 区县
	Providers       dbs.JSON `field:"providers"`       // ISP服务商
	Code            string   `field:"code"`            // 文件代号
	Password        string   `field:"password"`        // 密码
	CreatedAt       uint64   `field:"createdAt"`       // 上传时间
	State           uint8    `field:"state"`           // 状态
}

type IPLibraryFileOperator struct {
	Id              any // ID
	Name            any // IP库名称
	FileId          any // 原始文件ID
	Template        any // 模板
	EmptyValues     any // 空值列表
	GeneratedFileId any // 生成的文件ID
	GeneratedAt     any // 生成时间
	IsFinished      any // 是否已经完成
	Countries       any // 国家/地区
	Provinces       any // 省份
	Cities          any // 城市
	Towns           any // 区县
	Providers       any // ISP服务商
	Code            any // 文件代号
	Password        any // 密码
	CreatedAt       any // 上传时间
	State           any // 状态
}

func NewIPLibraryFileOperator() *IPLibraryFileOperator {
	return &IPLibraryFileOperator{}
}
