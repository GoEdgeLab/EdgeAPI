package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReady(func() {
		task := NewNodeMonitorTask(60)
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for range ticker.C {
				err := task.loop()
				if err != nil {
					logs.Println("[TASK][NODE_MONITOR]" + err.Error())
				}
			}
		}()
	})
}

// NodeMonitorTask 边缘节点监控任务
type NodeMonitorTask struct {
	intervalSeconds int
}

func NewNodeMonitorTask(intervalSeconds int) *NodeMonitorTask {
	return &NodeMonitorTask{
		intervalSeconds: intervalSeconds,
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
	for _, node := range inactiveNodes {
		subject := "节点\"" + node.Name + "\"已处于离线状态"
		msg := "节点\"" + node.Name + "\"已处于离线状态"
		err = models.SharedMessageDAO.CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, clusterId, int64(node.Id), models.MessageTypeNodeInactive, models.LevelError, subject, msg, nil, false)
		if err != nil {
			return err
		}

		// 修改在线状态
		err = models.SharedNodeDAO.UpdateNodeActive(nil, int64(node.Id), false)
		if err != nil {
			return err
		}
	}

	// TODO 检查恢复连接

	// 检查CPU、内存、磁盘不足节点，而且离线的节点不再重复提示
	// TODO 需要实现

	return nil
}
