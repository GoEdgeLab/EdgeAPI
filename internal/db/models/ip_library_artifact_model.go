package models

import "github.com/iwind/TeaGo/dbs"

// IPLibraryArtifact IP库制品
type IPLibraryArtifact struct {
	Id            uint32   `field:"id"`            // ID
	Name          string   `field:"name"`          // 名称
	FileId        uint64   `field:"fileId"`        // 文件ID
	LibraryFileId uint32   `field:"libraryFileId"` // IP库文件ID
	CreatedAt     uint64   `field:"createdAt"`     // 创建时间
	Meta          dbs.JSON `field:"meta"`          // 元数据
	IsPublic      bool     `field:"isPublic"`      // 是否为公用
	Code          string   `field:"code"`          // 代号
	State         uint8    `field:"state"`         // 状态
}

type IPLibraryArtifactOperator struct {
	Id            any // ID
	Name          any // 名称
	FileId        any // 文件ID
	LibraryFileId any // IP库文件ID
	CreatedAt     any // 创建时间
	Meta          any // 元数据
	IsPublic      any // 是否为公用
	Code          any // 代号
	State         any // 状态
}

func NewIPLibraryArtifactOperator() *IPLibraryArtifactOperator {
	return &IPLibraryArtifactOperator{}
}
