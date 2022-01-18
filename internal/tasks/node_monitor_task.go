package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"strings"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		var task = NewNodeMonitorTask(60)
		var ticker = time.NewTicker(60 * time.Second)
		goman.New(func() {
			for range ticker.C {
				err := task.loop()
				if err != nil {
					logs.Println("[TASK][NODE_MONITOR]" + err.Error())
				}
			}
		})
	})
}

// NodeMonitorTask 边缘节点监控任务
type NodeMonitorTask struct {
	intervalSeconds int

	inactiveMap map[string]int  // cluster@nodeId => count
	notifiedMap map[int64]int64 // nodeId => timestamp
}

func NewNodeMonitorTask(intervalSeconds int) *NodeMonitorTask {
	return &NodeMonitorTask{
		intervalSeconds: intervalSeconds,
		inactiveMap:     map[string]int{},
		notifiedMap:     map[int64]int64{},
	}
}

func (this *NodeMonitorTask) Run() {

}

func (this *NodeMonitorTask) loop() error {
	// 检查上次运行时间，防止重复运行
	settingKey := systemconfigs.SettingCodeNodeMonitor + "Loop"
	timestamp := time.Now().Unix()
	c, err := models.SharedSysSettingDAO.CompareInt64Setting(nil, settingKey, timestamp-int64(this.intervalSeconds))
	if err != nil {
		return err
	}
	if c > 0 {
		return nil
	}

	// 记录时间
	err = models.SharedSysSettingDAO.UpdateSetting(nil, settingKey, []byte(numberutils.FormatInt64(timestamp)))
	if err != nil {
		return err
	}

	clusters, err := models.SharedNodeClusterDAO.FindAllEnableClusters(nil)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		err := this.monitorCluster(cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *NodeMonitorTask) monitorCluster(cluster *models.NodeCluster) error {
	clusterId := int64(cluster.Id)

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

				subject := "节点\"" + node.Name + "\"已处于离线状态"
				msg := "集群'" + cluster.Name + "'节点\"" + node.Name + "\"已处于离线状态，请检查节点是否异常"
				err = models.SharedMessageDAO.CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, clusterId, int64(node.Id), models.MessageTypeNodeInactive, models.LevelError, subject, msg, nil, false)
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
