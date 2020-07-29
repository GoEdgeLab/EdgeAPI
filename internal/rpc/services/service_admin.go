package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
)

type AdminService struct {
	debug bool
}

func (this *AdminService) LoginAdmin(ctx context.Context, req *pb.LoginAdminRequest) (*pb.LoginAdminResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.Username) == 0 || len(req.Password) == 0 {
		return &pb.LoginAdminResponse{
			AdminId: 0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	adminId, err := models.SharedAdminDAO.CheckAdminPassword(req.Username, req.Password)
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	if adminId <= 0 {
		return &pb.LoginAdminResponse{
			AdminId: 0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	return &pb.LoginAdminResponse{
		AdminId: int64(adminId),
		IsOk:    true,
	}, nil
}

func (this *AdminService) CreateAdminLog(ctx context.Context, req *pb.CreateAdminLogRequest) (*pb.CreateAdminLogResponse, error) {
	_, userId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	err = models.SharedLogDAO.CreateAdminLog(userId, req.Level, req.Description, req.Action, req.Ip)
	return &pb.CreateAdminLogResponse{
		IsOk: err != nil,
	}, err
}

func (this *AdminService) CheckAdminExists(ctx context.Context, req *pb.CheckAdminExistsRequest) (*pb.CheckAdminExistsResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.AdminId <= 0 {
		return &pb.CheckAdminExistsResponse{
			IsOk: false,
		}, nil
	}

	ok, err := models.SharedAdminDAO.ExistEnabledAdmin(int(req.AdminId))
	if err != nil {
		return nil, err
	}

	return &pb.CheckAdminExistsResponse{
		IsOk: ok,
	}, nil
}

func (this *AdminService) FindAdminFullname(ctx context.Context, req *pb.FindAdminNameRequest) (*pb.FindAdminNameResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	fullname, err := models.SharedAdminDAO.FindAdminFullname(int(req.AdminId))
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	return &pb.FindAdminNameResponse{
		Fullname: fullname,
	}, nil
}