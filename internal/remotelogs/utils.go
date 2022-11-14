package remotelogs

import (
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/cespare/xxhash"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"time"
)

var logChan = make(chan *pb.NodeLog, 1024)
var sharedDAO DAOInterface

func init() {
	// 定期上传日志
	ticker := time.NewTicker(60 * time.Second)
	goman.New(func() {
		for range ticker.C {
			err := uploadLogs()
			if err != nil {
				logs.Println("[LOG]" + err.Error())
			}
		}
	})
}

// Println 打印普通信息
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

// Warn 打印警告信息
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

// Error 打印错误信息
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

// SetDAO 设置存储接口
func SetDAO(dao DAOInterface) {
	sharedDAO = dao
}

// 上传日志
func uploadLogs() error {
	if sharedDAO == nil {
		return nil
	}

	const hashSize = 5
	var hashList = []uint64{}

Loop:
	for {
		select {
		case log := <-logChan:
			// 是否已存在
			var hash = xxhash.Sum64String(types.String(log.NodeId) + "_" + log.Description)
			var found = false
			for _, h := range hashList {
				if h == hash {
					found = true
					break
				}
			}

			// 加入
			if !found {
				hashList = append(hashList, hash)
				if len(hashList) > hashSize {
					hashList = hashList[1:]
				}
				err := sharedDAO.CreateLog(nil, nodeconfigs.NodeRoleAPI, log.NodeId, log.ServerId, log.OriginId, log.Level, log.Tag, log.Description, log.CreatedAt, "", nil)
				if err != nil {
					return err
				}
			}
		default:
			break Loop
		}
	}

	return nil
}
