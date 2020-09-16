package tasks

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/logs"
	"time"
)

// TODO 考虑多个API服务同时运行的冲突
func init() {
	task := &ServerUpdateTask{}
	go task.Run()
}

// 更新服务配置
type ServerUpdateTask struct {
}

func (this *ServerUpdateTask) Run() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		this.loop()
	}
}

func (this *ServerUpdateTask) loop() {
	serverIds, err := models.SharedServerDAO.FindUpdatingServerIds()
	if err != nil {
		logs.Println("[ServerUpdateTask]" + err.Error())
		return
	}
	if len(serverIds) == 0 {
		return
	}
	for _, serverId := range serverIds {
		// 查找配置
		config, err := models.SharedServerDAO.ComposeServerConfig(serverId)
		if err != nil {
			logs.Println("[ServerUpdateTask]" + err.Error())
			continue
		}
		if config == nil {
			err = models.SharedServerDAO.UpdateServerIsUpdating(serverId, false)
			if err != nil {
				logs.Println("[ServerUpdateTask]" + err.Error())
				continue
			}
		}
		configData, err := json.Marshal(config)
		if err != nil {
			logs.Println("[ServerUpdateTask]" + err.Error())
			continue
		}

		// 修改配置
		err = models.SharedServerDAO.UpdateServerConfig(serverId, configData)
		if err != nil {
			logs.Println("[ServerUpdateTask]" + err.Error())
			continue
		}

		// 修改更新状态
		err = models.SharedServerDAO.UpdateServerIsUpdating(serverId, false)
		if err != nil {
			logs.Println("[ServerUpdateTask]" + err.Error())
			continue
		}
	}
}
