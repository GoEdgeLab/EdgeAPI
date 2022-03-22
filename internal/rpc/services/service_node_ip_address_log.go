// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NodeIPAddressLogService IP地址相关日志
type NodeIPAddressLogService struct {
	BaseService
}

// CountAllNodeIPAddressLogs 计算日志数量
func (this *NodeIPAddressLogService) CountAllNodeIPAddressLogs(ctx context.Context, req *pb.CountAllNodeIPAddressLogsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNodeIPAddressLogDAO.CountLogs(tx, req.NodeIPAddressId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListNodeIPAddressLogs 列出单页日志
func (this *NodeIPAddressLogService) ListNodeIPAddressLogs(ctx context.Context, req *pb.ListNodeIPAddressLogsRequest) (*pb.ListNodeIPAddressLogsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	logs, err := models.SharedNodeIPAddressLogDAO.ListLogs(tx, req.NodeIPAddressId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbLogs = []*pb.NodeIPAddressLog{}
	for _, log := range logs {
		var pbAddr *pb.NodeIPAddress
		addr, err := models.SharedNodeIPAddressDAO.FindEnabledAddress(tx, int64(log.AddressId))
		if err != nil {
			return nil, err
		}
		if addr != nil {
			pbAddr = &pb.NodeIPAddress{
				Id:          int64(addr.Id),
				NodeId:      int64(addr.NodeId),
				Name:        addr.Name,
				Ip:          addr.Ip,
				Description: addr.Description,
				Role:        addr.Role,
			}
		}

		var pbAdmin *pb.Admin
		if log.AdminId > 0 {
			admin, err := models.SharedAdminDAO.FindEnabledAdmin(tx, int64(log.AdminId))
			if err != nil {
				return nil, err
			}
			if admin != nil {
				pbAdmin = &pb.Admin{
					Id:       int64(admin.Id),
					Fullname: admin.Fullname,
					Username: admin.Username,
				}
			}
		}

		pbLogs = append(pbLogs, &pb.NodeIPAddressLog{
			Id:            int64(log.Id),
			Description:   log.Description,
			CreatedAt:     int64(log.CreatedAt),
			IsOn:          log.IsOn,
			IsUp:          log.IsUp,
			CanAccess:     log.CanAccess,
			BackupIP:      log.BackupIP,
			NodeIPAddress: pbAddr,
			Admin:         pbAdmin,
		})
	}
	return &pb.ListNodeIPAddressLogsResponse{NodeIPAddressLogs: pbLogs}, nil
}
