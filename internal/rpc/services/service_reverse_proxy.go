package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/types"
)

type ReverseProxyService struct {
	BaseService
}

// CreateReverseProxy 创建反向代理
func (this *ReverseProxyService) CreateReverseProxy(ctx context.Context, req *pb.CreateReverseProxyRequest) (*pb.CreateReverseProxyResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验源站
	}

	var tx = this.NullTx()

	reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, userId, req.SchedulingJSON, req.PrimaryOriginsJSON, req.BackupOriginsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateReverseProxyResponse{ReverseProxyId: reverseProxyId}, nil
}

// FindEnabledReverseProxy 查找反向代理
func (this *ReverseProxyService) FindEnabledReverseProxy(ctx context.Context, req *pb.FindEnabledReverseProxyRequest) (*pb.FindEnabledReverseProxyResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	reverseProxy, err := models.SharedReverseProxyDAO.FindEnabledReverseProxy(tx, req.ReverseProxyId)
	if err != nil {
		return nil, err
	}
	if reverseProxy == nil {
		return &pb.FindEnabledReverseProxyResponse{ReverseProxy: nil}, nil
	}

	result := &pb.ReverseProxy{
		Id:                 int64(reverseProxy.Id),
		SchedulingJSON:     reverseProxy.Scheduling,
		PrimaryOriginsJSON: reverseProxy.PrimaryOrigins,
		BackupOriginsJSON:  reverseProxy.BackupOrigins,
	}
	return &pb.FindEnabledReverseProxyResponse{ReverseProxy: result}, nil
}

// FindEnabledReverseProxyConfig 查找反向代理配置
func (this *ReverseProxyService) FindEnabledReverseProxyConfig(ctx context.Context, req *pb.FindEnabledReverseProxyConfigRequest) (*pb.FindEnabledReverseProxyConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	config, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, req.ReverseProxyId, nil, nil)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledReverseProxyConfigResponse{ReverseProxyJSON: configData}, nil
}

// UpdateReverseProxyScheduling 修改反向代理调度算法
func (this *ReverseProxyService) UpdateReverseProxyScheduling(ctx context.Context, req *pb.UpdateReverseProxySchedulingRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	err = models.SharedReverseProxyDAO.UpdateReverseProxyScheduling(tx, req.ReverseProxyId, req.SchedulingJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateReverseProxyPrimaryOrigins 修改主要源站信息
func (this *ReverseProxyService) UpdateReverseProxyPrimaryOrigins(ctx context.Context, req *pb.UpdateReverseProxyPrimaryOriginsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	err = models.SharedReverseProxyDAO.UpdateReverseProxyPrimaryOrigins(tx, req.ReverseProxyId, req.OriginsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateReverseProxyBackupOrigins 修改备用源站信息
func (this *ReverseProxyService) UpdateReverseProxyBackupOrigins(ctx context.Context, req *pb.UpdateReverseProxyBackupOriginsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	err = models.SharedReverseProxyDAO.UpdateReverseProxyBackupOrigins(tx, req.ReverseProxyId, req.OriginsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateReverseProxy 修改是否启用
func (this *ReverseProxyService) UpdateReverseProxy(ctx context.Context, req *pb.UpdateReverseProxyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedReverseProxyDAO.CheckUserReverseProxy(nil, userId, req.ReverseProxyId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	// 校验参数
	var connTimeout = &shared.TimeDuration{}
	if len(req.ConnTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ConnTimeoutJSON, connTimeout)
		if err != nil {
			return nil, err
		}
	}

	var readTimeout = &shared.TimeDuration{}
	if len(req.ReadTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ReadTimeoutJSON, readTimeout)
		if err != nil {
			return nil, err
		}
	}

	var idleTimeout = &shared.TimeDuration{}
	if len(req.IdleTimeoutJSON) > 0 {
		err = json.Unmarshal(req.IdleTimeoutJSON, idleTimeout)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedReverseProxyDAO.UpdateReverseProxy(tx, req.ReverseProxyId, types.Int8(req.RequestHostType), req.RequestHost, req.RequestHostExcludingPort, req.RequestURI, req.StripPrefix, req.AutoFlush, req.AddHeaders, connTimeout, readTimeout, idleTimeout, req.MaxConns, req.MaxIdleConns, req.ProxyProtocolJSON, req.FollowRedirects)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
