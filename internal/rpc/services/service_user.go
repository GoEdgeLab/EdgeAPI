package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// 用户相关服务
type UserService struct {
	BaseService
}

// 创建用户
func (this *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	userId, err := models.SharedUserDAO.CreateUser(req.Username, req.Password, req.Fullname, req.Mobile, req.Tel, req.Email, req.Remark, req.Source, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{UserId: userId}, nil
}

// 修改用户
func (this *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	err = models.SharedUserDAO.UpdateUser(req.UserId, req.Username, req.Password, req.Fullname, req.Mobile, req.Tel, req.Email, req.Remark, req.IsOn, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除用户
func (this *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	_, err = models.SharedUserDAO.DisableUser(req.UserId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 计算用户数量
func (this *UserService) CountAllEnabledUsers(ctx context.Context, req *pb.CountAllEnabledUsersRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedUserDAO.CountAllEnabledUsers(req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 列出单页用户
func (this *UserService) ListEnabledUsers(ctx context.Context, req *pb.ListEnabledUsersRequest) (*pb.ListEnabledUsersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	users, err := models.SharedUserDAO.ListEnabledUsers(req.Keyword)
	if err != nil {
		return nil, err
	}

	result := []*pb.User{}
	for _, user := range users {
		// 集群信息
		var pbCluster *pb.NodeCluster = nil
		if user.ClusterId > 0 {
			clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(user.ClusterId))
			if err != nil {
				return nil, err
			}
			pbCluster = &pb.NodeCluster{
				Id:   int64(user.ClusterId),
				Name: clusterName,
			}
		}

		result = append(result, &pb.User{
			Id:        int64(user.Id),
			Username:  user.Username,
			Fullname:  user.Fullname,
			Mobile:    user.Mobile,
			Tel:       user.Tel,
			Email:     user.Email,
			Remark:    user.Remark,
			IsOn:      user.IsOn == 1,
			CreatedAt: int64(user.CreatedAt),
			Cluster:   pbCluster,
		})
	}

	return &pb.ListEnabledUsersResponse{Users: result}, nil
}

// 查询单个用户信息
func (this *UserService) FindEnabledUser(ctx context.Context, req *pb.FindEnabledUserRequest) (*pb.FindEnabledUserResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}

	user, err := models.SharedUserDAO.FindEnabledUser(req.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &pb.FindEnabledUserResponse{User: nil}, nil
	}

	// 集群信息
	var pbCluster *pb.NodeCluster = nil
	if user.ClusterId > 0 {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(user.ClusterId))
		if err != nil {
			return nil, err
		}
		pbCluster = &pb.NodeCluster{
			Id:   int64(user.ClusterId),
			Name: clusterName,
		}
	}

	return &pb.FindEnabledUserResponse{User: &pb.User{
		Id:        int64(user.Id),
		Username:  user.Username,
		Fullname:  user.Fullname,
		Mobile:    user.Mobile,
		Tel:       user.Tel,
		Email:     user.Email,
		Remark:    user.Remark,
		IsOn:      user.IsOn == 1,
		CreatedAt: int64(user.CreatedAt),
		Cluster:   pbCluster,
	}}, nil
}

// 检查用户名是否存在
func (this *UserService) CheckUserUsername(ctx context.Context, req *pb.CheckUserUsernameRequest) (*pb.CheckUserUsernameResponse, error) {
	userType, userId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	// 校验权限
	if userType == rpcutils.UserTypeUser && userId != req.UserId {
		return nil, this.PermissionError()
	}

	b, err := models.SharedUserDAO.ExistUser(req.UserId, req.Username)
	if err != nil {
		return nil, err
	}
	return &pb.CheckUserUsernameResponse{Exists: b}, nil
}

// 登录
func (this *UserService) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.Username) == 0 || len(req.Password) == 0 {
		return &pb.LoginUserResponse{
			UserId:  0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	userId, err := models.SharedUserDAO.CheckUserPassword(req.Username, req.Password)
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

// 修改用户基本信息
func (this *UserService) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUser(ctx)
	if err != nil {
		return nil, err
	}

	if userId != req.UserId {
		return nil, this.PermissionError()
	}

	err = models.SharedUserDAO.UpdateUserInfo(req.UserId, req.Fullname)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 修改用户登录信息
func (this *UserService) UpdateUserLogin(ctx context.Context, req *pb.UpdateUserLoginRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUser(ctx)
	if err != nil {
		return nil, err
	}

	if userId != req.UserId {
		return nil, this.PermissionError()
	}

	err = models.SharedUserDAO.UpdateUserLogin(req.UserId, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 取得用户Dashboard数据
func (this *UserService) ComposeUserDashboard(ctx context.Context, req *pb.ComposeUserDashboardRequest) (*pb.ComposeUserDashboardResponse, error) {
	userId, err := this.ValidateUser(ctx)
	if err != nil {
		return nil, err
	}

	if userId != req.UserId {
		return nil, this.PermissionError()
	}

	// 网站数量
	countServers, err := models.SharedServerDAO.CountAllEnabledServersMatch(0, "", req.UserId)
	if err != nil {
		return nil, err
	}

	// 本月总流量
	month := timeutil.Format("Ym")
	monthlyTrafficBytes, err := models.SharedServerDailyStatDAO.SumUserMonthly(req.UserId, 0, month)
	if err != nil {
		return nil, err
	}

	// 本月带宽峰值
	monthlyPeekTrafficBytes, err := models.SharedServerDailyStatDAO.SumUserMonthly(req.UserId, 0, month)
	if err != nil {
		return nil, err
	}

	// 今日总流量
	day := timeutil.Format("Ymd")
	dailyTrafficBytes, err := models.SharedServerDailyStatDAO.SumUserDaily(req.UserId, 0, day)
	if err != nil {
		return nil, err
	}

	// 今日带宽峰值
	dailyPeekTrafficBytes, err := models.SharedServerDailyStatDAO.SumUserDailyPeek(req.UserId, 0, day)
	if err != nil {
		return nil, err
	}

	// 近 15 日流量带宽趋势
	dailyTrafficStats := []*pb.ComposeUserDashboardResponse_DailyStat{}
	dailyPeekTrafficStats := []*pb.ComposeUserDashboardResponse_DailyStat{}

	for i := 14; i >= 0; i-- {
		day := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -i))

		dailyTrafficBytes, err := models.SharedServerDailyStatDAO.SumUserDaily(req.UserId, 0, day)
		if err != nil {
			return nil, err
		}

		dailyPeekTrafficBytes, err := models.SharedServerDailyStatDAO.SumUserDailyPeek(req.UserId, 0, day)
		if err != nil {
			return nil, err
		}

		dailyTrafficStats = append(dailyTrafficStats, &pb.ComposeUserDashboardResponse_DailyStat{Day: day, Count: dailyTrafficBytes})
		dailyPeekTrafficStats = append(dailyPeekTrafficStats, &pb.ComposeUserDashboardResponse_DailyStat{Day: day, Count: dailyPeekTrafficBytes})
	}

	return &pb.ComposeUserDashboardResponse{
		CountServers:            countServers,
		MonthlyTrafficBytes:     monthlyTrafficBytes,
		MonthlyPeekTrafficBytes: monthlyPeekTrafficBytes,
		DailyTrafficBytes:       dailyTrafficBytes,
		DailyPeekTrafficBytes:   dailyPeekTrafficBytes,
		DailyTrafficStats:       dailyTrafficStats,
		DailyPeekTrafficStats:   dailyPeekTrafficStats,
	}, nil
}
