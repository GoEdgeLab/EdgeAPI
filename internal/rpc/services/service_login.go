package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// LoginService 管理员认证相关服务
type LoginService struct {
	BaseService
}

// FindEnabledLogin 查找认证
func (this *LoginService) FindEnabledLogin(ctx context.Context, req *pb.FindEnabledLoginRequest) (*pb.FindEnabledLoginResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()

	login, err := models.SharedLoginDAO.FindEnabledLoginWithType(tx, req.AdminId, req.UserId, req.Type)
	if err != nil {
		return nil, err
	}
	if login == nil {
		return &pb.FindEnabledLoginResponse{Login: nil}, nil
	}
	return &pb.FindEnabledLoginResponse{Login: &pb.Login{
		Id:         int64(login.Id),
		Type:       login.Type,
		ParamsJSON: login.Params,
		IsOn:       login.IsOn,
		AdminId:    int64(login.AdminId),
		UserId:     int64(login.UserId),
	}}, nil
}

// UpdateLogin 修改认证
func (this *LoginService) UpdateLogin(ctx context.Context, req *pb.UpdateLoginRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if req.Login == nil {
		return nil, errors.New("'login' should not be nil")
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.Login.UserId = userId
	}

	if req.Login.IsOn {
		var params = maps.Map{}
		if len(req.Login.ParamsJSON) > 0 {
			err = json.Unmarshal(req.Login.ParamsJSON, &params)
			if err != nil {
				return nil, err
			}
		}
		err = models.SharedLoginDAO.UpdateLogin(tx, req.Login.AdminId, req.Login.UserId, req.Login.Type, params, req.Login.IsOn)
		if err != nil {
			return nil, err
		}
	} else {
		err = models.SharedLoginDAO.DisableLoginWithType(tx, req.Login.AdminId, req.Login.UserId, req.Login.Type)
		if err != nil {
			return nil, err
		}
	}
	return this.Success()
}
