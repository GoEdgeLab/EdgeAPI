package rpcutils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/authority"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/encrypt"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"google.golang.org/grpc/metadata"
	"time"
)

type UserType = string

const (
	UserTypeNone      = "none"
	UserTypeAdmin     = "admin"
	UserTypeUser      = "user"
	UserTypeProvider  = "provider"
	UserTypeNode      = "node"
	UserTypeCluster   = "cluster"
	UserTypeMonitor   = "monitor"
	UserTypeStat      = "stat"
	UserTypeDNS       = "dns"
	UserTypeLog       = "log"
	UserTypeAPI       = "api"
	UserTypeAuthority = "authority"
)

// ValidateRequest 校验请求
func ValidateRequest(ctx context.Context, userTypes ...UserType) (userType UserType, userId int64, err error) {
	if ctx == nil {
		err = errors.New("context should not be nil")
		return
	}

	// 支持直接认证
	plainCtx, ok := ctx.(*PlainContext)
	if ok {
		userType = plainCtx.UserType
		userId = plainCtx.UserId

		if len(userTypes) > 0 && !lists.ContainsString(userTypes, userType) {
			userType = UserTypeNone
			userId = 0
		}

		if userId <= 0 {
			err = errors.New("context: can not find user or permission denied")
		}

		return
	}

	// 是否是模拟测试
	{
		mockCtx, isMock := ctx.(*MockNodeContext)
		if isMock {
			return UserTypeNode, mockCtx.NodeId, nil
		}
	}

	{
		mockCtx, isMock := ctx.(*MockAdminNodeContext)
		if isMock {
			return UserTypeAdmin, mockCtx.AdminId, nil
		}
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return UserTypeNone, 0, errors.New("context: need 'nodeId'")
	}
	nodeIds := md.Get("nodeid")
	if len(nodeIds) == 0 || len(nodeIds[0]) == 0 {
		return UserTypeNone, 0, errors.New("context: need 'nodeId'")
	}
	nodeId := nodeIds[0]

	// 获取角色Node信息
	apiToken, err := models.SharedApiTokenDAO.FindEnabledTokenWithNodeCacheable(nil, nodeId)
	if err != nil {
		utils.PrintError(err)
		return UserTypeNone, 0, err
	}
	nodeUserId := int64(0)
	if apiToken == nil {
		return UserTypeNode, 0, errors.New("context: can not find api token for node '" + nodeId + "'")
	}

	tokens := md.Get("token")
	if len(tokens) == 0 || len(tokens[0]) == 0 {
		return UserTypeNone, 0, errors.New("context: need 'token'")
	}
	token := tokens[0]

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return UserTypeNone, 0, err
	}

	method, err := encrypt.NewMethodInstance(teaconst.EncryptMethod, apiToken.Secret, nodeId)
	if err != nil {
		utils.PrintError(err)
		return UserTypeNone, 0, err
	}
	data, err = method.Decrypt(data)
	if err != nil {
		return UserTypeNone, 0, err
	}
	if len(data) == 0 {
		return UserTypeNone, 0, errors.New("invalid token")
	}

	m := maps.Map{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return UserTypeNone, 0, errors.New("decode token error: " + err.Error())
	}

	timestamp := m.GetInt64("timestamp")
	if time.Now().Unix()-timestamp > 600 {
		// 请求超过10分钟认为超时
		return UserTypeNone, 0, errors.New("authenticate timeout, please check your system clock")
	}

	t := m.GetString("type")
	if len(userTypes) > 0 && !lists.ContainsString(userTypes, t) {
		return UserTypeNone, 0, errors.New("not supported node type: '" + t + "'")
	}

	switch apiToken.Role {
	case UserTypeNode:
		// TODO 需要检查集群是否已经删除
		nodeIntId, err := models.SharedNodeDAO.FindEnabledNodeIdWithUniqueIdCacheable(nil, nodeId)
		if err != nil {
			return UserTypeNode, 0, errors.New("context: " + err.Error())
		}
		if nodeIntId <= 0 {
			return UserTypeNode, 0, errors.New("context: not found node with id '" + nodeId + "'")
		}
		nodeUserId = nodeIntId
	case UserTypeCluster:
		clusterId, err := models.SharedNodeClusterDAO.FindEnabledClusterIdWithUniqueId(nil, nodeId)
		if err != nil {
			return UserTypeCluster, 0, errors.New("context: " + err.Error())
		}
		if clusterId <= 0 {
			return UserTypeCluster, 0, errors.New("context: not found cluster with id '" + nodeId + "'")
		}
		nodeUserId = clusterId
	case UserTypeUser:
	case UserTypeMonitor:
	case UserTypeAuthority:
		nodeIntId, err := authority.SharedAuthorityNodeDAO.FindEnabledAuthorityNodeIdWithUniqueId(nil, nodeId)
		if err != nil {
			return UserTypeNode, 0, errors.New("context: " + err.Error())
		}
		if nodeIntId <= 0 {
			return UserTypeNode, 0, errors.New("context: not found node with id '" + nodeId + "'")
		}
		nodeUserId = nodeIntId
	case UserTypeDNS:
		nodeIntId, err := nameservers.SharedNSNodeDAO.FindEnabledNodeIdWithUniqueId(nil, nodeId)
		if err != nil {
			return UserTypeNode, 0, errors.New("context: " + err.Error())
		}
		if nodeIntId <= 0 {
			return UserTypeNode, 0, errors.New("context: not found node with id '" + nodeId + "'")
		}
		nodeUserId = nodeIntId
	}

	if nodeUserId > 0 {
		return t, nodeUserId, nil
	} else {
		return t, m.GetInt64("userId"), nil
	}
}

// Wrap 包装错误
func Wrap(description string, err error) error {
	if err == nil {
		return errors.New(description)
	}
	return errors.New(description + ": " + err.Error())
}
