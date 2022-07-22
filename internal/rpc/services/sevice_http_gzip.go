package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
)

type HTTPGzipService struct {
	BaseService
}

// CreateHTTPGzip 创建Gzip配置
func (this *HTTPGzipService) CreateHTTPGzip(ctx context.Context, req *pb.CreateHTTPGzipRequest) (*pb.CreateHTTPGzipResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	minLengthJSON := []byte{}
	if req.MinLength != nil {
		minLengthJSON, err = (&shared.SizeCapacity{
			Count: req.MinLength.Count,
			Unit:  req.MinLength.Unit,
		}).AsJSON()
		if err != nil {
			return nil, err
		}
	}

	maxLengthJSON := []byte{}
	if req.MaxLength != nil {
		maxLengthJSON, err = (&shared.SizeCapacity{
			Count: req.MaxLength.Count,
			Unit:  req.MaxLength.Unit,
		}).AsJSON()
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	gzipId, err := models.SharedHTTPGzipDAO.CreateGzip(tx, int(req.Level), minLengthJSON, maxLengthJSON, req.CondsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPGzipResponse{HttpGzipId: gzipId}, nil
}

// FindEnabledHTTPGzipConfig 查找Gzip
func (this *HTTPGzipService) FindEnabledHTTPGzipConfig(ctx context.Context, req *pb.FindEnabledGzipConfigRequest) (*pb.FindEnabledGzipConfigResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	config, err := models.SharedHTTPGzipDAO.ComposeGzipConfig(tx, req.HttpGzipId)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledGzipConfigResponse{HttpGzipJSON: configData}, nil
}

// UpdateHTTPGzip 修改Gzip配置
func (this *HTTPGzipService) UpdateHTTPGzip(ctx context.Context, req *pb.UpdateHTTPGzipRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	minLengthJSON := []byte{}
	if req.MinLength != nil {
		minLengthJSON, err = (&shared.SizeCapacity{
			Count: req.MinLength.Count,
			Unit:  req.MinLength.Unit,
		}).AsJSON()
		if err != nil {
			return nil, err
		}
	}

	maxLengthJSON := []byte{}
	if req.MaxLength != nil {
		maxLengthJSON, err = (&shared.SizeCapacity{
			Count: req.MaxLength.Count,
			Unit:  req.MaxLength.Unit,
		}).AsJSON()
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

	err = models.SharedHTTPGzipDAO.UpdateGzip(tx, req.HttpGzipId, int(req.Level), minLengthJSON, maxLengthJSON, req.CondsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
