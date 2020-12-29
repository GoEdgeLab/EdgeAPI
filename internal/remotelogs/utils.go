package remotelogs

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/logs"
	"time"
)

var logChan = make(chan *pb.NodeLog, 1024)

func init() {
	// 定期上传日志
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			// TODO
		}
	}()
}

// 打印普通信息
func Println(tag string, description string) {
	logs.Println("[" + tag + "]" + description)

	nodeConfig, _ := nodeconfigs.SharedNodeConfig()
	if nodeConfig == nil {
		return
	}

	select {
	case logChan <- &pb.NodeLog{
		Role:        teaconst.Role,
		Tag:         tag,
		Description: description,
		Level:       "info",
		NodeId:      nodeConfig.Id,
		CreatedAt:   time.Now().Unix(),
	}:
	default:

	}
}

// 打印警告信息
func Warn(tag string, description string) {
	logs.Println("[" + tag + "]" + description)

	nodeConfig, _ := nodeconfigs.SharedNodeConfig()
	if nodeConfig == nil {
		return
	}

	select {
	case logChan <- &pb.NodeLog{
		Role:        teaconst.Role,
		Tag:         tag,
		Description: description,
		Level:       "warning",
		NodeId:      nodeConfig.Id,
		CreatedAt:   time.Now().Unix(),
	}:
	default:

	}
}

// 打印错误信息
func Error(tag string, description string) {
	logs.Println("[" + tag + "]" + description)

	nodeConfig, _ := nodeconfigs.SharedNodeConfig()
	if nodeConfig == nil {
		return
	}

	select {
	case logChan <- &pb.NodeLog{
		Role:        teaconst.Role,
		Tag:         tag,
		Description: description,
		Level:       "error",
		NodeId:      nodeConfig.Id,
		CreatedAt:   time.Now().Unix(),
	}:
	default:

	}
}
