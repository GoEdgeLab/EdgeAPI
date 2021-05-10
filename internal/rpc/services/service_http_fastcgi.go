// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// HTTPFastcgiService HTTP Fastcgi服务
type HTTPFastcgiService struct {
	BaseService
}

// CreateHTTPFastcgi 创建Fastcgi
func (this *HTTPFastcgiService) CreateHTTPFastcgi(ctx context.Context, req *pb.CreateHTTPFastcgiRequest) (*pb.CreateHTTPFastcgiResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	fastcgiId, err := models.SharedHTTPFastcgiDAO.CreateFastcgi(tx, adminId, userId, req.IsOn, req.Address, req.ParamsJSON, req.ReadTimeoutJSON, req.ConnTimeoutJSON, req.PoolSize, req.PathInfoPattern)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPFastcgiResponse{HttpFastcgiId: fastcgiId}, nil
}

// UpdateHTTPFastcgi 修改Fastcgi
func (this *HTTPFastcgiService) UpdateHTTPFastcgi(ctx context.Context, req *pb.UpdateHTTPFastcgiRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedHTTPFastcgiDAO.CheckUserFastcgi(tx, userId, req.HttpFastcgiId)
		if err != nil {
			return nil, err
		}
	}
	err = models.SharedHTTPFastcgiDAO.UpdateFastcgi(tx, req.HttpFastcgiId, req.IsOn, req.Address, req.ParamsJSON, req.ReadTimeoutJSON, req.ConnTimeoutJSON, req.PoolSize, req.PathInfoPattern)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledHTTPFastcgi 获取Fastcgi详情
func (this *HTTPFastcgiService) FindEnabledHTTPFastcgi(ctx context.Context, req *pb.FindEnabledHTTPFastcgiRequest) (*pb.FindEnabledHTTPFastcgiResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedHTTPFastcgiDAO.CheckUserFastcgi(tx, userId, req.HttpFastcgiId)
		if err != nil {
			return nil, err
		}
	}

	fastcgi, err := models.SharedHTTPFastcgiDAO.FindEnabledHTTPFastcgi(tx, req.HttpFastcgiId)
	if err != nil {
		return nil, err
	}
	if fastcgi == nil {
		return &pb.FindEnabledHTTPFastcgiResponse{HttpFastcgi: nil}, nil
	}
	return &pb.FindEnabledHTTPFastcgiResponse{HttpFastcgi: &pb.HTTPFastcgi{
		Id:              int64(fastcgi.Id),
		IsOn:            fastcgi.IsOn == 1,
		Address:         fastcgi.Address,
		ParamsJSON:      []byte(fastcgi.Params),
		ReadTimeoutJSON: []byte(fastcgi.ReadTimeout),
		ConnTimeoutJSON: []byte(fastcgi.ConnTimeout),
		PoolSize:        types.Int32(fastcgi.PoolSize),
		PathInfoPattern: fastcgi.PathInfoPattern,
	}}, nil
}

// FindEnabledHTTPFastcgiConfig 获取Fastcgi配置
func (this *HTTPFastcgiService) FindEnabledHTTPFastcgiConfig(ctx context.Context, req *pb.FindEnabledHTTPFastcgiConfigRequest) (*pb.FindEnabledHTTPFastcgiConfigResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedHTTPFastcgiDAO.CheckUserFastcgi(tx, userId, req.HttpFastcgiId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPFastcgiDAO.ComposeFastcgiConfig(tx, req.HttpFastcgiId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledHTTPFastcgiConfigResponse{HttpFastcgiJSON: configJSON}, nil
}
