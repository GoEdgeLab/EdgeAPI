package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewNSNodeMonitorTask(1 * time.Minute).Start()
		})
	})
}

// NSNodeMonitorTask 边缘节点监控任务
type NSNodeMonitorTask struct {
	BaseTask

	ticker *time.Ticker
}

func NewNSNodeMonitorTask(duration time.Duration) *NSNodeMonitorTask {
	return &NSNodeMonitorTask{
		ticker: time.NewTicker(duration),
	}
}

func (this *NSNodeMonitorTask) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("NS_NODE_MONITOR", err.Error())
		}
	}
}

func (this *NSNodeMonitorTask) Loop() error {
	// 检查是否为主节点
	if !models.SharedAPINodeDAO.CheckAPINodeIsPrimaryWithoutErr() {
		return nil
	}

	clusters, err := models.SharedNSClusterDAO.FindAllEnabledClusters(nil)
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

func (this *NSNodeMonitorTask) monitorCluster(cluster *models.NSCluster) error {
	clusterId := int64(cluster.Id)

	// 检查离线节点
	inactiveNodes, err := models.SharedNSNodeDAO.FindAllNotifyingInactiveNodesWithClusterId(nil, clusterId)
	if err != nil {
		return err
	}
	for _, node := range inactiveNodes {
		subject := "DNS节点\"" + node.Name + "\"已处于离线状态"
		msg := "DNS节点\"" + node.Name + "\"已处于离线状态"
		err = models.SharedMessageDAO.CreateNodeMessage(nil, nodeconfigs.NodeRoleDNS, clusterId, int64(node.Id), models.MessageTypeNSNodeInactive, models.LevelError, subject, msg, nil, false)
		if err != nil {
			return err
		}

		// 修改在线状态
		err = models.SharedNSNodeDAO.UpdateNodeStatusIsNotified(nil, int64(node.Id))
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
