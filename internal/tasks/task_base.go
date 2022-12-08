// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
)

type BaseTask struct {
}

func (this *BaseTask) logErr(taskType string, errString string) {
	remotelogs.Error("TASK", "run '"+taskType+"' failed: "+errString)
}

func (this *BaseTask) IsPrimaryNode() bool {
	return models.SharedAPINodeDAO.CheckAPINodeIsPrimaryWithoutErr()
}
