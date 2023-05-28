package tasks

import (
	"bytes"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewHealthCheckTask(1 * time.Minute).Start()
		})
	})
}

// HealthCheckTask 节点健康检查任务
type HealthCheckTask struct {
	BaseTask

	ticker   *time.Ticker
	tasksMap map[int64]*HealthCheckClusterTask // taskId => task
}

func NewHealthCheckTask(duration time.Duration) *HealthCheckTask {
	return &HealthCheckTask{
		ticker:   time.NewTicker(duration),
		tasksMap: map[int64]*HealthCheckClusterTask{},
	}
}

func (this *HealthCheckTask) Start() {
	err := this.Loop()
	if err != nil {
		this.logErr("HealthCheckTask", err.Error())
	}

	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("HealthCheckTask", err.Error())
		}
	}
}

func (this *HealthCheckTask) Loop() error {
	clusters, err := models.NewNodeClusterDAO().FindAllEnableClusters(nil)
	if err != nil {
		return err
	}
	clusterIds := []int64{}
	for _, cluster := range clusters {
		clusterIds = append(clusterIds, int64(cluster.Id))
	}

	// 停掉删除的
	for clusterId, task := range this.tasksMap {
		if !lists.ContainsInt64(clusterIds, clusterId) {
			task.Stop()
			delete(this.tasksMap, clusterId)
		}
	}

	// 启动新的或更新老的
	for _, cluster := range clusters {
		var clusterId = int64(cluster.Id)

		if !cluster.IsOn {
			this.stopClusterTask(clusterId)
			continue
		}

		// 检查当前集群上是否有服务，如果尚没有部署服务，则直接跳过
		countServers, err := models.SharedServerDAO.CountAllEnabledServersWithNodeClusterId(nil, clusterId)
		if err != nil {
			return err
		}
		if countServers == 0 {
			this.stopClusterTask(clusterId)
			continue
		}

		var config = &serverconfigs.HealthCheckConfig{}
		if len(cluster.HealthCheck) > 0 {
			err = json.Unmarshal(cluster.HealthCheck, config)
			if err != nil {
				this.logErr("HealthCheckTask", err.Error())
				this.stopClusterTask(clusterId)
				continue
			}
			if !config.IsOn {
				this.stopClusterTask(clusterId)
				continue
			}
		} else {
			this.stopClusterTask(clusterId)
			continue
		}

		task, ok := this.tasksMap[clusterId]
		if ok {
			// 检查是否有变化
			newJSON, _ := json.Marshal(config)
			oldJSON, _ := json.Marshal(task.Config())
			if !bytes.Equal(oldJSON, newJSON) {
				remotelogs.Println("TASK", "[HealthCheckTask]update cluster '"+numberutils.FormatInt64(clusterId)+"'")
				goman.New(func() {
					task.Reset(config)
				})
			}
		} else {
			task = NewHealthCheckClusterTask(clusterId, config)
			this.tasksMap[clusterId] = task
			goman.New(func() {
				task.Run()
			})
		}
	}

	return nil
}

func (this *HealthCheckTask) stopClusterTask(clusterId int64) {
	var task = this.tasksMap[clusterId]
	if task != nil {
		task.Stop()
		delete(this.tasksMap, clusterId)
	}
}
