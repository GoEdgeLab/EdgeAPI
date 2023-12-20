package users

import (
	"context"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strings"
	"time"
)

// UserService 用户相关服务
type UserService struct {
	services.BaseService
}

// CreateUser 创建用户
func (this *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	userId, err := models.SharedUserDAO.CreateUser(tx, req.Username, req.Password, req.Fullname, req.Mobile, req.Tel, req.Email, req.Remark, req.Source, req.NodeClusterId, nil, "", true)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{UserId: userId}, nil
}

// VerifyUser 审核用户
func (this *UserService) VerifyUser(ctx context.Context, req *pb.VerifyUserRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserDAO.UpdateUserIsVerified(tx, req.UserId, req.IsRejected, req.RejectReason)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateUser 修改用户
func (this *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	oldClusterId, err := models.SharedUserDAO.FindUserClusterId(tx, req.UserId)
	if err != nil {
		return nil, err
	}

	err = models.SharedUserDAO.UpdateUser(tx, req.UserId, req.Username, req.Password, req.Fullname, req.Mobile, req.Tel, req.Email, req.Remark, req.IsOn, req.NodeClusterId, req.BandwidthAlgo)
	if err != nil {
		return nil, err
	}

	if oldClusterId != req.NodeClusterId {
		err = models.SharedServerDAO.UpdateUserServersClusterId(tx, req.UserId, oldClusterId, req.NodeClusterId)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// DeleteUser 删除用户
func (this *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 删除其下的Server
	serverIds, err := models.SharedServerDAO.FindAllEnabledServerIdsWithUserId(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	for _, serverId := range serverIds {
		err := models.SharedServerDAO.DisableServer(tx, serverId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedUserDAO.DisableUser(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledUsers 计算用户数量
func (this *UserService) CountAllEnabledUsers(ctx context.Context, req *pb.CountAllEnabledUsersRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedUserDAO.CountAllEnabledUsers(tx, 0, req.Keyword, req.IsVerifying)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledUsers 列出单页用户
func (this *UserService) ListEnabledUsers(ctx context.Context, req *pb.ListEnabledUsersRequest) (*pb.ListEnabledUsersResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	users, err := models.SharedUserDAO.ListEnabledUsers(tx, 0, req.Keyword, req.IsVerifying, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.User{}
	for _, user := range users {
		// 集群信息
		var pbCluster *pb.NodeCluster = nil
		if user.ClusterId > 0 {
			clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(user.ClusterId))
			if err != nil {
				return nil, err
			}
			pbCluster = &pb.NodeCluster{
				Id:   int64(user.ClusterId),
				Name: clusterName,
			}
		}

		result = append(result, &pb.User{
			Id:           int64(user.Id),
			Username:     user.Username,
			Fullname:     user.Fullname,
			Mobile:       user.Mobile,
			Tel:          user.Tel,
			Email:        user.Email,
			Remark:       user.Remark,
			IsOn:         user.IsOn,
			RegisteredIP: user.RegisteredIP,
			IsVerified:   user.IsVerified,
			IsRejected:   user.IsRejected,
			CreatedAt:    int64(user.CreatedAt),
			NodeCluster:  pbCluster,
		})
	}

	return &pb.ListEnabledUsersResponse{Users: result}, nil
}

// FindEnabledUser 查询单个用户信息
func (this *UserService) FindEnabledUser(ctx context.Context, req *pb.FindEnabledUserRequest) (*pb.FindEnabledUserResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId
	}

	user, err := models.SharedUserDAO.FindEnabledUser(tx, req.UserId, nil)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &pb.FindEnabledUserResponse{User: nil}, nil
	}

	// 集群信息
	var pbCluster *pb.NodeCluster = nil
	if user.ClusterId > 0 {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(user.ClusterId))
		if err != nil {
			return nil, err
		}
		pbCluster = &pb.NodeCluster{
			Id:   int64(user.ClusterId),
			Name: clusterName,
		}
	}

	// 认证信息
	isIndividualIdentified, err := models.SharedUserIdentityDAO.CheckUserIdentityIsVerified(tx, req.UserId, userconfigs.UserIdentityOrgTypeIndividual)
	if err != nil {
		return nil, err
	}

	isEnterpriseIdentified, err := models.SharedUserIdentityDAO.CheckUserIdentityIsVerified(tx, req.UserId, userconfigs.UserIdentityOrgTypeEnterprise)
	if err != nil {
		return nil, err
	}

	// OTP认证
	var pbOtpAuth *pb.Login = nil
	{
		userAuth, err := models.SharedLoginDAO.FindEnabledLoginWithType(tx, 0, req.UserId, models.LoginTypeOTP)
		if err != nil {
			return nil, err
		}
		if userAuth != nil {
			pbOtpAuth = &pb.Login{
				Id:         int64(userAuth.Id),
				Type:       userAuth.Type,
				ParamsJSON: userAuth.Params,
				IsOn:       userAuth.IsOn,
			}
		}
	}

	return &pb.FindEnabledUserResponse{
		User: &pb.User{
			Id:                     int64(user.Id),
			Username:               user.Username,
			Fullname:               user.Fullname,
			Mobile:                 user.Mobile,
			Tel:                    user.Tel,
			Email:                  user.Email,
			VerifiedEmail:          user.VerifiedEmail,
			VerifiedMobile:         user.VerifiedMobile,
			Remark:                 user.Remark,
			IsOn:                   user.IsOn,
			CreatedAt:              int64(user.CreatedAt),
			RegisteredIP:           user.RegisteredIP,
			IsVerified:             user.IsVerified,
			IsRejected:             user.IsRejected,
			RejectReason:           user.RejectReason,
			NodeCluster:            pbCluster,
			IsIndividualIdentified: isIndividualIdentified,
			IsEnterpriseIdentified: isEnterpriseIdentified,
			BandwidthAlgo:          user.BandwidthAlgo,
			OtpLogin:               pbOtpAuth,
			Lang:                   user.Lang,
		},
	}, nil
}

// CheckUserUsername 检查用户名是否存在
func (this *UserService) CheckUserUsername(ctx context.Context, req *pb.CheckUserUsernameRequest) (*pb.CheckUserUsernameResponse, error) {
	userType, _, userId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	// 校验权限
	if userType == rpcutils.UserTypeUser && userId != req.UserId {
		return nil, this.PermissionError()
	}

	var tx = this.NullTx()

	b, err := models.SharedUserDAO.ExistUser(tx, req.UserId, req.Username)
	if err != nil {
		return nil, err
	}
	return &pb.CheckUserUsernameResponse{Exists: b}, nil
}

// LoginUser 登录
func (this *UserService) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	_, err := this.ValidateUserNode(ctx, false)
	if err != nil {
		return nil, err
	}

	if !teaconst.IsPlus {
		return &pb.LoginUserResponse{
			UserId:  0,
			IsOk:    false,
			Message: "你正在使用的系统版本为非商业版本或已过期，请管理员续费后才能登录",
		}, nil
	}

	if len(req.Username) == 0 || len(req.Password) == 0 {
		return &pb.LoginUserResponse{
			UserId:  0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	var tx = this.NullTx()

	// 邮箱登录
	var registerConfig *userconfigs.UserRegisterConfig
	if strings.Contains(req.Username, "@") {
		// 是否允许
		if registerConfig == nil {
			registerConfig, err = models.SharedSysSettingDAO.ReadUserRegisterConfig(tx)
			if err != nil {
				return nil, err
			}
		}
		if registerConfig != nil && registerConfig.EmailVerification.CanLogin {
			userId, err := models.SharedUserDAO.CheckUserEmailPassword(tx, req.Username, req.Password)
			if err != nil {
				return nil, err
			}
			if userId > 0 {
				return &pb.LoginUserResponse{
					UserId: userId,
					IsOk:   true,
				}, nil
			}
		}
	}

	// 手机号登录
	if utils.IsValidMobile(req.Username) {
		// 是否允许
		if registerConfig == nil {
			registerConfig, err = models.SharedSysSettingDAO.ReadUserRegisterConfig(tx)
			if err != nil {
				return nil, err
			}
		}
		if registerConfig != nil && registerConfig.MobileVerification.CanLogin {
			userId, err := models.SharedUserDAO.CheckUserMobilePassword(tx, req.Username, req.Password)
			if err != nil {
				return nil, err
			}
			if userId > 0 {
				return &pb.LoginUserResponse{
					UserId: userId,
					IsOk:   true,
				}, nil
			}
		}
	}

	// 用户名登录
	userId, err := models.SharedUserDAO.CheckUserPassword(tx, req.Username, req.Password)
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	if userId <= 0 {
		return &pb.LoginUserResponse{
			UserId:  0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	return &pb.LoginUserResponse{
		UserId: userId,
		IsOk:   true,
	}, nil
}

// UpdateUserInfo 修改用户基本信息
func (this *UserService) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId != req.UserId {
		return nil, this.PermissionError()
	}

	var tx = this.NullTx()

	err = models.SharedUserDAO.UpdateUserInfo(tx, req.UserId, req.Fullname, req.Mobile, req.Email)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateUserLogin 修改用户登录信息
func (this *UserService) UpdateUserLogin(ctx context.Context, req *pb.UpdateUserLoginRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId != req.UserId {
		return nil, this.PermissionError()
	}

	var tx = this.NullTx()

	err = models.SharedUserDAO.UpdateUserLogin(tx, req.UserId, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ComposeUserDashboard 取得用户Dashboard数据
func (this *UserService) ComposeUserDashboard(ctx context.Context, req *pb.ComposeUserDashboardRequest) (*pb.ComposeUserDashboardResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()

	bandwidthAglo, err := models.SharedUserDAO.FindUserBandwidthAlgoForView(tx, req.UserId, nil)
	if err != nil {
		return nil, err
	}

	// 网站数量
	countServers, err := models.SharedServerDAO.CountAllEnabledServersMatch(tx, 0, "", req.UserId, 0, configutils.BoolStateAll, []string{}, 0)
	if err != nil {
		return nil, err
	}

	// 时间相关
	var currentMonth = timeutil.Format("Ym")
	var currentDay = timeutil.Format("Ymd")

	// 本月总流量
	monthlyTrafficBytes, err := models.SharedUserBandwidthStatDAO.SumUserMonthly(tx, req.UserId, currentMonth)
	if err != nil {
		return nil, err
	}

	// 本月带宽峰值
	var monthlyPeekBandwidthBytes int64 = 0
	{
		stat, err := models.SharedUserBandwidthStatDAO.FindUserPeekBandwidthInMonth(tx, req.UserId, currentMonth, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
		if err != nil {
			return nil, err
		}
		if stat != nil {
			monthlyPeekBandwidthBytes = int64(stat.Bytes)
		}
	}

	// 本日带宽峰值
	var dailyPeekBandwidthBytes int64 = 0
	{
		stat, err := models.SharedUserBandwidthStatDAO.FindUserPeekBandwidthInDay(tx, req.UserId, currentDay, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
		if err != nil {
			return nil, err
		}
		if stat != nil {
			dailyPeekBandwidthBytes = int64(stat.Bytes)
		}
	}

	// 今日总流量
	dailyTrafficStat, err := models.SharedUserBandwidthStatDAO.SumUserDaily(tx, req.UserId, 0, currentDay)
	if err != nil {
		return nil, err
	}
	if dailyTrafficStat == nil {
		dailyTrafficStat = &models.UserBandwidthStat{}
	}

	// 近 30 日流量带宽趋势
	var dailyTrafficStats = []*pb.ComposeUserDashboardResponse_DailyTrafficStat{}
	var dailyPeekBandwidthStats = []*pb.ComposeUserDashboardResponse_DailyPeekBandwidthStat{}

	var dayFrom = ""
	var dayTo = ""
	for i := 30; i >= 0; i-- {
		var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -i))
		if len(dayFrom) == 0 {
			dayFrom = day
		}
		if len(dayTo) == 0 || day > dayTo {
			dayTo = day
		}

		// 流量
		trafficStat, err := models.SharedUserBandwidthStatDAO.SumUserDaily(tx, req.UserId, 0, day)
		if err != nil {
			return nil, err
		}
		if trafficStat == nil {
			trafficStat = &models.UserBandwidthStat{}
		}

		// 峰值带宽
		peekBandwidthBytesStat, err := models.SharedUserBandwidthStatDAO.FindUserPeekBandwidthInDay(tx, req.UserId, day, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
		if err != nil {
			return nil, err
		}
		var peekBandwidthBytes int64 = 0
		if peekBandwidthBytesStat != nil {
			peekBandwidthBytes = int64(peekBandwidthBytesStat.Bytes)
		}

		dailyTrafficStats = append(dailyTrafficStats, &pb.ComposeUserDashboardResponse_DailyTrafficStat{
			Day:                 day,
			Bytes:               int64(trafficStat.TotalBytes),
			CachedBytes:         int64(trafficStat.CachedBytes),
			AttackBytes:         int64(trafficStat.AttackBytes),
			CountRequests:       int64(trafficStat.CountRequests),
			CountCachedRequests: int64(trafficStat.CountCachedRequests),
			CountAttackRequests: int64(trafficStat.CountAttackRequests),
		})
		dailyPeekBandwidthStats = append(dailyPeekBandwidthStats, &pb.ComposeUserDashboardResponse_DailyPeekBandwidthStat{Day: day, Bytes: peekBandwidthBytes})
	}
	var result = &pb.ComposeUserDashboardResponse{
		CountServers:              countServers,
		MonthlyTrafficBytes:       monthlyTrafficBytes,
		MonthlyPeekBandwidthBytes: monthlyPeekBandwidthBytes,
		DailyTrafficBytes:         int64(dailyTrafficStat.TotalBytes),
		DailyPeekBandwidthBytes:   dailyPeekBandwidthBytes,
		DailyTrafficStats:         dailyTrafficStats,
		DailyPeekBandwidthStats:   dailyPeekBandwidthStats,
	}

	// 带宽百分位
	var bandwidthPercentile = systemconfigs.DefaultBandwidthPercentile
	userConfig, _ := models.SharedSysSettingDAO.ReadUserUIConfig(tx)
	if userConfig != nil && userConfig.TrafficStats.BandwidthPercentile > 0 {
		bandwidthPercentile = userConfig.TrafficStats.BandwidthPercentile
	}
	result.BandwidthPercentile = bandwidthPercentile
	stat, err := models.SharedUserBandwidthStatDAO.FindPercentileBetweenDays(tx, req.UserId, 0, dayFrom, dayTo, bandwidthPercentile, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
	if err != nil {
		return nil, err
	}
	if stat != nil {
		result.BandwidthPercentileBits = int64(stat.Bytes) * 8
	}

	return result, nil
}

// FindUserNodeClusterId 获取用户所在的集群ID
func (this *UserService) FindUserNodeClusterId(ctx context.Context, req *pb.FindUserNodeClusterIdRequest) (*pb.FindUserNodeClusterIdResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	clusterId, err := models.SharedUserDAO.FindUserClusterId(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.FindUserNodeClusterIdResponse{NodeClusterId: clusterId}, nil
}

// UpdateUserFeatures 设置用户能使用的功能
func (this *UserService) UpdateUserFeatures(ctx context.Context, req *pb.UpdateUserFeaturesRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	featuresJSON, err := json.Marshal(req.FeatureCodes)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedUserDAO.UpdateUserFeatures(tx, req.UserId, featuresJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateAllUsersFeatures 设置所有用户能使用的功能
func (this *UserService) UpdateAllUsersFeatures(ctx context.Context, req *pb.UpdateAllUsersFeaturesRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserDAO.UpdateUsersFeatures(tx, req.FeatureCodes, req.Overwrite)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllUserFeatureDefinitions 获取所有的功能定义
func (this *UserService) FindAllUserFeatureDefinitions(ctx context.Context, req *pb.FindAllUserFeatureDefinitionsRequest) (*pb.FindAllUserFeatureDefinitionsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var features = userconfigs.FindAllUserFeatures()
	var result = []*pb.UserFeature{}
	for _, feature := range features {
		result = append(result, feature.ToPB())
	}
	return &pb.FindAllUserFeatureDefinitionsResponse{Features: result}, nil
}

// ComposeUserGlobalBoard 组合全局的看板数据
func (this *UserService) ComposeUserGlobalBoard(ctx context.Context, req *pb.ComposeUserGlobalBoardRequest) (*pb.ComposeUserGlobalBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var result = &pb.ComposeUserGlobalBoardResponse{}
	var tx = this.NullTx()
	totalUsers, err := models.SharedUserDAO.CountAllEnabledUsers(tx, 0, "", false)
	if err != nil {
		return nil, err
	}
	result.TotalUsers = totalUsers

	// 等待审核的用户
	countVerifyingUsers, err := models.SharedUserDAO.CountAllVerifyingUsers(tx)
	if err != nil {
		return nil, err
	}
	result.CountVerifyingUsers = countVerifyingUsers

	countTodayUsers, err := models.SharedUserDAO.SumDailyUsers(tx, timeutil.Format("Ymd"), timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	result.CountTodayUsers = countTodayUsers

	{
		w := types.Int(timeutil.Format("w"))
		if w == 0 {
			w = 7
		}
		weekFrom := time.Now().AddDate(0, 0, -w+1)
		dayFrom := timeutil.Format("Ymd", weekFrom)
		dayTo := timeutil.Format("Ymd", weekFrom.AddDate(0, 0, 6))
		countWeeklyUsers, err := models.SharedUserDAO.SumDailyUsers(tx, dayFrom, dayTo)
		if err != nil {
			return nil, err
		}
		result.CountWeeklyUsers = countWeeklyUsers
	}

	// 用户节点数量
	countUserNodes, err := models.SharedUserNodeDAO.CountAllEnabledUserNodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountUserNodes = countUserNodes

	// 离线用户节点
	countOfflineUserNodes, err := models.SharedUserNodeDAO.CountAllEnabledAndOnOfflineNodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountOfflineUserNodes = countOfflineUserNodes

	// 用户增长趋势
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dayStats, err := models.SharedUserDAO.CountDailyUsers(tx, dayFrom, timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	result.DailyStats = dayStats

	// CPU、内存、负载
	cpuValues, err := models.SharedNodeValueDAO.ListValuesForUserNodes(tx, nodeconfigs.NodeValueItemCPU, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range cpuValues {
		result.CpuNodeValues = append(result.CpuNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	memoryValues, err := models.SharedNodeValueDAO.ListValuesForUserNodes(tx, nodeconfigs.NodeValueItemMemory, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range memoryValues {
		result.MemoryNodeValues = append(result.MemoryNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	loadValues, err := models.SharedNodeValueDAO.ListValuesForUserNodes(tx, nodeconfigs.NodeValueItemLoad, "load1m", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range loadValues {
		result.LoadNodeValues = append(result.LoadNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	// 流量排行
	var hourFrom = timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	var hourTo = timeutil.Format("YmdH")
	topUserStats, err := models.SharedServerDailyStatDAO.FindTopUserStats(tx, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range topUserStats {
		userName, err := models.SharedUserDAO.FindUserFullname(tx, int64(stat.UserId))
		if err != nil {
			return nil, err
		}
		result.TopTrafficStats = append(result.TopTrafficStats, &pb.ComposeUserGlobalBoardResponse_TrafficStat{
			UserId:        int64(stat.UserId),
			UserName:      userName,
			CountRequests: int64(stat.CountRequests),
			Bytes:         int64(stat.Bytes),
		})
	}

	return result, nil
}

// CheckUserOTPWithUsername 检查是否需要输入OTP
func (this *UserService) CheckUserOTPWithUsername(ctx context.Context, req *pb.CheckUserOTPWithUsernameRequest) (*pb.CheckUserOTPWithUsernameResponse, error) {
	_, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(req.Username) == 0 {
		return &pb.CheckUserOTPWithUsernameResponse{RequireOTP: false}, nil
	}

	var tx = this.NullTx()

	userId, err := models.SharedUserDAO.FindEnabledUserIdWithUsername(tx, req.Username)
	if err != nil {
		return nil, err
	}
	if userId <= 0 {
		return &pb.CheckUserOTPWithUsernameResponse{RequireOTP: false}, nil
	}

	otpIsOn, err := models.SharedLoginDAO.CheckLoginIsOn(tx, 0, userId, "otp")
	if err != nil {
		return nil, err
	}
	return &pb.CheckUserOTPWithUsernameResponse{RequireOTP: otpIsOn}, nil
}

// CheckUserServersState 检查用户服务可用状态
func (this *UserService) CheckUserServersState(ctx context.Context, req *pb.CheckUserServersStateRequest) (*pb.CheckUserServersStateResponse, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	isEnabled, err := models.SharedUserDAO.CheckUserServersEnabled(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.CheckUserServersStateResponse{
		IsEnabled: isEnabled,
	}, nil
}

// RenewUserServersState 更新用户服务可用状态
func (this *UserService) RenewUserServersState(ctx context.Context, req *pb.RenewUserServersStateRequest) (*pb.RenewUserServersStateResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	isEnabled, err := models.SharedUserDAO.RenewUserServersState(tx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.RenewUserServersStateResponse{
		IsEnabled: isEnabled,
	}, nil
}

// CheckUserEmail 检查邮箱是否被使用
func (this *UserService) CheckUserEmail(ctx context.Context, req *pb.CheckUserEmailRequest) (*pb.CheckUserEmailResponse, error) {
	userId, err := this.ValidateUserNode(ctx, false)
	if err != nil {
		return nil, err
	}

	if len(req.Email) == 0 {
		return nil, errors.New("'email' required")
	}

	var tx = this.NullTx()
	emailOwnerUserId, err := models.SharedUserDAO.FindUserIdWithVerifiedEmail(tx, req.Email)
	if err != nil {
		return nil, err
	}
	if emailOwnerUserId > 0 && userId != emailOwnerUserId {
		return &pb.CheckUserEmailResponse{Exists: true}, nil
	}
	return &pb.CheckUserEmailResponse{Exists: false}, nil
}

// CheckUserMobile 检查手机号码是否被使用
func (this *UserService) CheckUserMobile(ctx context.Context, req *pb.CheckUserMobileRequest) (*pb.CheckUserMobileResponse, error) {
	userId, err := this.ValidateUserNode(ctx, false)
	if err != nil {
		return nil, err
	}

	if len(req.Mobile) == 0 {
		return nil, errors.New("'mobile' required")
	}

	var tx = this.NullTx()
	mobileOwnerUserId, err := models.SharedUserDAO.FindUserIdWithVerifiedMobile(tx, req.Mobile)
	if err != nil {
		return nil, err
	}
	if mobileOwnerUserId > 0 && userId != mobileOwnerUserId {
		return &pb.CheckUserMobileResponse{Exists: true}, nil
	}
	return &pb.CheckUserMobileResponse{Exists: false}, nil
}

// FindUserVerifiedEmailWithUsername 根据用户名查询用户绑定的邮箱
func (this *UserService) FindUserVerifiedEmailWithUsername(ctx context.Context, req *pb.FindUserVerifiedEmailWithUsernameRequest) (*pb.FindUserVerifiedEmailWithUsernameResponse, error) {
	_, err := this.ValidateUserNode(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userId, err := models.SharedUserDAO.FindEnabledUserIdWithUsername(tx, req.Username)
	if err != nil {
		return nil, err
	}
	if userId <= 0 {
		return &pb.FindUserVerifiedEmailWithUsernameResponse{
			Email: "",
		}, nil
	}

	email, err := models.SharedUserDAO.FindUserVerifiedEmail(tx, userId)
	if err != nil {
		return nil, err
	}

	return &pb.FindUserVerifiedEmailWithUsernameResponse{
		Email: email,
	}, nil
}
