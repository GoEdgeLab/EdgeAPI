package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type SysSettingService struct {
	BaseService
}

// 更改配置
func (this *SysSettingService) UpdateSysSetting(ctx context.Context, req *pb.UpdateSysSettingRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedSysSettingDAO.UpdateSetting(tx, req.Code, req.ValueJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 读取配置
func (this *SysSettingService) ReadSysSetting(ctx context.Context, req *pb.ReadSysSettingRequest) (*pb.ReadSysSettingResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(tx, req.Code)
	if err != nil {
		return nil, err
	}

	return &pb.ReadSysSettingResponse{ValueJSON: valueJSON}, nil
}
