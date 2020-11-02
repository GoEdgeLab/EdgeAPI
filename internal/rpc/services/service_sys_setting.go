package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type SysSettingService struct {
}

// 更改配置
func (this *SysSettingService) UpdateSysSetting(ctx context.Context, req *pb.UpdateSysSettingRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedSysSettingDAO.UpdateSetting(req.Code, req.ValueJSON)
	if err != nil {
		return nil, err
	}
	
	return rpcutils.RPCUpdateSuccess()
}

// 读取配置
func (this *SysSettingService) ReadSysSetting(ctx context.Context, req *pb.ReadSysSettingRequest) (*pb.ReadSysSettingResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(req.Code)
	if err != nil {
		return nil, err
	}

	return &pb.ReadSysSettingResponse{ValueJSON: valueJSON}, nil
}
