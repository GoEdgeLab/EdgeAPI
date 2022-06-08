package accesslogs

import (
	"bytes"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/logs"
	"os/exec"
	"runtime"
	"strconv"
)

type SyslogStorageProtocol = string

const (
	SyslogStorageProtocolTCP    SyslogStorageProtocol = "tcp"
	SyslogStorageProtocolUDP    SyslogStorageProtocol = "udp"
	SyslogStorageProtocolNone   SyslogStorageProtocol = "none"
	SyslogStorageProtocolSocket SyslogStorageProtocol = "socket"
)

type SyslogStoragePriority = int

// SyslogStorage syslog存储策略
type SyslogStorage struct {
	BaseStorage

	config *serverconfigs.AccessLogSyslogStorageConfig

	exe string
}

func NewSyslogStorage(config *serverconfigs.AccessLogSyslogStorageConfig) *SyslogStorage {
	return &SyslogStorage{config: config}
}

func (this *SyslogStorage) Config() interface{} {
	return this.config
}

// Start 开启
func (this *SyslogStorage) Start() error {
	if runtime.GOOS != "linux" {
		return errors.New("'syslog' storage only works on linux")
	}

	exe, err := exec.LookPath("logger")
	if err != nil {
		return err
	}

	this.exe = exe

	return nil
}

// 写入日志
func (this *SyslogStorage) Write(accessLogs []*pb.HTTPAccessLog) error {
	if len(accessLogs) == 0 {
		return nil
	}

	args := []string{}
	if len(this.config.Tag) > 0 {
		args = append(args, "-t", this.config.Tag)
	}

	if this.config.Priority >= 0 {
		args = append(args, "-p", strconv.Itoa(this.config.Priority))
	}

	switch this.config.Protocol {
	case SyslogStorageProtocolTCP:
		args = append(args, "-T")
		if len(this.config.ServerAddr) > 0 {
			args = append(args, "-n", this.config.ServerAddr)
		}
		if this.config.ServerPort > 0 {
			args = append(args, "-P", strconv.Itoa(this.config.ServerPort))
		}
	case SyslogStorageProtocolUDP:
		args = append(args, "-d")
		if len(this.config.ServerAddr) > 0 {
			args = append(args, "-n", this.config.ServerAddr)
		}
		if this.config.ServerPort > 0 {
			args = append(args, "-P", strconv.Itoa(this.config.ServerPort))
		}
	case SyslogStorageProtocolSocket:
		args = append(args, "-u")
		args = append(args, this.config.Socket)
	case SyslogStorageProtocolNone:
		// do nothing
	}

	args = append(args, "-S", "10240")

	var cmd = exec.Command(this.exe, args...)
	var stderrBuffer = &bytes.Buffer{}
	cmd.Stderr = stderrBuffer

	w, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	for _, accessLog := range accessLogs {
		if this.firewallOnly && accessLog.FirewallPolicyId == 0 {
			continue
		}
		data, err := this.Marshal(accessLog)
		if err != nil {
			remotelogs.Error("ACCESS_LOG_POLICY_SYSLOG", "marshal accesslog failed: "+err.Error())
			continue
		}
		_, err = w.Write(data)
		if err != nil {
			logs.Error(err)
		}

		_, err = w.Write([]byte("\n"))
		if err != nil {
			remotelogs.Error("ACCESS_LOG_POLICY_SYSLOG", "write accesslog failed: "+err.Error())
		}
	}

	_ = w.Close()

	err = cmd.Wait()
	if err != nil {
		return errors.New("send syslog failed: " + err.Error() + ", stderr: " + stderrBuffer.String())
	}

	return nil
}

// Close 关闭
func (this *SyslogStorage) Close() error {
	return nil
}
