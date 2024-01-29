package services

import (
	"context"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/tasks"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// AdminService 管理员相关服务
type AdminService struct {
	BaseService

	debug bool
}

// LoginAdmin 登录
func (this *AdminService) LoginAdmin(ctx context.Context, req *pb.LoginAdminRequest) (*pb.LoginAdminResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx)
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

	var tx = this.NullTx()

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

// CheckAdminExists 检查管理员是否存在
func (this *AdminService) CheckAdminExists(ctx context.Context, req *pb.CheckAdminExistsRequest) (*pb.CheckAdminExistsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.AdminId <= 0 {
		return &pb.CheckAdminExistsResponse{
			IsOk: false,
		}, nil
	}

	var tx = this.NullTx()

	ok, err := models.SharedAdminDAO.ExistEnabledAdmin(tx, req.AdminId)
	if err != nil {
		return nil, err
	}

	return &pb.CheckAdminExistsResponse{
		IsOk: ok,
	}, nil
}

// CheckAdminUsername 检查用户名是否存在
func (this *AdminService) CheckAdminUsername(ctx context.Context, req *pb.CheckAdminUsernameRequest) (*pb.CheckAdminUsernameResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	exists, err := models.SharedAdminDAO.CheckAdminUsername(tx, req.AdminId, req.Username)
	if err != nil {
		return nil, err
	}

	return &pb.CheckAdminUsernameResponse{Exists: exists}, nil
}

// FindAdminWithUsername 使用用管理员户名查找管理员信息
func (this *AdminService) FindAdminWithUsername(ctx context.Context, req *pb.FindAdminWithUsernameRequest) (*pb.FindAdminWithUsernameResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if len(req.Username) == 0 {
		return nil, errors.New("require 'username'")
	}
	admin, err := models.SharedAdminDAO.FindAdminWithUsername(tx, req.Username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return &pb.FindAdminWithUsernameResponse{Admin: nil}, nil
	}

	return &pb.FindAdminWithUsernameResponse{
		Admin: &pb.Admin{
			Id:       int64(admin.Id),
			Fullname: admin.Fullname,
			Username: admin.Username,
			IsOn:     admin.IsOn,
			IsSuper:  admin.IsSuper,
			CanLogin: admin.CanLogin,
		},
	}, nil
}

// FindAdminFullname 获取管理员名称
func (this *AdminService) FindAdminFullname(ctx context.Context, req *pb.FindAdminFullnameRequest) (*pb.FindAdminFullnameResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	fullname, err := models.SharedAdminDAO.FindAdminFullname(tx, req.AdminId)
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	return &pb.FindAdminFullnameResponse{
		Fullname: fullname,
	}, nil
}

// FindEnabledAdmin 获取管理员信息
func (this *AdminService) FindEnabledAdmin(ctx context.Context, req *pb.FindEnabledAdminRequest) (*pb.FindEnabledAdminResponse, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	var tx = this.NullTx()

	// 超级管理员才能查看是否为弱密码
	isSuperAdmin, err := models.SharedAdminDAO.CheckSuperAdmin(tx, adminId)
	if err != nil {
		return nil, err
	}

	admin, err := models.SharedAdminDAO.FindEnabledAdmin(tx, req.AdminId)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return &pb.FindEnabledAdminResponse{Admin: nil}, nil
	}

	var pbModules = []*pb.AdminModule{}
	modules := []*systemconfigs.AdminModule{}
	if len(admin.Modules) > 0 {
		err = json.Unmarshal(admin.Modules, &modules)
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
		adminAuth, err := models.SharedLoginDAO.FindEnabledLoginWithType(tx, int64(admin.Id), 0, models.LoginTypeOTP)
		if err != nil {
			return nil, err
		}
		if adminAuth != nil {
			pbOtpAuth = &pb.Login{
				Id:         int64(adminAuth.Id),
				Type:       adminAuth.Type,
				ParamsJSON: adminAuth.Params,
				IsOn:       adminAuth.IsOn,
			}
		}
	}

	result := &pb.Admin{
		Id:              int64(admin.Id),
		Fullname:        admin.Fullname,
		Username:        admin.Username,
		IsOn:            admin.IsOn,
		IsSuper:         admin.IsSuper,
		Modules:         pbModules,
		OtpLogin:        pbOtpAuth,
		CanLogin:        admin.CanLogin,
		HasWeakPassword: isSuperAdmin && admin.HasWeakPassword(),
	}
	return &pb.FindEnabledAdminResponse{Admin: result}, nil
}

