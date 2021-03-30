package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
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

	tx := this.NullTx()

	adminId, err := models.SharedAdminDAO.CheckAdminPassword(tx, req.Username, req.Password)
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

	tx := this.NullTx()

	ok, err := models.SharedAdminDAO.ExistEnabledAdmin(tx, req.AdminId)
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

	tx := this.NullTx()

	exists, err := models.SharedAdminDAO.CheckAdminUsername(tx, req.AdminId, req.Username)
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

	tx := this.NullTx()

	fullname, err := models.SharedAdminDAO.FindAdminFullname(tx, req.AdminId)
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

	tx := this.NullTx()

	admin, err := models.SharedAdminDAO.FindEnabledAdmin(tx, req.AdminId)
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

	// OTP认证
	var pbOtpAuth *pb.Login = nil
	{
		adminAuth, err := models.SharedLoginDAO.FindEnabledLoginWithAdminId(tx, int64(admin.Id), models.LoginTypeOTP)
		if err != nil {
			return nil, err
		}
		if adminAuth != nil {
			pbOtpAuth = &pb.Login{
				Id:         int64(adminAuth.Id),
				Type:       adminAuth.Type,
				ParamsJSON: []byte(adminAuth.Params),
				IsOn:       adminAuth.IsOn == 1,
			}
		}
	}

	result := &pb.Admin{
		Id:       int64(admin.Id),
		Fullname: admin.Fullname,
		Username: admin.Username,
		IsOn:     admin.IsOn == 1,
		IsSuper:  admin.IsSuper == 1,
		Modules:  pbModules,
		OtpLogin: pbOtpAuth,
		CanLogin: admin.CanLogin == 1,
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

	tx := this.NullTx()

	adminId, err := models.SharedAdminDAO.FindAdminIdWithUsername(tx, req.Username)
	if err != nil {
		return nil, err
	}
	if adminId > 0 {
		err = models.SharedAdminDAO.UpdateAdminPassword(tx, adminId, req.Password)
		if err != nil {
			return nil, err
		}
		return &pb.CreateOrUpdateAdminResponse{AdminId: adminId}, nil
	}
	adminId, err = models.SharedAdminDAO.CreateAdmin(tx, req.Username, true, req.Password, "管理员", true, nil)
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

	tx := this.NullTx()

	err = models.SharedAdminDAO.UpdateAdminInfo(tx, req.AdminId, req.Fullname)
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

	tx := this.NullTx()

	exists, err := models.SharedAdminDAO.CheckAdminUsername(tx, req.AdminId, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already been token")
	}

	err = models.SharedAdminDAO.UpdateAdminLogin(tx, req.AdminId, req.Username, req.Password)
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

	tx := this.NullTx()

	admins, err := models.SharedAdminDAO.FindAllAdminModules(tx)
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

	tx := this.NullTx()

	adminId, err := models.SharedAdminDAO.CreateAdmin(tx, req.Username, req.CanLogin, req.Password, req.Fullname, req.IsSuper, req.ModulesJSON)
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

	tx := this.NullTx()

	err = models.SharedAdminDAO.UpdateAdmin(tx, req.AdminId, req.Username, req.CanLogin, req.Password, req.Fullname, req.IsSuper, req.ModulesJSON, req.IsOn)
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

	tx := this.NullTx()

	count, err := models.SharedAdminDAO.CountAllEnabledAdmins(tx)
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

	tx := this.NullTx()

	admins, err := models.SharedAdminDAO.ListEnabledAdmins(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.Admin{}
	for _, admin := range admins {
		var pbOtpAuth *pb.Login = nil
		{
			adminAuth, err := models.SharedLoginDAO.FindEnabledLoginWithAdminId(tx, int64(admin.Id), models.LoginTypeOTP)
			if err != nil {
				return nil, err
			}
			if adminAuth != nil {
				pbOtpAuth = &pb.Login{
					Id:         int64(adminAuth.Id),
					Type:       adminAuth.Type,
					ParamsJSON: []byte(adminAuth.Params),
					IsOn:       adminAuth.IsOn == 1,
				}
			}
		}

		result = append(result, &pb.Admin{
			Id:        int64(admin.Id),
			Fullname:  admin.Fullname,
			Username:  admin.Username,
			IsOn:      admin.IsOn == 1,
			IsSuper:   admin.IsSuper == 1,
			CreatedAt: int64(admin.CreatedAt),
			OtpLogin:  pbOtpAuth,
			CanLogin:  admin.CanLogin == 1,
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

	tx := this.NullTx()

	// TODO 超级管理员用户是不能删除的，或者要至少留一个超级管理员用户

	_, err = models.SharedAdminDAO.DisableAdmin(tx, req.AdminId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 检查是否需要输入OTP
func (this *AdminService) CheckAdminOTPWithUsername(ctx context.Context, req *pb.CheckAdminOTPWithUsernameRequest) (*pb.CheckAdminOTPWithUsernameResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if len(req.Username) == 0 {
		return &pb.CheckAdminOTPWithUsernameResponse{RequireOTP: false}, nil
	}

	tx := this.NullTx()

	adminId, err := models.SharedAdminDAO.FindAdminIdWithUsername(tx, req.Username)
	if err != nil {
		return nil, err
	}
	if adminId <= 0 {
		return &pb.CheckAdminOTPWithUsernameResponse{RequireOTP: false}, nil
	}

	otpIsOn, err := models.SharedLoginDAO.CheckLoginIsOn(tx, adminId, "otp")
	if err != nil {
		return nil, err
	}
	return &pb.CheckAdminOTPWithUsernameResponse{RequireOTP: otpIsOn}, nil
}

// 取得管理员Dashboard数据
func (this *AdminService) ComposeAdminDashboard(ctx context.Context, req *pb.ComposeAdminDashboardRequest) (*pb.ComposeAdminDashboardResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	resp := &pb.ComposeAdminDashboardResponse{}

	var tx = this.NullTx()

	// 集群数
	countClusters, err := models.SharedNodeClusterDAO.CountAllEnabledClusters(tx, "")
	if err != nil {
		return nil, err
	}
	resp.CountNodeClusters = countClusters

	// 节点数
	countNodes, err := models.SharedNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	resp.CountNodes = countNodes

	// 服务数
	countServers, err := models.SharedServerDAO.CountAllEnabledServers(tx)
	if err != nil {
		return nil, err
	}
	resp.CountServers = countServers

	// 用户数
	countUsers, err := models.SharedUserDAO.CountAllEnabledUsers(tx, "")
	if err != nil {
		return nil, err
	}
	resp.CountUsers = countUsers

	// API节点数
	countAPINodes, err := models.SharedAPINodeDAO.CountAllEnabledAPINodes(tx)
	if err != nil {
		return nil, err
	}
	resp.CountAPINodes = countAPINodes

	// 数据库节点数
	countDBNodes, err := models.SharedDBNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	resp.CountDBNodes = countDBNodes

	// 用户节点数
	countUserNodes, err := models.SharedUserNodeDAO.CountAllEnabledUserNodes(tx)
	if err != nil {
		return nil, err
	}
	resp.CountUserNodes = countUserNodes

	// 按日流量统计
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := stats.SharedTrafficDailyStatDAO.FindDailyStats(tx, dayFrom, timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	for _, stat := range dailyTrafficStats {
		resp.DailyTrafficStats = append(resp.DailyTrafficStats, &pb.ComposeAdminDashboardResponse_DailyTrafficStat{
			Day:   stat.Day,
			Bytes: int64(stat.Bytes),
		})
	}

	// 小时流量统计
	hourFrom := timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	hourTo := timeutil.Format("YmdH")
	hourlyTrafficStats, err := stats.SharedTrafficHourlyStatDAO.FindHourlyStats(tx, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range hourlyTrafficStats {
		resp.HourlyTrafficStats = append(resp.HourlyTrafficStats, &pb.ComposeAdminDashboardResponse_HourlyTrafficStat{
			Hour:  stat.Hour,
			Bytes: int64(stat.Bytes),
		})
	}

	return resp, nil
}
