package accesslogs

import (
	"bytes"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/logs"
	"os/exec"
	"sync"
)

// CommandStorage 通过命令行存储
type CommandStorage struct {
	BaseStorage

	config *serverconfigs.AccessLogCommandStorageConfig

	writeLocker sync.Mutex
}

func NewCommandStorage(config *serverconfigs.AccessLogCommandStorageConfig) *CommandStorage {
	return &CommandStorage{config: config}
}

func (this *CommandStorage) Config() interface{} {
	return this.config
}

// Start 启动
func (this *CommandStorage) Start() error {
	if len(this.config.Command) == 0 {
		return errors.New("'command' should not be empty")
	}
	return nil
}

// 写入日志
func (this *CommandStorage) Write(accessLogs []*pb.HTTPAccessLog) error {
	if len(accessLogs) == 0 {
		return nil
	}

	this.writeLocker.Lock()
	defer this.writeLocker.Unlock()

	cmd := exec.Command(this.config.Command, this.config.Args...)
	if len(this.config.Dir) > 0 {
		cmd.Dir = this.config.Dir
	}

	stdout := bytes.NewBuffer([]byte{})
	cmd.Stdout = stdout

	w, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	for _, accessLog := range accessLogs {
		data, err := this.Marshal(accessLog)
		if err != nil {
			logs.Error(err)
			continue
		}
		_, err = w.Write(data)
		if err != nil {
			logs.Error(err)
		}

		_, err = w.Write([]byte("\n"))
		if err != nil {
			logs.Error(err)
		}
	}
	_ = w.Close()
	err = cmd.Wait()
	if err != nil {
		logs.Error(err)

		if stdout.Len() > 0 {
			logs.Error(errors.New(string(stdout.Bytes())))
		}
	}

	return nil
}

// Close 关闭
func (this *CommandStorage) Close() error {
	return nil
}
