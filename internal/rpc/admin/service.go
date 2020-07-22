package admin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/encrypt"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/maps"
	"google.golang.org/grpc/metadata"
	"time"
)

type Service struct {
	debug bool
}

func (this *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	_, err := this.validateRequest(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.Username) == 0 || len(req.Password) == 0 {
		return &LoginResponse{
			AdminId: 0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	adminId, err := models.SharedAdminDAO.CheckAdminPassword(req.Username, req.Password)
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	if adminId <= 0 {
		return &LoginResponse{
			AdminId: 0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	return &LoginResponse{
		AdminId: int64(adminId),
		IsOk:    true,
	}, nil
}

func (this *Service) CreateLog(ctx context.Context, req *CreateLogRequest) (*CreateLogResponse, error) {
	adminId, err := this.validateAdminRequest(ctx)
	if err != nil {
		return nil, err
	}
	err = models.SharedLogDAO.CreateAdminLog(adminId, req.Level, req.Description, req.Action, req.Ip)
	return &CreateLogResponse{
		IsOk: err != nil,
	}, err
}

func (this *Service) CheckAdminExists(ctx context.Context, req *CheckAdminExistsRequest) (*CheckAdminExistsResponse, error) {
	_, err := this.validateRequest(ctx)
	if err != nil {
		return nil, err
	}

	if req.AdminId <= 0 {
		return &CheckAdminExistsResponse{
			IsOk: false,
		}, nil
	}

	ok, err := models.SharedAdminDAO.ExistEnabledAdmin(int(req.AdminId))
	if err != nil {
		return nil, err
	}

	return &CheckAdminExistsResponse{
		IsOk: ok,
	}, nil
}

func (this *Service) FindAdminFullname(ctx context.Context, req *FindAdminNameRequest) (*FindAdminNameResponse, error) {
	_, err := this.validateRequest(ctx)
	if err != nil {
		return nil, err
	}

	fullname, err := models.SharedAdminDAO.FindAdminFullname(int(req.AdminId))
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	return &FindAdminNameResponse{
		Fullname: fullname,
	}, nil
}

func (this *Service) validateRequest(ctx context.Context) (adminId int, err error) {
	var md metadata.MD
	var ok bool
	if this.debug {
		md, ok = metadata.FromOutgoingContext(ctx)
	} else {
		md, ok = metadata.FromIncomingContext(ctx)
	}
	if !ok {
		return 0, errors.New("context: need 'nodeId'")
	}
	nodeIds := md.Get("nodeid")
	if len(nodeIds) == 0 || len(nodeIds[0]) == 0 {
		return 0, errors.New("context: need 'nodeId'")
	}
	nodeId := nodeIds[0]

	// 获取Node信息
	apiToken, err := models.SharedApiTokenDAO.FindEnabledTokenWithNode(nodeId)
	if err != nil {
		utils.PrintError(err)
		return 0, err
	}
	if apiToken == nil {
		return 0, errors.New("can not find token from node id: " + err.Error())
	}

	tokens := md.Get("token")
	if len(tokens) == 0 || len(tokens[0]) == 0 {
		return 0, errors.New("context: need 'token'")
	}
	token := tokens[0]

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}

	method, err := encrypt.NewMethodInstance(teaconst.EncryptMethod, apiToken.Secret, nodeId)
	if err != nil {
		utils.PrintError(err)
		return 0, err
	}
	data, err = method.Decrypt(data)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, errors.New("invalid token")
	}

	m := maps.Map{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return 0, errors.New("decode token error: " + err.Error())
	}

	timestamp := m.GetInt64("timestamp")
	if time.Now().Unix()-timestamp > 600 {
		// 请求超过10分钟认为超时
		return 0, errors.New("authenticate timeout")
	}

	adminId = m.GetInt("adminId")
	return
}

func (this *Service) validateAdminRequest(ctx context.Context) (adminId int, err error) {
	adminId, err = this.validateRequest(ctx)
	if err != nil {
		return 0, err
	}
	if adminId <= 0 {
		return 0, errors.New("invalid admin id")
	}
	return
}
