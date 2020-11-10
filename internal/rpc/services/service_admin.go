package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type AdminService struct {
	debug bool
}

// 登录
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

// 检查管理员是否存在
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

	ok, err := models.SharedAdminDAO.ExistEnabledAdmin(req.AdminId)
	if err != nil {
		return nil, err
	}

	return &pb.CheckAdminExistsResponse{
		IsOk: ok,
	}, nil
}

// 检查用户名是否存在
func (this *AdminService) CheckAdminUsername(ctx context.Context, req *pb.CheckAdminUsernameRequest) (*pb.CheckAdminUsernameResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	exists, err := models.SharedAdminDAO.CheckAdminUsername(req.AdminId, req.Username)
	if err != nil {
		return nil, err
	}

	return &pb.CheckAdminUsernameResponse{Exists: exists}, nil
}

// 获取管理员名称
func (this *AdminService) FindAdminFullname(ctx context.Context, req *pb.FindAdminFullnameRequest) (*pb.FindAdminFullnameResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	fullname, err := models.SharedAdminDAO.FindAdminFullname(req.AdminId)
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	return &pb.FindAdminFullnameResponse{
		Fullname: fullname,
	}, nil
}

// 获取管理员信息
func (this *AdminService) FindEnabledAdmin(ctx context.Context, req *pb.FindEnabledAdminRequest) (*pb.FindEnabledAdminResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	admin, err := models.SharedAdminDAO.FindEnabledAdmin(req.AdminId)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return &pb.FindEnabledAdminResponse{Admin: nil}, nil
	}

	result := &pb.Admin{
		Id:       int64(admin.Id),
		Fullname: admin.Fullname,
		Username: admin.Username,
		IsOn:     admin.IsOn == 1,
	}
	return &pb.FindEnabledAdminResponse{Admin: result}, nil
}

// 创建或修改管理员
func (this *AdminService) CreateOrUpdateAdmin(ctx context.Context, req *pb.CreateOrUpdateAdminRequest) (*pb.CreateOrUpdateAdminResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	adminId, err := models.SharedAdminDAO.FindAdminIdWithUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if adminId > 0 {
		err = models.SharedAdminDAO.UpdateAdminPassword(adminId, req.Password)
		if err != nil {
			return nil, err
		}
		return &pb.CreateOrUpdateAdminResponse{AdminId: adminId}, nil
	}
	adminId, err = models.SharedAdminDAO.CreateAdmin(req.Username, req.Password, "管理员")
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateAdminResponse{AdminId: adminId}, nil
}

// 修改管理员信息
func (this *AdminService) UpdateAdmin(ctx context.Context, req *pb.UpdateAdminRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	err = models.SharedAdminDAO.UpdateAdmin(req.AdminId, req.Fullname)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 修改管理员登录信息
func (this *AdminService) UpdateAdminLogin(ctx context.Context, req *pb.UpdateAdminLoginRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	exists, err := models.SharedAdminDAO.CheckAdminUsername(req.AdminId, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already been token")
	}

	err = models.SharedAdminDAO.UpdateAdminLogin(req.AdminId, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}
