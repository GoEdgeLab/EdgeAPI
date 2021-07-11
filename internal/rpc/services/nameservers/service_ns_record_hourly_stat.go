// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

// NSRecordHourlyStatService NS记录小时统计
type NSRecordHourlyStatService struct {
	services.BaseService
}

// UploadNSRecordHourlyStats 上传统计
func (this *NSRecordHourlyStatService) UploadNSRecordHourlyStats(ctx context.Context, req *pb.UploadNSRecordHourlyStatsRequest) (*pb.RPCSuccess, error) {
	_, nodeId, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}
	if nodeId <= 0 {
		return nil, errors.New("invalid nodeId")
	}
	if len(req.Stats) == 0 {
		return this.Success()
	}

	var tx = this.NullTx()
	clusterId, err := nameservers.SharedNSNodeDAO.FindNodeClusterId(tx, nodeId)
	if err != nil {
		return nil, err
	}

	// 增加小时统计
	for _, stat := range req.Stats {
		err := nameservers.SharedNSRecordHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, nodeId, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.NsDomainId, stat.NsRecordId, stat.CountRequests, stat.Bytes)
		if err != nil {
			return nil, err
		}
	}


	return this.Success()
}
