package tasks

import (
	"bytes"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewHealthCheckTask().Run()
		})
	})
}

// HealthCheckTask 节点健康检查任务
type HealthCheckTask struct {
	tasksMap map[int64]*HealthCheckClusterTask // taskId => task
}

func NewHealthCheckTask() *HealthCheckTask {
	return &HealthCheckTask{
		tasksMap: map[int64]*HealthCheckClusterTask{},
	}
}

func (this *HealthCheckTask) Run() {
	err := this.loop()
	if err != nil {
		logs.Println("[TASK][HEALTH_CHECK]" + err.Error())
	}

	ticker := utils.NewTicker(60 * time.Second)
	for ticker.Wait() {
		err := this.loop()
		if err != nil {
			logs.Println("[TASK][HEALTH_CHECK]" + err.Error())
		}
	}
}

func (this *HealthCheckTask) loop() error {
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
		clusterId := int64(cluster.Id)

		config := &serverconfigs.HealthCheckConfig{}
		if len(cluster.HealthCheck) > 0 {
			err = json.Unmarshal(cluster.HealthCheck, config)
			if err != nil {
				logs.Println("[TASK][HEALTH_CHECK]" + err.Error())
				continue
			}
		}

		task, ok := this.tasksMap[clusterId]
		if ok {
			// 检查是否有变化
			newJSON, _ := json.Marshal(config)
			oldJSON, _ := json.Marshal(task.Config())
			if bytes.Compare(oldJSON, newJSON) != 0 {
				logs.Println("[TASK][HEALTH_CHECK]update cluster '" + numberutils.FormatInt64(clusterId) + "'")
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
