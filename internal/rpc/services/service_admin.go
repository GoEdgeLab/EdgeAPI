package services

import (
	"context"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/authority"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

// CheckAdminExists 检查管理员是否存在
func (this *AdminService) CheckAdminExists(ctx context.Context, req *pb.CheckAdminExistsRequest) (*pb.CheckAdminExistsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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

// CheckAdminUsername 检查用户名是否存在
func (this *AdminService) CheckAdminUsername(ctx context.Context, req *pb.CheckAdminUsernameRequest) (*pb.CheckAdminUsernameResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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

// FindAdminFullname 获取管理员名称
func (this *AdminService) FindAdminFullname(ctx context.Context, req *pb.FindAdminFullnameRequest) (*pb.FindAdminFullnameResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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

// FindEnabledAdmin 获取管理员信息
func (this *AdminService) FindEnabledAdmin(ctx context.Context, req *pb.FindEnabledAdminRequest) (*pb.FindEnabledAdminResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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

// CreateOrUpdateAdmin 创建或修改管理员
func (this *AdminService) CreateOrUpdateAdmin(ctx context.Context, req *pb.CreateOrUpdateAdminRequest) (*pb.CreateOrUpdateAdminResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
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

// UpdateAdminInfo 修改管理员信息
func (this *AdminService) UpdateAdminInfo(ctx context.Context, req *pb.UpdateAdminInfoRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
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

// UpdateAdminLogin 修改管理员登录信息
func (this *AdminService) UpdateAdminLogin(ctx context.Context, req *pb.UpdateAdminLoginRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeAPI)
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

// FindAllAdminModules 获取所有管理员的权限列表
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
			Theme:    admin.Theme,
			Modules:  pbModules,
		}
		result = append(result, list)
	}

	return &pb.FindAllAdminModulesResponse{AdminModules: result}, nil
}

// CreateAdmin 创建管理员
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

// UpdateAdmin 修改管理员
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

// CountAllEnabledAdmins 计算管理员数量
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

// ListEnabledAdmins 列出单页的管理员
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

// DeleteAdmin 删除管理员
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

// CheckAdminOTPWithUsername 检查是否需要输入OTP
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

// ComposeAdminDashboard 取得管理员Dashboard数据
func (this *AdminService) ComposeAdminDashboard(ctx context.Context, req *pb.ComposeAdminDashboardRequest) (*pb.ComposeAdminDashboardResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	result := &pb.ComposeAdminDashboardResponse{}

	var tx = this.NullTx()

	// 集群数
	countClusters, err := models.SharedNodeClusterDAO.CountAllEnabledClusters(tx, "")
	if err != nil {
		return nil, err
	}
	result.CountNodeClusters = countClusters

	// 节点数
	countNodes, err := models.SharedNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountNodes = countNodes

	// 服务数
	countServers, err := models.SharedServerDAO.CountAllEnabledServers(tx)
	if err != nil {
		return nil, err
	}
	result.CountServers = countServers

	// 用户数
	countUsers, err := models.SharedUserDAO.CountAllEnabledUsers(tx, 0, "")
	if err != nil {
		return nil, err
	}
	result.CountUsers = countUsers

	// API节点数
	countAPINodes, err := models.SharedAPINodeDAO.CountAllEnabledAPINodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountAPINodes = countAPINodes

	// 数据库节点数
	countDBNodes, err := models.SharedDBNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountDBNodes = countDBNodes

	// 用户节点数
	countUserNodes, err := models.SharedUserNodeDAO.CountAllEnabledUserNodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountUserNodes = countUserNodes

	// 按日流量统计
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := stats.SharedTrafficDailyStatDAO.FindDailyStats(tx, dayFrom, timeutil.Format("Ymd"))
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
	hourFrom := timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	hourTo := timeutil.Format("YmdH")
	hourlyTrafficStats, err := stats.SharedTrafficHourlyStatDAO.FindHourlyStats(tx, hourFrom, hourTo)
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

	// 是否是企业版
	isPlus, err := authority.SharedAuthorityKeyDAO.IsPlus(tx)
	if err != nil {
		return nil, err
	}

	// 边缘节点升级信息
	{
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: teaconst.NodeVersion,
		}
		countNodes, err := models.SharedNodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.NodeUpgradeInfo = upgradeInfo
	}

	// 监控节点升级信息
	if isPlus {
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: teaconst.MonitorNodeVersion,
		}
		countNodes, err := models.SharedMonitorNodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.MonitorNodeUpgradeInfo = upgradeInfo
	}

	// 认证节点升级信息
	if isPlus {
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: teaconst.AuthorityNodeVersion,
		}
		countNodes, err := authority.SharedAuthorityNodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.AuthorityNodeUpgradeInfo = upgradeInfo
	}

	// 用户节点升级信息
	if isPlus {
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: teaconst.UserNodeVersion,
		}
		countNodes, err := models.SharedUserNodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.UserNodeUpgradeInfo = upgradeInfo
	}

	// API节点升级信息
	{
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: teaconst.Version,
		}
		countNodes, err := models.SharedAPINodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.ApiNodeUpgradeInfo = upgradeInfo
	}

	// DNS节点升级信息
	if isPlus {
		upgradeInfo := &pb.ComposeAdminDashboardResponse_UpgradeInfo{
			NewVersion: teaconst.DNSNodeVersion,
		}
		countNodes, err := models.SharedNSNodeDAO.CountAllLowerVersionNodes(tx, upgradeInfo.NewVersion)
		if err != nil {
			return nil, err
		}
		upgradeInfo.CountNodes = countNodes
		result.NsNodeUpgradeInfo = upgradeInfo
	}

	// 域名排行
	if isPlus {
		topDomainStats, err := stats.SharedServerDomainHourlyStatDAO.FindTopDomainStats(tx, hourFrom, hourTo, 10)
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
	}

	// 节点排行
	if isPlus {
		topNodeStats, err := stats.SharedNodeTrafficHourlyStatDAO.FindTopNodeStats(tx, "node", hourFrom, hourTo)
		if err != nil {
			return nil, err
		}
		for _, stat := range topNodeStats {
			nodeName, err := models.SharedNodeDAO.FindNodeName(tx, int64(stat.NodeId))
			if err != nil {
				return nil, err
			}
			if len(nodeName) == 0 {
				continue
			}
			result.TopNodeStats = append(result.TopNodeStats, &pb.ComposeAdminDashboardResponse_NodeStat{
				NodeId:        int64(stat.NodeId),
				NodeName:      nodeName,
				CountRequests: int64(stat.CountRequests),
				Bytes:         int64(stat.Bytes),
			})
		}
	}

	// 指标数据
	pbCharts, err := this.findMetricDataCharts(tx)
	if err != nil {
		return nil, err
	}
	result.MetricDataCharts = pbCharts

	return result, nil
}

