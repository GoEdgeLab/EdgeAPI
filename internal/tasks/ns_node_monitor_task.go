package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReady(func() {
		task := NewNSNodeMonitorTask(60)
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for range ticker.C {
				err := task.loop()
				if err != nil {
					logs.Println("[TASK][NS_NODE_MONITOR]" + err.Error())
				}
			}
		}()
	})
}

// NSNodeMonitorTask 边缘节点监控任务
type NSNodeMonitorTask struct {
	intervalSeconds int
}

func NewNSNodeMonitorTask(intervalSeconds int) *NSNodeMonitorTask {
	return &NSNodeMonitorTask{
		intervalSeconds: intervalSeconds,
	}
}

func (this *NSNodeMonitorTask) Run() {

}

func (this *NSNodeMonitorTask) loop() error {
	// 检查上次运行时间，防止重复运行
	settingKey := systemconfigs.SettingCodeNSNodeMonitor + "Loop"
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

	clusters, err := nameservers.SharedNSClusterDAO.FindAllEnabledClusters(nil)
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

func (this *NSNodeMonitorTask) monitorCluster(cluster *nameservers.NSCluster) error {
	clusterId := int64(cluster.Id)

	// 检查离线节点
	inactiveNodes, err := nameservers.SharedNSNodeDAO.FindAllNotifyingInactiveNodesWithClusterId(nil, clusterId)
	if err != nil {
		return err
	}
	for _, node := range inactiveNodes {
		subject := "DNS节点\"" + node.Name + "\"已处于离线状态"
		msg := "DNS节点\"" + node.Name + "\"已处于离线状态"
		err = models.SharedMessageDAO.CreateNodeMessage(nil, nodeconfigs.NodeRoleDNS, clusterId, int64(node.Id), models.MessageTypeNSNodeInactive, models.LevelError, subject, msg, nil)
		if err != nil {
			return err
		}

		// 修改在线状态
		err = nameservers.SharedNSNodeDAO.UpdateNodeStatusIsNotified(nil, int64(node.Id))
		if err != nil {
			return err
		}
	}

	// TODO 检查恢复连接

	// 检查CPU、内存、磁盘不足节点，而且离线的节点不再重复提示
	// TODO 需要实现

	// TODO 检查53/tcp、53/udp是否能够访问

	return nil
}
