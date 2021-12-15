package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
)

type HTTPHeaderService struct {
	BaseService
}

// CreateHTTPHeader 创建Header
func (this *HTTPHeaderService) CreateHTTPHeader(ctx context.Context, req *pb.CreateHTTPHeaderRequest) (*pb.CreateHTTPHeaderResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户权限
	}

	tx := this.NullTx()

	// status
	var newStatus = []int{}
	for _, status := range req.Status {
		newStatus = append(newStatus, int(status))
	}

	// replace values
	var replaceValues = []*shared.HTTPHeaderReplaceValue{}
	if len(req.ReplaceValuesJSON) > 0 {
		err = json.Unmarshal(req.ReplaceValuesJSON, &replaceValues)
		if err != nil {
			return nil, errors.New("decode replace values failed: " + err.Error() + ", json: " + string(req.ReplaceValuesJSON))
		}
	}

	headerId, err := models.SharedHTTPHeaderDAO.CreateHeader(tx, userId, req.Name, req.Value, newStatus, req.DisableRedirect, req.ShouldAppend, req.ShouldReplace, replaceValues, req.Methods, req.Domains)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPHeaderResponse{HeaderId: headerId}, nil
}

// UpdateHTTPHeader 修改Header
func (this *HTTPHeaderService) UpdateHTTPHeader(ctx context.Context, req *pb.UpdateHTTPHeaderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户权限
	}

	tx := this.NullTx()

	// status
	var newStatus = []int{}
	for _, status := range req.Status {
		newStatus = append(newStatus, int(status))
	}

	// replace values
	var replaceValues = []*shared.HTTPHeaderReplaceValue{}
	if len(req.ReplaceValuesJSON) > 0 {
		err = json.Unmarshal(req.ReplaceValuesJSON, &replaceValues)
		if err != nil {
			return nil, errors.New("decode replace values failed: " + err.Error())
		}
	}

	err = models.SharedHTTPHeaderDAO.UpdateHeader(tx, req.HeaderId, req.Name, req.Value, newStatus, req.DisableRedirect, req.ShouldAppend, req.ShouldReplace, replaceValues, req.Methods, req.Domains)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledHTTPHeaderConfig 查找配置
func (this *HTTPHeaderService) FindEnabledHTTPHeaderConfig(ctx context.Context, req *pb.FindEnabledHTTPHeaderConfigRequest) (*pb.FindEnabledHTTPHeaderConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户权限
	}

	tx := this.NullTx()

	config, err := models.SharedHTTPHeaderDAO.ComposeHeaderConfig(tx, req.HeaderId)
	if err != nil {
		return nil, err
	}
	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledHTTPHeaderConfigResponse{HeaderJSON: configData}, nil
}
