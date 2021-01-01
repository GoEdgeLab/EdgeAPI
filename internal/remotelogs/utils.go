package remotelogs

import (
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
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
			err := uploadLogs()
			if err != nil {
				logs.Println("[LOG]" + err.Error())
			}
		}
	}()
}

// 打印普通信息
func Println(tag string, description string) {
	logs.Println("[" + tag + "]" + description)

	nodeConfig, _ := configs.SharedAPIConfig()
	if nodeConfig == nil {
		return
	}

	select {
	case logChan <- &pb.NodeLog{
		Role:        teaconst.Role,
		Tag:         tag,
		Description: description,
		Level:       "info",
		NodeId:      nodeConfig.NumberId(),
		CreatedAt:   time.Now().Unix(),
	}:
	default:

	}
}

// 打印警告信息
func Warn(tag string, description string) {
	logs.Println("[" + tag + "]" + description)

	nodeConfig, _ := configs.SharedAPIConfig()
	if nodeConfig == nil {
		return
	}

	select {
	case logChan <- &pb.NodeLog{
		Role:        teaconst.Role,
		Tag:         tag,
		Description: description,
		Level:       "warning",
		NodeId:      nodeConfig.NumberId(),
		CreatedAt:   time.Now().Unix(),
	}:
	default:

	}
}

// 打印错误信息
func Error(tag string, description string) {
	logs.Println("[" + tag + "]" + description)

	nodeConfig, _ := configs.SharedAPIConfig()
	if nodeConfig == nil {
		return
	}

	select {
	case logChan <- &pb.NodeLog{
		Role:        teaconst.Role,
		Tag:         tag,
		Description: description,
		Level:       "error",
		NodeId:      nodeConfig.NumberId(),
		CreatedAt:   time.Now().Unix(),
	}:
	default:

	}
}

// 上传日志
func uploadLogs() error {
Loop:
	for {
		select {
		case log := <-logChan:
			err := models.SharedNodeLogDAO.CreateLog(nil, models.NodeRoleAPI, log.NodeId, log.Level, log.Tag, log.Description, log.CreatedAt)
			if err != nil {
				return err
			}
		default:
			break Loop
		}
	}

	return nil
}
