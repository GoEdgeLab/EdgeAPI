package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strings"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewNodeMonitorTask(1 * time.Minute).Start()
		})
	})
}

// 节点启动尝试
type nodeStartingTry struct {
	count     int
	timestamp int64
}

// NodeMonitorTask 边缘节点监控任务
type NodeMonitorTask struct {
	BaseTask

	ticker *time.Ticker

	inactiveMap map[string]int  // cluster@nodeId => count
	notifiedMap map[int64]int64 // nodeId => timestamp

	recoverMap map[int64]*nodeStartingTry // nodeId => *nodeStartingTry
}

func NewNodeMonitorTask(duration time.Duration) *NodeMonitorTask {
	return &NodeMonitorTask{
		ticker:      time.NewTicker(duration),
		inactiveMap: map[string]int{},
		notifiedMap: map[int64]int64{},
		recoverMap:  map[int64]*nodeStartingTry{},
	}
}

func (this *NodeMonitorTask) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("NodeMonitorTask", err.Error())
		}
	}
}

func (this *NodeMonitorTask) Loop() error {
	// 检查是否为主节点
	if !this.IsPrimaryNode() {
		return nil
	}

	clusters, err := models.SharedNodeClusterDAO.FindAllEnableClusters(nil)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		err := this.MonitorCluster(cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *NodeMonitorTask) MonitorCluster(cluster *models.NodeCluster) error {
	var clusterId = int64(cluster.Id)

	// 检查离线节点
	inactiveNodes, err := models.SharedNodeDAO.FindAllInactiveNodesWithClusterId(nil, clusterId)
	if err != nil {
		return err
	}

	// 尝试自动远程启动
	if cluster.AutoRemoteStart {
		var nodeQueue = installers.NewNodeQueue()
		for _, node := range inactiveNodes {
			var nodeId = int64(node.Id)
			tryInfo, ok := this.recoverMap[nodeId]
			if !ok {
				tryInfo = &nodeStartingTry{
					count:     1,
					timestamp: time.Now().Unix(),
				}
				this.recoverMap[nodeId] = tryInfo
			} else {
				if tryInfo.count >= 3 /** 3次 **/ { // N 秒内超过 M 次就暂时不再重新尝试，防止阻塞当前任务
					if tryInfo.timestamp+10*60 /** 10 分钟 **/ > time.Now().Unix() {
						continue
					}
					tryInfo.timestamp = time.Now().Unix()
					tryInfo.count = 0
				}
				tryInfo.count++
			}

			// TODO 如果用户手工安装的位置不在标准位置，需要节点自身记住最近启动的位置
			err = nodeQueue.StartNode(nodeId)
			if err != nil {
				if !installers.IsGrantError(err) {
					_ = models.SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleNode, nodeId, 0, 0, models.LevelInfo, "NODE", "start node from remote API failed: "+err.Error(), time.Now().Unix(), "", nil)
				}
			} else {
				_ = models.SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleNode, nodeId, 0, 0, models.LevelSuccess, "NODE", "start node from remote API successfully", time.Now().Unix(), "", nil)
			}
		}
	}

	var nodeMap = map[int64]*models.Node{} // nodeId => Node
	for _, node := range inactiveNodes {
		var nodeId = int64(node.Id)
		nodeMap[nodeId] = node
		this.inactiveMap[types.String(clusterId)+"@"+types.String(nodeId)]++
	}

	const maxInactiveTries = 5

	// 处理现有的离线状态
	for key, count := range this.inactiveMap {
		var pieces = strings.Split(key, "@")
		if pieces[0] != types.String(clusterId) {
			continue
		}
		var nodeId = types.Int64(pieces[1])
		node, ok := nodeMap[nodeId]
		if ok {
			// 连续 N 次离线发送通知
			// 同时也要确保两次发送通知的时间不会过近
			if count >= maxInactiveTries && time.Now().Unix()-this.notifiedMap[nodeId] > 3600 {
				this.inactiveMap[key] = 0
				this.notifiedMap[nodeId] = time.Now().Unix()

				var subject = "节点\"" + node.Name + "\"已处于离线状态"
				var msg = "集群 \"" + cluster.Name + "\" 节点 \"" + node.Name + "\" 已处于离线状态，请检查节点是否异常"
				err = models.SharedMessageDAO.CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, clusterId, int64(node.Id), models.MessageTypeNodeInactive, models.LevelError, subject, msg, nil, false)
				if err != nil {
					return err
				}

				// 设置通知时间
				err = models.SharedNodeDAO.UpdateNodeInactiveNotifiedAt(nil, nodeId, time.Now().Unix())
				if err != nil {
					return err
				}
			}
		} else {
			delete(this.inactiveMap, key)
		}
	}

	// 检查CPU、内存、硬盘不足节点
	// TODO 需要实现

	return nil
}
