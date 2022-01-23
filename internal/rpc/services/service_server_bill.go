// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ServerBillService 服务账单相关服务
type ServerBillService struct {
	BaseService
}

// CountAllServerBills 查询服务账单数量
func (this *ServerBillService) CountAllServerBills(ctx context.Context, req *pb.CountAllServerBillsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := models.SharedServerBillDAO.CountServerBills(tx, req.UserId, req.Month)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListServerBills 查询服务账单列表
func (this *ServerBillService) ListServerBills(ctx context.Context, req *pb.ListServerBillsRequest) (*pb.ListServerBillsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	serverBills, err := models.SharedServerBillDAO.ListServerBills(tx, req.UserId, req.Month, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbServerBills = []*pb.ServerBill{}
	var cacheMap = utils.NewCacheMap()
	for _, bill := range serverBills {
		// user
		user, err := models.SharedUserDAO.FindBasicUserWithoutState(tx, int64(bill.UserId))
		if err != nil {
			return nil, err
		}
		var pbUser = &pb.User{Id: int64(bill.UserId)}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(bill.UserId),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}

		// plan
		var pbPlan *pb.Plan
		if bill.PlanId > 0 {
			plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(bill.PlanId))
			if err != nil {
				return nil, err
			}
			if plan != nil {
				pbPlan = &pb.Plan{
					Id:        int64(plan.Id),
					Name:      plan.Name,
					PriceType: plan.PriceType,
				}
			}
		}

		// user plan
		var pbUserPlan *pb.UserPlan
		if bill.UserPlanId > 0 {
			userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, int64(bill.UserPlanId), cacheMap)
			if err != nil {
				return nil, err
			}
			if userPlan != nil {
				pbUserPlan = &pb.UserPlan{
					Id: int64(userPlan.Id),
				}
			}
		}

		// server
		var pbServer *pb.Server
		if bill.ServerId > 0 {
			server, err := models.SharedServerDAO.FindEnabledServerBasic(tx, int64(bill.ServerId))
			if err != nil {
				return nil, err
			}
			if server != nil {
				pbServer = &pb.Server{Id: int64(bill.ServerId), Name: server.Name}
			}
		}

		pbServerBills = append(pbServerBills, &pb.ServerBill{
			Id:                       int64(bill.Id),
			UserId:                   int64(bill.UserId),
			ServerId:                 int64(bill.ServerId),
			Amount:                   float32(bill.Amount),
			CreatedAt:                int64(bill.CreatedAt),
			UserPlanId:               int64(bill.UserPlanId),
			PlanId:                   int64(bill.PlanId),
			TotalTrafficBytes:        int64(bill.TotalTrafficBytes),
			BandwidthPercentileBytes: int64(bill.BandwidthPercentileBytes),
			BandwidthPercentile:      int32(bill.BandwidthPercentile),
			User:                     pbUser,
			Plan:                     pbPlan,
			UserPlan:                 pbUserPlan,
			Server:                   pbServer,
		})
	}

	return &pb.ListServerBillsResponse{ServerBills: pbServerBills}, nil
}
