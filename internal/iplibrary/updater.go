package iplibrary

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"os"
	"time"
)

func init() {
	dbs.OnReady(func() {
		updater := NewUpdater()
		updater.Start()
	})
}

// Updater IP库更新程序
type Updater struct {
}

// NewUpdater 获取新对象
func NewUpdater() *Updater {
	return &Updater{}
}

// Start 开始更新
func (this *Updater) Start() {
	// 这里不需要太频繁检查更新，因为通常不需要更新IP库
	ticker := time.NewTicker(1 * time.Hour)
	goman.New(func() {
		for range ticker.C {
			err := this.loop()
			if err != nil {
				logs.Println("[IP_LIBRARY]" + err.Error())
			}
		}
	})
}

// 单次任务
func (this *Updater) loop() error {
	config, err := models.SharedSysSettingDAO.ReadGlobalConfig(nil)
	if err != nil {
		return err
	}
	code := config.IPLibrary.Code
	if len(code) == 0 {
		code = serverconfigs.DefaultIPLibraryType
	}
	lib, err := models.SharedIPLibraryDAO.FindLatestIPLibraryWithType(nil, code)
	if err != nil {
		return err
	}
	if lib == nil {
		return nil
	}

	typeInfo := serverconfigs.FindIPLibraryWithType(code)
	if typeInfo == nil {
		return errors.New("invalid ip library code '" + code + "'")
	}

	path := Tea.Root + "/resources/ipdata/" + code + "/" + code + "." + fmt.Sprintf("%d", lib.CreatedAt) + typeInfo.GetString("ext")

	// 是否已经存在
	_, err = os.Stat(path)
	if err == nil {
		return nil
	}

	// 开始下载
	chunkIds, err := models.SharedFileChunkDAO.FindAllFileChunkIds(nil, int64(lib.FileId))
	if err != nil {
		return err
	}
	isOk := false

	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	defer func() {
		// 如果保存不成功就直接删除
		if !isOk {
			_ = fp.Close()
			_ = os.Remove(path)
		}
	}()
	for _, chunkId := range chunkIds {
		chunk, err := models.SharedFileChunkDAO.FindFileChunk(nil, chunkId)
		if err != nil {
			return err
		}
		if chunk == nil {
			continue
		}
		_, err = fp.Write([]byte(chunk.Data))
		if err != nil {
			return err
		}
	}

	err = fp.Close()
	if err != nil {
		return err
	}

	// 重新加载
	library, err := SharedManager.Load()
	if err != nil {
		return err
	}
	SharedLibrary = library

	isOk = true

	return nil
}
