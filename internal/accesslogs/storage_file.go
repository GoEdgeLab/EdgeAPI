package accesslogs

import (
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/logs"
	"os"
	"path/filepath"
	"sync"
)

// FileStorage 文件存储策略
type FileStorage struct {
	BaseStorage

	config *serverconfigs.AccessLogFileStorageConfig

	writeLocker sync.Mutex

	files       map[string]*os.File // path => *File
	filesLocker sync.Mutex
}

func NewFileStorage(config *serverconfigs.AccessLogFileStorageConfig) *FileStorage {
	return &FileStorage{
		config: config,
	}
}

func (this *FileStorage) Config() interface{} {
	return this.config
}

// Start 开启
func (this *FileStorage) Start() error {
	if len(this.config.Path) == 0 {
		return errors.New("'path' should not be empty")
	}

	this.files = map[string]*os.File{}

	return nil
}

// Write 写入日志
func (this *FileStorage) Write(accessLogs []*pb.HTTPAccessLog) error {
	if len(accessLogs) == 0 {
		return nil
	}

	fp := this.fp()
	if fp == nil {
		return errors.New("file pointer should not be nil")
	}
	this.writeLocker.Lock()
	defer this.writeLocker.Unlock()

	for _, accessLog := range accessLogs {
		data, err := this.Marshal(accessLog)
		if err != nil {
			logs.Error(err)
			continue
		}
		_, err = fp.Write(data)
		if err != nil {
			_ = this.Close()
			break
		}
		_, _ = fp.WriteString("\n")
	}
	return nil
}

// Close 关闭
func (this *FileStorage) Close() error {
	this.filesLocker.Lock()
	defer this.filesLocker.Unlock()

	var resultErr error
	for _, f := range this.files {
		err := f.Close()
		if err != nil {
			resultErr = err
		}
	}
	return resultErr
}

func (this *FileStorage) fp() *os.File {
	path := this.FormatVariables(this.config.Path)

	this.filesLocker.Lock()
	defer this.filesLocker.Unlock()
	fp, ok := this.files[path]
	if ok {
		return fp
	}

	// 关闭其他的文件
	for _, f := range this.files {
		_ = f.Close()
	}

	// 是否创建文件目录
	if this.config.AutoCreate {
		dir := filepath.Dir(path)
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0777)
			if err != nil {
				logs.Error(err)
				return nil
			}
		}
	}

	// 打开新文件
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logs.Error(err)
		return nil
	}
	this.files[path] = fp

	return fp
}
