package accesslogs

import (
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/logs"
	"net"
	"sync"
)

// TCPStorage TCP存储策略
type TCPStorage struct {
	BaseStorage

	config *serverconfigs.AccessLogTCPStorageConfig

	writeLocker sync.Mutex

	connLocker sync.Mutex
	conn       net.Conn
}

func NewTCPStorage(config *serverconfigs.AccessLogTCPStorageConfig) *TCPStorage {
	return &TCPStorage{config: config}
}

func (this *TCPStorage) Config() interface{} {
	return this.config
}

// Start 开启
func (this *TCPStorage) Start() error {
	if len(this.config.Network) == 0 {
		return errors.New("'network' should not be empty")
	}
	if len(this.config.Addr) == 0 {
		return errors.New("'addr' should not be empty")
	}
	return nil
}

// 写入日志
func (this *TCPStorage) Write(accessLogs []*pb.HTTPAccessLog) error {
	if len(accessLogs) == 0 {
		return nil
	}

	err := this.connect()
	if err != nil {
		return err
	}

	conn := this.conn
	if conn == nil {
		return errors.New("connection should not be nil")
	}

	this.writeLocker.Lock()
	defer this.writeLocker.Unlock()

	for _, accessLog := range accessLogs {
		data, err := this.Marshal(accessLog)
		if err != nil {
			logs.Error(err)
			continue
		}
		_, err = conn.Write(data)
		if err != nil {
			_ = this.Close()
			break
		}
		_, err = conn.Write([]byte("\n"))
		if err != nil {
			_ = this.Close()
			break
		}
	}

	return nil
}

// Close 关闭
func (this *TCPStorage) Close() error {
	this.connLocker.Lock()
	defer this.connLocker.Unlock()

	if this.conn != nil {
		err := this.conn.Close()
		this.conn = nil
		return err
	}
	return nil
}

func (this *TCPStorage) connect() error {
	this.connLocker.Lock()
	defer this.connLocker.Unlock()

	if this.conn != nil {
		return nil
	}

	conn, err := net.Dial(this.config.Network, this.config.Addr)
	if err != nil {
		return err
	}
	this.conn = conn

	return nil
}
