package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// 管理员认证相关服务
type LoginService struct {
	BaseService
}

// 查找认证
func (this *LoginService) FindEnabledLogin(ctx context.Context, req *pb.FindEnabledLoginRequest) (*pb.FindEnabledLoginResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	login, err := models.SharedLoginDAO.FindEnabledLoginWithAdminId(tx, req.AdminId, req.Type)
	if err != nil {
		return nil, err
	}
	if login == nil {
		return &pb.FindEnabledLoginResponse{Login: nil}, nil
	}
	return &pb.FindEnabledLoginResponse{Login: &pb.Login{
		Id:         int64(login.Id),
		Type:       login.Type,
		ParamsJSON: []byte(login.Params),
		IsOn:       login.IsOn == 1,
		AdminId:    int64(login.AdminId),
		UserId:     int64(login.UserId),
	}}, nil
}

// 修改认证
func (this *LoginService) UpdateLogin(ctx context.Context, req *pb.UpdateLoginRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if req.Login == nil {
		return nil, errors.New("'login' should not be nil")
	}

	tx := this.NullTx()

	if req.Login.IsOn {
		params := maps.Map{}
		if len(req.Login.ParamsJSON) > 0 {
			err = json.Unmarshal(req.Login.ParamsJSON, &params)
			if err != nil {
				return nil, err
			}
		}
		err = models.SharedLoginDAO.UpdateLogin(tx, req.Login.AdminId, req.Login.Type, params, req.Login.IsOn)
		if err != nil {
			return nil, err
		}
	} else {
		err = models.SharedLoginDAO.DisableLoginWithAdminId(tx, req.Login.AdminId, req.Login.Type)
		if err != nil {
			return nil, err
		}
	}
	return this.Success()
}
