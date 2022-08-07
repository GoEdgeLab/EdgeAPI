package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
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

// NodeMonitorTask 边缘节点监控任务
type NodeMonitorTask struct {
	BaseTask

	ticker *time.Ticker

	inactiveMap map[string]int  // cluster@nodeId => count
	notifiedMap map[int64]int64 // nodeId => timestamp
}

func NewNodeMonitorTask(duration time.Duration) *NodeMonitorTask {
	return &NodeMonitorTask{
		ticker:      time.NewTicker(duration),
		inactiveMap: map[string]int{},
		notifiedMap: map[int64]int64{},
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
	if !models.SharedAPINodeDAO.CheckAPINodeIsPrimaryWithoutErr() {
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

	var nodeMap = map[int64]*models.Node{}
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
				var msg = "集群'" + cluster.Name + "'节点\"" + node.Name + "\"已处于离线状态，请检查节点是否异常"
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

	// 检查CPU、内存、磁盘不足节点
	// TODO 需要实现

	return nil
}
