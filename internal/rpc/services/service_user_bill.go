package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/accounts"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
)

// UserBillService 账单相关服务
type UserBillService struct {
	BaseService
}

// GenerateAllUserBills 手工生成订单
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

	tx := this.NullTx()

	err = models.SharedUserBillDAO.GenerateBills(tx, req.Month)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllUserBills 计算所有账单数量
func (this *UserBillService) CountAllUserBills(ctx context.Context, req *pb.CountAllUserBillsRequest) (*pb.RPCCountResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedUserBillDAO.CountAllUserBills(tx, req.PaidFlag, req.UserId, req.Month)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserBills 列出单页账单
func (this *UserBillService) ListUserBills(ctx context.Context, req *pb.ListUserBillsRequest) (*pb.ListUserBillsResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	bills, err := models.SharedUserBillDAO.ListUserBills(tx, req.PaidFlag, req.UserId, req.Month, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.UserBill{}
	for _, bill := range bills {
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(bill.UserId))
		if err != nil {
			return nil, err
		}
		if user == nil {
			user = &models.User{Id: bill.UserId}
		}

		result = append(result, &pb.UserBill{
			Id: int64(bill.Id),
			User: &pb.User{
				Id:       int64(bill.UserId),
				Fullname: user.Fullname,
				Username: user.Username,
			},
			Type:        bill.Type,
			TypeName:    models.SharedUserBillDAO.BillTypeName(bill.Type),
			Description: bill.Description,
			Amount:      float32(bill.Amount),
			Month:       bill.Month,
			IsPaid:      bill.IsPaid == 1,
			PaidAt:      int64(bill.PaidAt),
			Code:        bill.Code,
		})
	}
	return &pb.ListUserBillsResponse{UserBills: result}, nil
}

// FindUserBill 查找账单信息
func (this *UserBillService) FindUserBill(ctx context.Context, req *pb.FindUserBillRequest) (*pb.FindUserBillResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedUserBillDAO.CheckUserBill(tx, userId, req.UserBillId)
		if err != nil {
			return nil, err
		}
	}

	bill, err := models.SharedUserBillDAO.FindUserBill(tx, req.UserBillId)
	if err != nil {
		return nil, err
	}

	if bill == nil {
		return &pb.FindUserBillResponse{UserBill: nil}, nil
	}

	// 用户
	var pbUser = &pb.User{Id: int64(bill.UserId)}
	user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(bill.UserId), nil)
	if err != nil {
		return nil, err
	}
	if user != nil {
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	return &pb.FindUserBillResponse{
		UserBill: &pb.UserBill{
			Id:          int64(bill.Id),
			User:        pbUser,
			Type:        bill.Type,
			TypeName:    models.SharedUserBillDAO.BillTypeName(bill.Type),
			Description: bill.Description,
			Amount:      float32(bill.Amount),
			Month:       bill.Month,
			IsPaid:      bill.IsPaid == 1,
			PaidAt:      int64(bill.PaidAt),
			Code:        bill.Code,
		},
	}, nil
}

// PayUserBill 支付账单
func (this *UserBillService) PayUserBill(ctx context.Context, req *pb.PayUserBillRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		// 检查用户
		if userId > 0 {
			err = models.SharedUserBillDAO.CheckUserBill(tx, userId, req.UserBillId)
			if err != nil {
				return err
			}
		}

		// 是否存在
		bill, err := models.SharedUserBillDAO.FindUserBill(tx, req.UserBillId)
		if err != nil {
			return err
		}
		if bill == nil {
			return nil
		}
		userId = int64(bill.UserId)

		// 是否已支付
		if bill.IsPaid == 1 {
			return nil
		}

		if bill.Amount <= 0 {
			// 直接修改为已支付
			return models.SharedUserBillDAO.UpdateUserBillIsPaid(tx, req.UserBillId, true)
		}

		// 余额是否足够
		account, err := accounts.SharedUserAccountDAO.FindUserAccountWithUserId(tx, userId)
		if err != nil {
			return err
		}
		if account == nil {
			return errors.New("can not find user account")
		}

		if account.Total < bill.Amount {
			return errors.New("not enough balance to pay")
		}

		err = accounts.SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), -float32(bill.Amount), userconfigs.AccountEventTypePayBill, "支付账单"+bill.Code, maps.Map{"billId": bill.Id})
		if err != nil {
			return err
		}

		// 修改为已支付
		return models.SharedUserBillDAO.UpdateUserBillIsPaid(tx, req.UserBillId, true)
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// SumUserUnpaidBills 计算用户所有未支付账单总额
func (this *UserBillService) SumUserUnpaidBills(ctx context.Context, req *pb.SumUserUnpaidBillsRequest) (*pb.SumUserUnpaidBillsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	sum, err := models.SharedUserBillDAO.SumUnpaidUserBill(tx, userId)
	if err != nil {
		return nil, err
	}
	return &pb.SumUserUnpaidBillsResponse{Amount: sum}, nil
}