// CreateOrUpdateAdmin 创建或修改管理员
func (this *AdminService) CreateOrUpdateAdmin(ctx context.Context, req *pb.CreateOrUpdateAdminRequest) (*pb.CreateOrUpdateAdminResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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

// UpdateAdminInfo 修改管理员信息
func (this *AdminService) UpdateAdminInfo(ctx context.Context, req *pb.UpdateAdminInfoRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedAdminDAO.UpdateAdminInfo(tx, req.AdminId, req.Fullname)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateAdminLogin 修改管理员登录信息
func (this *AdminService) UpdateAdminLogin(ctx context.Context, req *pb.UpdateAdminLoginRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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

// FindAllAdminModules 获取所有管理员的权限列表
func (this *AdminService) FindAllAdminModules(ctx context.Context, req *pb.FindAllAdminModulesRequest) (*pb.FindAllAdminModulesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	var tx = this.NullTx()

	admins, err := models.SharedAdminDAO.FindAllAdminModules(tx)
	if err != nil {
		return nil, err
	}

	var result = []*pb.AdminModuleList{}
	for _, admin := range admins {
		modules := []*systemconfigs.AdminModule{}
		if len(admin.Modules) > 0 {
			err = json.Unmarshal(admin.Modules, &modules)
			if err != nil {
				return nil, err
			}
		}
		var pbModules = []*pb.AdminModule{}
		for _, module := range modules {
			pbModules = append(pbModules, &pb.AdminModule{
				AllowAll: module.AllowAll,
				Code:     module.Code,
				Actions:  module.Actions,
			})
		}

		var list = &pb.AdminModuleList{
			AdminId:  int64(admin.Id),
			IsSuper:  admin.IsSuper,
			Fullname: admin.Fullname,
			Theme:    admin.Theme,
			Lang:     admin.Lang,
			Modules:  pbModules,
		}
		result = append(result, list)
	}

	return &pb.FindAllAdminModulesResponse{AdminModules: result}, nil
}

// CreateAdmin 创建管理员
func (this *AdminService) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest) (*pb.CreateAdminResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	var tx = this.NullTx()

	adminId, err := models.SharedAdminDAO.CreateAdmin(tx, req.Username, req.CanLogin, req.Password, req.Fullname, req.IsSuper, req.ModulesJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateAdminResponse{AdminId: adminId}, nil
}

// UpdateAdmin 修改管理员
func (this *AdminService) UpdateAdmin(ctx context.Context, req *pb.UpdateAdminRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	var tx = this.NullTx()

	err = models.SharedAdminDAO.UpdateAdmin(tx, req.AdminId, req.Username, req.CanLogin, req.Password, req.Fullname, req.IsSuper, req.ModulesJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllEnabledAdmins 计算管理员数量
func (this *AdminService) CountAllEnabledAdmins(ctx context.Context, req *pb.CountAllEnabledAdminsRequest) (*pb.RPCCountResponse, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	var tx = this.NullTx()

	// 超级管理员才能查看是否为弱密码
	isSuperAdmin, err := models.SharedAdminDAO.CheckSuperAdmin(tx, adminId)
	if err != nil {
		return nil, err
	}

	if !isSuperAdmin && req.HasWeakPassword {
		return this.SuccessCount(0)
	}

	count, err := models.SharedAdminDAO.CountAllEnabledAdmins(tx, req.Keyword, req.HasWeakPassword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledAdmins 列出单页的管理员
func (this *AdminService) ListEnabledAdmins(ctx context.Context, req *pb.ListEnabledAdminsRequest) (*pb.ListEnabledAdminsResponse, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	var tx = this.NullTx()

	// 超级管理员才能查看是否为弱密码
	isSuperAdmin, err := models.SharedAdminDAO.CheckSuperAdmin(tx, adminId)
	if err != nil {
		return nil, err
	}

	if !isSuperAdmin && req.HasWeakPassword {
		return &pb.ListEnabledAdminsResponse{Admins: nil}, nil
	}

	admins, err := models.SharedAdminDAO.ListEnabledAdmins(tx, req.Keyword, req.HasWeakPassword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var result = []*pb.Admin{}
	for _, admin := range admins {
		var pbOtpAuth *pb.Login = nil
		{
			adminAuth, err := models.SharedLoginDAO.FindEnabledLoginWithType(tx, int64(admin.Id), 0, models.LoginTypeOTP)
			if err != nil {
				return nil, err
			}
			if adminAuth != nil {
				pbOtpAuth = &pb.Login{
					Id:         int64(adminAuth.Id),
					Type:       adminAuth.Type,
					ParamsJSON: adminAuth.Params,
					IsOn:       adminAuth.IsOn,
				}
			}
		}

		result = append(result, &pb.Admin{
			Id:              int64(admin.Id),
			Fullname:        admin.Fullname,
			Username:        admin.Username,
			IsOn:            admin.IsOn,
			IsSuper:         admin.IsSuper,
			CreatedAt:       int64(admin.CreatedAt),
			OtpLogin:        pbOtpAuth,
			CanLogin:        admin.CanLogin,
			HasWeakPassword: isSuperAdmin && admin.HasWeakPassword(),
		})
	}

	return &pb.ListEnabledAdminsResponse{Admins: result}, nil
}

// DeleteAdmin 删除管理员
func (this *AdminService) DeleteAdmin(ctx context.Context, req *pb.DeleteAdminRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	var tx = this.NullTx()

	// TODO 超级管理员用户是不能删除的，或者要至少留一个超级管理员用户

	err = models.SharedAdminDAO.DisableAdmin(tx, req.AdminId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CheckAdminOTPWithUsername 检查是否需要输入OTP
func (this *AdminService) CheckAdminOTPWithUsername(ctx context.Context, req *pb.CheckAdminOTPWithUsernameRequest) (*pb.CheckAdminOTPWithUsernameResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.Username) == 0 {
		return &pb.CheckAdminOTPWithUsernameResponse{RequireOTP: false}, nil
	}

	var tx = this.NullTx()

	adminId, err := models.SharedAdminDAO.FindAdminIdWithUsername(tx, req.Username)
	if err != nil {
		return nil, err
	}
	if adminId <= 0 {
		return &pb.CheckAdminOTPWithUsernameResponse{RequireOTP: false}, nil
	}

	otpIsOn, err := models.SharedLoginDAO.CheckLoginIsOn(tx, adminId, 0, "otp")
	if err != nil {
		return nil, err
	}
	return &pb.CheckAdminOTPWithUsernameResponse{RequireOTP: otpIsOn}, nil
}

// ComposeAdminDashboard 取得管理员Dashboard数据
func (this *AdminService) ComposeAdminDashboard(ctx context.Context, req *pb.ComposeAdminDashboardRequest) (*pb.ComposeAdminDashboardResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	result := &pb.ComposeAdminDashboardResponse{}

	var tx = this.NullTx()

	// 默认集群
	this.BeginTag(ctx, "SharedNodeClusterDAO.ListEnabledClusters")
	nodeClusters, err := models.SharedNodeClusterDAO.ListEnabledClusters(tx, "", true, false, 0, 1)
	this.EndTag(ctx, "SharedNodeClusterDAO.ListEnabledClusters")
	if err != nil {
		return nil, err
	}
	if len(nodeClusters) > 0 {
		result.DefaultNodeClusterId = int64(nodeClusters[0].Id)
	}

	// 集群数
	this.BeginTag(ctx, "SharedNodeClusterDAO.CountAllEnabledClusters")
	countClusters, err := models.SharedNodeClusterDAO.CountAllEnabledClusters(tx, "")
	this.EndTag(ctx, "SharedNodeClusterDAO.CountAllEnabledClusters")
	if err != nil {
		return nil, err
	}
	result.CountNodeClusters = countClusters

	// 节点数
	this.BeginTag(ctx, "SharedNodeDAO.CountAllEnabledNodes")
	countNodes, err := models.SharedNodeDAO.CountAllEnabledNodes(tx)
	this.EndTag(ctx, "SharedNodeDAO.CountAllEnabledNodes")
	if err != nil {
		return nil, err
	}
	result.CountNodes = countNodes

	// 离线节点
	this.BeginTag(ctx, "SharedNodeDAO.CountAllEnabledOfflineNodes")
	countOfflineNodes, err := models.SharedNodeDAO.CountAllEnabledOfflineNodes(tx)
	this.EndTag(ctx, "SharedNodeDAO.CountAllEnabledOfflineNodes")
	if err != nil {
		return nil, err
	}
	result.CountOfflineNodes = countOfflineNodes

	// 服务数
	this.BeginTag(ctx, "SharedServerDAO.CountAllEnabledServers")
	countServers, err := models.SharedServerDAO.CountAllEnabledServers(tx)
	this.EndTag(ctx, "SharedServerDAO.CountAllEnabledServers")
	if err != nil {
		return nil, err
	}
	result.CountServers = countServers

	this.BeginTag(ctx, "SharedServerDAO.CountAllEnabledServersMatch")
	countAuditingServers, err := models.SharedServerDAO.CountAllEnabledServersMatch(tx, 0, "", 0, 0, configutils.BoolStateYes, nil, 0)
	this.EndTag(ctx, "SharedServerDAO.CountAllEnabledServersMatch")
	if err != nil {
		return nil, err
	}
	result.CountAuditingServers = countAuditingServers

	// 用户数
	this.BeginTag(ctx, "SharedUserDAO.CountAllEnabledUsers")
	countUsers, err := models.SharedUserDAO.CountAllEnabledUsers(tx, 0, "", false)
	this.EndTag(ctx, "SharedUserDAO.CountAllEnabledUsers")
	if err != nil {
		return nil, err
	}
	result.CountUsers = countUsers

	// API节点数
	this.BeginTag(ctx, "SharedAPINodeDAO.CountAllEnabledAndOnAPINodes")
	countAPINodes, err := models.SharedAPINodeDAO.CountAllEnabledAndOnAPINodes(tx)
	this.EndTag(ctx, "SharedAPINodeDAO.CountAllEnabledAndOnAPINodes")
	if err != nil {
		return nil, err
	}
	result.CountAPINodes = countAPINodes

	// 离线API节点
	this.BeginTag(ctx, "SharedAPINodeDAO.CountAllEnabledAndOnOfflineAPINodes")
	countOfflineAPINodes, err := models.SharedAPINodeDAO.CountAllEnabledAndOnOfflineAPINodes(tx)
	this.EndTag(ctx, "SharedAPINodeDAO.CountAllEnabledAndOnOfflineAPINodes")
	if err != nil {
		return nil, err
	}
	result.CountOfflineAPINodes = countOfflineAPINodes

	// 数据库节点数
	this.BeginTag(ctx, "SharedDBNodeDAO.CountAllEnabledNodes")
	countDBNodes, err := models.SharedDBNodeDAO.CountAllEnabledNodes(tx)
	this.EndTag(ctx, "SharedDBNodeDAO.CountAllEnabledNodes")
	if err != nil {
		return nil, err
	}
	result.CountDBNodes = countDBNodes

	// 用户节点数
	this.BeginTag(ctx, "SharedUserNodeDAO.CountAllEnabledAndOnUserNodes")
	countUserNodes, err := models.SharedUserNodeDAO.CountAllEnabledAndOnUserNodes(tx)
	this.EndTag(ctx, "SharedUserNodeDAO.CountAllEnabledAndOnUserNodes")
	if err != nil {
		return nil, err
	}
	result.CountUserNodes = countUserNodes

	// 离线用户节点数
	this.BeginTag(ctx, "SharedUserNodeDAO.CountAllEnabledAndOnOfflineNodes")
	countOfflineUserNodes, err := models.SharedUserNodeDAO.CountAllEnabledAndOnOfflineNodes(tx)
	this.EndTag(ctx, "SharedUserNodeDAO.CountAllEnabledAndOnOfflineNodes")
	if err != nil {
		return nil, err
	}
	result.CountOfflineUserNodes = countOfflineUserNodes

	// 按日流量统计
	this.BeginTag(ctx, "SharedTrafficDailyStatDAO.FindDailyStats")
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := stats.SharedTrafficDailyStatDAO.FindDailyStats(tx, dayFrom, timeutil.Format("Ymd"))
	this.EndTag(ctx, "SharedTrafficDailyStatDAO.FindDailyStats")
	if err != nil {
		return nil, err
	}
	for _, stat := range dailyTrafficStats {
		result.DailyTrafficStats = append(result.DailyTrafficStats, &pb.ComposeAdminDashboardResponse_DailyTrafficStat{
			Day:                 stat.Day,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	// 小时流量统计
	var hourFrom = timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	var hourTo = timeutil.Format("YmdH")
	this.BeginTag(ctx, "SharedTrafficHourlyStatDAO.FindHourlyStats")
	hourlyTrafficStats, err := stats.SharedTrafficHourlyStatDAO.FindHourlyStats(tx, hourFrom, hourTo)
	this.EndTag(ctx, "SharedTrafficHourlyStatDAO.FindHourlyStats")
	if err != nil {
		return nil, err
	}
	for _, stat := range hourlyTrafficStats {
		result.HourlyTrafficStats = append(result.HourlyTrafficStats, &pb.ComposeAdminDashboardResponse_HourlyTrafficStat{
			Hour:                stat.Hour,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	// 边缘节点升级信息
	{
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: teaconst.NodeVersion,
		}
		this.BeginTag(ctx, "SharedNodeDAO.CountAllLowerVersionNodes")
		countNodes, err := models.SharedNodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		this.EndTag(ctx, "SharedNodeDAO.CountAllLowerVersionNodes")
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.NodeUpgradeInfo = upgradeInfo
	}

	// API节点升级信息
	{
		var apiVersion = req.ApiVersion
		if len(apiVersion) == 0 {
			apiVersion = teaconst.Version
		}
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: apiVersion,
		}
		this.BeginTag(ctx, "SharedAPINodeDAO.CountAllLowerVersionNodes")
		countNodes, err := models.SharedAPINodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		this.EndTag(ctx, "SharedAPINodeDAO.CountAllLowerVersionNodes")
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.ApiNodeUpgradeInfo = upgradeInfo
	}

	// 额外的检查节点版本
	err = this.composeAdminDashboardExt(tx, ctx, result)
	if err != nil {
		return nil, err
	}

	// 域名排行
	this.BeginTag(ctx, "SharedServerDomainHourlyStatDAO.FindTopDomainStats")
	var topDomainStats []*stats.ServerDomainHourlyStat
	topDomainStatsCache, ok := tasks.SharedCacheTaskManager.GetGlobalTopDomains()
	if ok {
		topDomainStats = topDomainStatsCache.([]*stats.ServerDomainHourlyStat)
	}
	this.EndTag(ctx, "SharedServerDomainHourlyStatDAO.FindTopDomainStats")
	if err != nil {
		return nil, err
	}
	for _, stat := range topDomainStats {
		result.TopDomainStats = append(result.TopDomainStats, &pb.ComposeAdminDashboardResponse_DomainStat{
			ServerId:      int64(stat.ServerId),
			Domain:        stat.Domain,
			CountRequests: int64(stat.CountRequests),
			Bytes:         int64(stat.Bytes),
		})
	}

	// 指标数据
	this.BeginTag(ctx, "findMetricDataCharts")
	var pbCharts []*pb.MetricDataChart
	pbChartsCache, ok := tasks.SharedCacheTaskManager.Get(tasks.CacheKeyFindAllMetricDataCharts)
	if ok {
		pbCharts = pbChartsCache.([]*pb.MetricDataChart)
	}
	this.EndTag(ctx, "findMetricDataCharts")
	if err != nil {
		return nil, err
	}
	result.MetricDataCharts = pbCharts

	return result, nil
}

// UpdateAdminTheme 修改管理员使用的界面风格
func (this *AdminService) UpdateAdminTheme(ctx context.Context, req *pb.UpdateAdminThemeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = models.SharedAdminDAO.UpdateAdminTheme(tx, req.AdminId, req.Theme)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateAdminLang 修改管理员使用的语言
func (this *AdminService) UpdateAdminLang(ctx context.Context, req *pb.UpdateAdminLangRequest) (*pb.RPCSuccess, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()

	err = models.SharedAdminDAO.UpdateAdminLang(tx, adminId, req.LangCode)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
