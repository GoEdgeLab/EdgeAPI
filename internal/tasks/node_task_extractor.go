package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		go NewNodeTaskExtractor().Start()
	})
}

// 节点任务
type NodeTaskExtractor struct {
}

func NewNodeTaskExtractor() *NodeTaskExtractor {
	return &NodeTaskExtractor{}
}

func (this *NodeTaskExtractor) Start() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		err := this.Loop()
		if err != nil {
			logs.Println("[TASK][NODE_TASK_EXTRACTOR]" + err.Error())
		}
	}
}

func (this *NodeTaskExtractor) Loop() error {
	ok, err := models.SharedSysLockerDAO.Lock(nil, "node_task_extractor", 10-1) // 假设执行时间为1秒
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// 这里不解锁，是为了让任务N秒钟之内只运行一次

	for _, role := range []string{nodeconfigs.NodeRoleNode, nodeconfigs.NodeRoleDNS} {
		err = models.SharedNodeTaskDAO.ExtractAllClusterTasks(nil, role)
		if err != nil {
			return err
		}
	}

	return nil
}
