// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestServerDAO_CopyServerConfigToServers(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	var dao = models.NewServerDAO()

	err := dao.CopyServerConfigToServers(tx, 10170, []int64{23, 10171}, serverconfigs.ConfigCodeStat)
	if err != nil {
		t.Fatal(err)
	}
}
