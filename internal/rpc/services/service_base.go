package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/authority"
	"github.com/TeaOSLab/EdgeAPI/internal/encrypt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type BaseService struct {
}

// ValidateAdmin 校验管理员
func (this *BaseService) ValidateAdmin(ctx context.Context) (adminId int64, err error) {
	_, _, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return
	}
	return reqUserId, nil
}

// ValidateAdminAndUser 校验管理员和用户
func (this *BaseService) ValidateAdminAndUser(ctx context.Context, canRest bool) (adminId int64, userId int64, err error) {
	reqUserType, _, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
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
	case rpcutils.UserTypeUser:
		userId = reqUserId
		if userId < 0 { // 允许等于0
			err = errors.New("invalid 'userId'")
			return
		}
	default:
		err = errors.New("invalid user type")
	}

	if err != nil {
		return
	}

	if userId > 0 && !canRest && rpcutils.IsRest(ctx) {
		err = errors.New("can not be called by rest")
		return
	}

	return
}

// ValidateNode 校验边缘节点
func (this *BaseService) ValidateNode(ctx context.Context) (nodeId int64, err error) {
	_, _, nodeId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	return
}

// ValidateNSNode 校验DNS节点
func (this *BaseService) ValidateNSNode(ctx context.Context) (nodeId int64, err error) {
	_, _, nodeId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeDNS)
	return
}

// ValidateUserNode 校验用户节点
func (this *BaseService) ValidateUserNode(ctx context.Context, canRest bool) (userId int64, err error) {
	// 不允许REST调用
	if !canRest && rpcutils.IsRest(ctx) {
		err = errors.New("can not be called by rest")
		return
	}

	_, _, userId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeUser)
	return
}

// ValidateAuthorityNode 校验认证节点
func (this *BaseService) ValidateAuthorityNode(ctx context.Context) (nodeId int64, err error) {
	_, _, nodeId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAuthority)
	return
}

// ValidateNodeId 获取节点ID
func (this *BaseService) ValidateNodeId(ctx context.Context, roles ...rpcutils.UserType) (role rpcutils.UserType, nodeIntId int64, err error) {
	// 默认包含大部分节点
	if len(roles) == 0 {
		roles = []rpcutils.UserType{rpcutils.UserTypeNode, rpcutils.UserTypeCluster, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser, rpcutils.UserTypeDNS, rpcutils.UserTypeReport, rpcutils.UserTypeLog, rpcutils.UserTypeAPI}
	}

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

	role = apiToken.Role
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
	case rpcutils.UserTypeDNS:
		nodeIntId, err = models.SharedNSNodeDAO.FindEnabledNodeIdWithUniqueId(nil, nodeId)
	case rpcutils.UserTypeReport:
		nodeIntId, err = models.SharedReportNodeDAO.FindEnabledNodeIdWithUniqueId(nil, nodeId)
	case rpcutils.UserTypeAuthority:
		nodeIntId, err = authority.SharedAuthorityNodeDAO.FindEnabledAuthorityNodeIdWithUniqueId(nil, nodeId)
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

// Exists 返回是否存在
func (this *BaseService) Exists(b bool) (*pb.RPCExists, error) {
	return &pb.RPCExists{Exists: b}, nil
}

// PermissionError 返回权限错误
func (this *BaseService) PermissionError() error {
	return errors.New("Permission Denied")
}

func (this *BaseService) NotImplementedYet() error {
	return status.Error(codes.Unimplemented, "not implemented yet")
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

// BeginTag 开始标签统计
func (this *BaseService) BeginTag(ctx context.Context, name string) {
	if !teaconst.Debug {
		return
	}
	traceCtx, ok := ctx.(*rpc.Context)
	if ok {
		traceCtx.Begin(name)
	}
}

// EndTag 结束标签统计
func (this *BaseService) EndTag(ctx context.Context, name string) {
	if !teaconst.Debug {
		return
	}
	traceCtx, ok := ctx.(*rpc.Context)
	if ok {
		traceCtx.End(name)
	}
}
