package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
)

type AdminService struct {
	BaseService

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
		AdminId: adminId,
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

	// TODO 检查权限

	admin, err := models.SharedAdminDAO.FindEnabledAdmin(req.AdminId)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return &pb.FindEnabledAdminResponse{Admin: nil}, nil
	}

	pbModules := []*pb.AdminModule{}
	modules := []*systemconfigs.AdminModule{}
	if len(admin.Modules) > 0 && admin.Modules != "null" {
		err = json.Unmarshal([]byte(admin.Modules), &modules)
		if err != nil {
			return nil, err
		}
		for _, module := range modules {
			pbModules = append(pbModules, &pb.AdminModule{
				AllowAll: module.AllowAll,
				Code:     module.Code,
				Actions:  module.Actions,
			})
		}
	}

	result := &pb.Admin{
		Id:       int64(admin.Id),
		Fullname: admin.Fullname,
		Username: admin.Username,
		IsOn:     admin.IsOn == 1,
		IsSuper:  admin.IsSuper == 1,
		Modules:  pbModules,
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
	adminId, err = models.SharedAdminDAO.CreateAdmin(req.Username, req.Password, "管理员", true, nil)
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateAdminResponse{AdminId: adminId}, nil
}

// 修改管理员信息
func (this *AdminService) UpdateAdminInfo(ctx context.Context, req *pb.UpdateAdminInfoRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	err = models.SharedAdminDAO.UpdateAdminInfo(req.AdminId, req.Fullname)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 修改管理员登录信息
func (this *AdminService) UpdateAdminLogin(ctx context.Context, req *pb.UpdateAdminLoginRequest) (*pb.RPCSuccess, error) {
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
	return this.Success()
}

// 获取所有管理员的权限列表
func (this *AdminService) FindAllAdminModules(ctx context.Context, req *pb.FindAllAdminModulesRequest) (*pb.FindAllAdminModulesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	admins, err := models.SharedAdminDAO.FindAllAdminModules()
	if err != nil {
		return nil, err
	}

	result := []*pb.AdminModuleList{}
	for _, admin := range admins {
		modules := []*systemconfigs.AdminModule{}
		if len(admin.Modules) > 0 && admin.Modules != "null" {
			err = json.Unmarshal([]byte(admin.Modules), &modules)
			if err != nil {
				return nil, err
			}
		}
		pbModules := []*pb.AdminModule{}
		for _, module := range modules {
			pbModules = append(pbModules, &pb.AdminModule{
				AllowAll: module.AllowAll,
				Code:     module.Code,
				Actions:  module.Actions,
			})
		}

		list := &pb.AdminModuleList{
			AdminId:  int64(admin.Id),
			IsSuper:  admin.IsSuper == 1,
			Fullname: admin.Fullname,
			Modules:  pbModules,
		}
		result = append(result, list)
	}

	return &pb.FindAllAdminModulesResponse{AdminModules: result}, nil
}

// 创建管理员
func (this *AdminService) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest) (*pb.CreateAdminResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	adminId, err := models.SharedAdminDAO.CreateAdmin(req.Username, req.Password, req.Fullname, req.IsSuper, req.ModulesJSON)
	if err != nil {
		return nil, err
	}
	return &pb.CreateAdminResponse{AdminId: adminId}, nil
}

// 修改管理员
func (this *AdminService) UpdateAdmin(ctx context.Context, req *pb.UpdateAdminRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	err = models.SharedAdminDAO.UpdateAdmin(req.AdminId, req.Username, req.Password, req.Fullname, req.IsSuper, req.ModulesJSON, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 计算管理员数量
func (this *AdminService) CountAllEnabledAdmins(ctx context.Context, req *pb.CountAllEnabledAdminsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	count, err := models.SharedAdminDAO.CountAllEnabledAdmins()
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 列出单页的管理员
func (this *AdminService) ListEnabledAdmins(ctx context.Context, req *pb.ListEnabledAdminsRequest) (*pb.ListEnabledAdminsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	admins, err := models.SharedAdminDAO.ListEnabledAdmins(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.Admin{}
	for _, admin := range admins {
		result = append(result, &pb.Admin{
			Id:        int64(admin.Id),
			Fullname:  admin.Fullname,
			Username:  admin.Username,
			IsOn:      admin.IsOn == 1,
			IsSuper:   admin.IsSuper == 1,
			CreatedAt: int64(admin.CreatedAt),
		})
	}

	return &pb.ListEnabledAdminsResponse{Admins: result}, nil
}

// 删除管理员
func (this *AdminService) DeleteAdmin(ctx context.Context, req *pb.DeleteAdminRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	// TODO 超级管理员用户是不能删除的，或者要至少留一个超级管理员用户

	_, err = models.SharedAdminDAO.DisableAdmin(req.AdminId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
