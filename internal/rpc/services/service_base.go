package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/encrypt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"google.golang.org/grpc/metadata"
	"time"
)

type BaseService struct {
}

// ValidateAdmin 校验管理员
func (this *BaseService) ValidateAdmin(ctx context.Context, reqAdminId int64) (adminId int64, err error) {
	_, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return
	}
	if reqAdminId > 0 && reqUserId != reqAdminId {
		return 0, this.PermissionError()
	}
	return reqUserId, nil
}

// ValidateAdminAndUser 校验管理员和用户
func (this *BaseService) ValidateAdminAndUser(ctx context.Context, requireAdminId int64, requireUserId int64) (adminId int64, userId int64, err error) {
	reqUserType, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return
	}

	adminId = int64(0)
	userId = int64(0)
	switch reqUserType {
	case rpcutils.UserTypeAdmin:
		adminId = reqUserId
		if adminId < 0 { // 允许AdminId = 0
			err = errors.New("invalid 'adminId'")
			return
		}
		if requireAdminId > 0 && adminId != requireAdminId {
			err = this.PermissionError()
			return
		}
	case rpcutils.UserTypeUser:
		userId = reqUserId
		if requireUserId >= 0 && userId <= 0 {
			err = errors.New("invalid 'userId'")
			return
		}
		if requireUserId > 0 && userId != requireUserId {
			err = this.PermissionError()
			return
		}
	default:
		err = errors.New("invalid user type")
	}

	return
}

// ValidateNode 校验边缘节点
func (this *BaseService) ValidateNode(ctx context.Context) (nodeId int64, err error) {
	_, nodeId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	return
}

// ValidateUser 校验用户节点
func (this *BaseService) ValidateUser(ctx context.Context) (userId int64, err error) {
	_, userId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeUser)
	return
}

// ValidateMonitor 校验监控节点
func (this *BaseService) ValidateMonitor(ctx context.Context) (nodeId int64, err error) {
	_, nodeId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeMonitor)
	return
}

// ValidateAuthority 校验认证节点
func (this *BaseService) ValidateAuthority(ctx context.Context) (nodeId int64, err error) {
	_, nodeId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAuthority)
	return
}

// ValidateNodeId 获取节点ID
func (this *BaseService) ValidateNodeId(ctx context.Context, roles ...rpcutils.UserType) (role rpcutils.UserType, nodeIntId int64, err error) {
	if ctx == nil {
		err = errors.New("context should not be nil")
		role = rpcutils.UserTypeNone
		return
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return rpcutils.UserTypeNone, 0, errors.New("context: need 'nodeId'")
	}
	nodeIds := md.Get("nodeid")
	if len(nodeIds) == 0 || len(nodeIds[0]) == 0 {
		return rpcutils.UserTypeNone, 0, errors.New("context: need 'nodeId'")
	}
	nodeId := nodeIds[0]

	// 获取角色Node信息
	// TODO 缓存节点ID相关信息
	apiToken, err := models.SharedApiTokenDAO.FindEnabledTokenWithNode(nil, nodeId)
	if err != nil {
		return rpcutils.UserTypeNone, 0, err
	}
	if apiToken == nil {
		return rpcutils.UserTypeNone, 0, errors.New("context: can not find api token for node '" + nodeId + "'")
	}
	if !lists.ContainsString(roles, apiToken.Role) {
		return rpcutils.UserTypeNone, 0, errors.New("context: unsupported role '" + apiToken.Role + "'")
	}

	tokens := md.Get("token")
	if len(tokens) == 0 || len(tokens[0]) == 0 {
		return rpcutils.UserTypeNone, 0, errors.New("context: need 'token'")
	}
	token := tokens[0]

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return rpcutils.UserTypeNone, 0, err
	}

	method, err := encrypt.NewMethodInstance(teaconst.EncryptMethod, apiToken.Secret, nodeId)
	if err != nil {
		utils.PrintError(err)
		return rpcutils.UserTypeNone, 0, err
	}
	data, err = method.Decrypt(data)
	if err != nil {
		return rpcutils.UserTypeNone, 0, err
	}
	if len(data) == 0 {
		return rpcutils.UserTypeNone, 0, errors.New("invalid token")
	}

	m := maps.Map{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return rpcutils.UserTypeNone, 0, errors.New("decode token error: " + err.Error())
	}

	timestamp := m.GetInt64("timestamp")
	if time.Now().Unix()-timestamp > 600 {
		// 请求超过10分钟认为超时
		return rpcutils.UserTypeNone, 0, errors.New("authenticate timeout")
	}

	switch apiToken.Role {
	case rpcutils.UserTypeNode:
		nodeIntId, err = models.SharedNodeDAO.FindEnabledNodeIdWithUniqueId(nil, nodeId)
		if err != nil {
			return rpcutils.UserTypeNode, 0, errors.New("context: " + err.Error())
		}
		if nodeIntId <= 0 {
			return rpcutils.UserTypeNode, 0, errors.New("context: not found node with id '" + nodeId + "'")
		}
	case rpcutils.UserTypeCluster:
		nodeIntId, err = models.SharedNodeClusterDAO.FindEnabledClusterIdWithUniqueId(nil, nodeId)
		if err != nil {
			return rpcutils.UserTypeCluster, 0, errors.New("context: " + err.Error())
		}
		if nodeIntId <= 0 {
			return rpcutils.UserTypeCluster, 0, errors.New("context: not found cluster with id '" + nodeId + "'")
		}
	case rpcutils.UserTypeUser:
		nodeIntId, err = models.SharedUserNodeDAO.FindEnabledUserNodeIdWithUniqueId(nil, nodeId)
	case rpcutils.UserTypeAdmin:
		nodeIntId = 0
	case rpcutils.UserTypeMonitor:
		nodeIntId, err = models.SharedMonitorNodeDAO.FindEnabledMonitorNodeIdWithUniqueId(nil, nodeId)
	default:
		err = errors.New("unsupported user role '" + apiToken.Role + "'")
	}

	return
}

// Success 返回成功
func (this *BaseService) Success() (*pb.RPCSuccess, error) {
	return &pb.RPCSuccess{}, nil
}

// SuccessCount 返回数字
func (this *BaseService) SuccessCount(count int64) (*pb.RPCCountResponse, error) {
	return &pb.RPCCountResponse{Count: count}, nil
}

// PermissionError 返回权限错误
func (this *BaseService) PermissionError() error {
	return errors.New("Permission Denied")
}

// NullTx 空的数据库事务
func (this *BaseService) NullTx() *dbs.Tx {
	return nil
}

// RunTx 在当前数据中执行一个事务
func (this *BaseService) RunTx(callback func(tx *dbs.Tx) error) error {
	db, err := dbs.Default()
	if err != nil {
		return err
	}
	return db.RunTx(callback)
}
