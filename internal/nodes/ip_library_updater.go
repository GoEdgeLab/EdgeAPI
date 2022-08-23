// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package nodes

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/iplibrary"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"io"
	"os"
	"time"
)

func init() {
	dbs.OnReady(func() {
		goman.New(func() {
			iplibrary.NewUpdater(NewIPLibraryUpdater(), 10*time.Minute).Start()
		})
	})
}

type IPLibraryUpdater struct {
}

func NewIPLibraryUpdater() *IPLibraryUpdater {
	return &IPLibraryUpdater{}
}

// DataDir 文件目录
func (this *IPLibraryUpdater) DataDir() string {
	// data/
	var dir = Tea.Root + "/data"
	stat, err := os.Stat(dir)
	if err == nil && stat.IsDir() {
		return dir
	}

	err = os.Mkdir(dir, 0666)
	if err == nil {
		return dir
	}

	remotelogs.Error("IP_LIBRARY_UPDATER", "create directory '"+dir+"' failed: "+err.Error())

	// 如果不能创建 data/ 目录，那么使用临时目录
	return os.TempDir()
}

// FindLatestFile 检查最新的IP库文件
func (this *IPLibraryUpdater) FindLatestFile() (code string, fileId int64, err error) {
	artifact, err := models.SharedIPLibraryArtifactDAO.FindPublicArtifact(nil)
	if err != nil {
		return "", 0, err
	}
	if artifact == nil {
		return "", 0, nil
	}

	return artifact.Code, int64(artifact.FileId), nil
}

// DownloadFile 下载文件
func (this *IPLibraryUpdater) DownloadFile(fileId int64, writer io.Writer) error {
	if fileId <= 0 {
		return errors.New("invalid fileId: " + types.String(fileId))
	}

	var tx *dbs.Tx
	chunkIds, err := models.SharedFileChunkDAO.FindAllFileChunkIds(tx, fileId)
	if err != nil {
		return err
	}
	for _, chunkId := range chunkIds {
		chunk, err := models.SharedFileChunkDAO.FindFileChunk(tx, chunkId)
		if err != nil {
			return err
		}
		if chunk == nil {
			return errors.New("can not find file chunk with chunk id '" + types.String(chunkId) + "'")
		}
		_, err = writer.Write(chunk.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

// LogInfo 普通日志
func (this *IPLibraryUpdater) LogInfo(message string) {
	remotelogs.Println("IP_LIBRARY_UPDATER", message)
}

// LogError 错误日志
func (this *IPLibraryUpdater) LogError(err error) {
	if err == nil {
		return
	}
	remotelogs.Error("IP_LIBRARY_UPDATER", err.Error())
}
