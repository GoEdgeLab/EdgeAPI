package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
)

// 账单相关服务
type UserBillService struct {
	BaseService
}

// 手工生成订单
func (this *UserBillService) GenerateAllUserBills(ctx context.Context, req *pb.GenerateAllUserBillsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 校验Month
	if !regexp.MustCompile(`^\d{6}$`).MatchString(req.Month) {
		return nil, errors.New("invalid month '" + req.Month + "'")
	}
	if req.Month >= timeutil.Format("Ym") {
		return nil, errors.New("invalid month '" + req.Month + "'")
	}

	err = models.SharedUserBillDAO.GenerateBills(req.Month)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 计算所有账单数量
func (this *UserBillService) CountAllUserBills(ctx context.Context, req *pb.CountAllUserBillsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedUserBillDAO.CountAllUserBills(req.PaidFlag, req.UserId, req.Month)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 列出单页账单
func (this *UserBillService) ListUserBills(ctx context.Context, req *pb.ListUserBillsRequest) (*pb.ListUserBillsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	bills, err := models.SharedUserBillDAO.ListUserBills(req.PaidFlag, req.UserId, req.Month, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.UserBill{}
	for _, bill := range bills {
		userFullname, err := models.SharedUserDAO.FindUserFullname(int64(bill.UserId))
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.UserBill{
			Id: int64(bill.Id),
			User: &pb.User{
				Id:       int64(bill.UserId),
				Fullname: userFullname,
			},
			Type:        bill.Type,
			TypeName:    models.SharedUserBillDAO.BillTypeName(bill.Type),
			Description: bill.Description,
			Amount:      float32(bill.Amount),
			Month:       bill.Month,
			IsPaid:      bill.IsPaid == 1,
			PaidAt:      int64(bill.PaidAt),
		})
	}
	return &pb.ListUserBillsResponse{UserBills: result}, nil
}