// UpdateAdminTheme 修改管理员使用的界面风格
func (this *AdminService) UpdateAdminTheme(ctx context.Context, req *pb.UpdateAdminThemeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, req.AdminId)
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

// 查找集群、节点和服务的指标数据
func (this *AdminService) findMetricDataCharts(tx *dbs.Tx) (result []*pb.MetricDataChart, err error) {
	// 集群指标
	items, err := models.SharedMetricItemDAO.FindAllPublicItems(tx)
	if err != nil {
		return nil, err
	}
	var pbMetricCharts = []*pb.MetricDataChart{}
	for _, item := range items {
		var itemId = int64(item.Id)
		charts, err := models.SharedMetricChartDAO.FindAllEnabledCharts(tx, itemId)
		if err != nil {
			return nil, err
		}

		for _, chart := range charts {
			if chart.IsOn == 0 {
				continue
			}

			var pbChart = &pb.MetricChart{
				Id:         int64(chart.Id),
				Name:       chart.Name,
				Type:       chart.Type,
				WidthDiv:   chart.WidthDiv,
				ParamsJSON: nil,
				IsOn:       chart.IsOn == 1,
				MaxItems:   types.Int32(chart.MaxItems),
				MetricItem: &pb.MetricItem{
					Id:         itemId,
					PeriodUnit: item.PeriodUnit,
					Period:     types.Int32(item.Period),
					Name:       item.Name,
					Value:      item.Value,
					Category:   item.Category,
					Keys:       item.DecodeKeys(),
					Code:       item.Code,
					IsOn:       item.IsOn == 1,
				},
			}
			var pbStats = []*pb.MetricStat{}
			switch chart.Type {
			case serverconfigs.MetricChartTypeTimeLine:
				itemStats, err := models.SharedMetricStatDAO.FindLatestItemStats(tx, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}

				for _, stat := range itemStats {
					// 当前时间总和
					count, total, err := models.SharedMetricSumStatDAO.FindSumAtTime(tx, stat.Time, itemId, types.Int32(item.Version))
					if err != nil {
						return nil, err
					}

					pbStats = append(pbStats, &pb.MetricStat{
						Id:          int64(stat.Id),
						Hash:        stat.Hash,
						ServerId:    0,
						ItemId:      0,
						Keys:        stat.DecodeKeys(),
						Value:       types.Float32(stat.Value),
						Time:        stat.Time,
						Version:     0,
						NodeCluster: nil,
						Node:        nil,
						Server:      nil,
						SumCount:    count,
						SumTotal:    total,
					})
				}
			default:
				itemStats, err := models.SharedMetricStatDAO.FindItemStatsAtLastTime(tx, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}
				for _, stat := range itemStats {
					count, total, err := models.SharedMetricSumStatDAO.FindSumAtTime(tx, stat.Time, itemId, types.Int32(item.Version))
					if err != nil {
						return nil, err
					}

					pbStats = append(pbStats, &pb.MetricStat{
						Id:          int64(stat.Id),
						Hash:        stat.Hash,
						ServerId:    0,
						ItemId:      0,
						Keys:        stat.DecodeKeys(),
						Value:       types.Float32(stat.Value),
						Time:        stat.Time,
						Version:     0,
						NodeCluster: nil,
						Node:        nil,
						Server:      nil,
						SumCount:    count,
						SumTotal:    total,
					})
				}
			}
			pbMetricCharts = append(pbMetricCharts, &pb.MetricDataChart{
				MetricChart: pbChart,
				MetricStats: pbStats,
			})
		}
	}
	return pbMetricCharts, nil
}
