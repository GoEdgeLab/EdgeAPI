package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type SysSettingService struct {
	BaseService
}

// UpdateSysSetting 更改配置
func (this *SysSettingService) UpdateSysSetting(ctx context.Context, req *pb.UpdateSysSettingRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	// 不要允许用户修改
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedSysSettingDAO.UpdateSetting(tx, req.Code, req.ValueJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ReadSysSetting 读取配置
func (this *SysSettingService) ReadSysSetting(ctx context.Context, req *pb.ReadSysSettingRequest) (*pb.ReadSysSettingResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(tx, req.Code)
	if err != nil {
		return nil, err
	}

	return &pb.ReadSysSettingResponse{ValueJSON: valueJSON}, nil
}
