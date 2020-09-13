package services

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type ServerService struct {
}

// 创建服务
func (this *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.CreateServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	serverId, err := models.SharedServerDAO.CreateServer(req.AdminId, req.UserId, req.Type, req.Name, req.Description, req.ClusterId, string(req.Config), string(req.IncludeNodesJSON), string(req.ExcludeNodesJSON))
	if err != nil {
		return nil, err
	}

	// 更新节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(req.ClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

// 修改服务
func (this *ServerService) UpdateServerBasic(ctx context.Context, req *pb.UpdateServerBasicRequest) (*pb.UpdateServerBasicResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	err = models.SharedServerDAO.UpdateServerBasic(req.ServerId, req.Name, req.Description, req.ClusterId)
	if err != nil {
		return nil, err
	}

	// 更新老的节点版本
	if req.ClusterId != int64(server.ClusterId) {
		err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
		if err != nil {
			return nil, err
		}
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(req.ClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateServerBasicResponse{}, nil
}

// 修改服务配置
func (this *ServerService) UpdateServerConfig(ctx context.Context, req *pb.UpdateServerConfigRequest) (*pb.UpdateServerConfigResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查找Server
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return &pb.UpdateServerConfigResponse{}, nil
	}

	// 修改
	err = models.SharedServerDAO.UpdateServerConfig(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.UpdateServerConfigResponse{}, nil
}

// 计算服务数量
func (this *ServerService) CountAllEnabledServers(ctx context.Context, req *pb.CountAllEnabledServersRequest) (*pb.CountAllEnabledServersResponse, error) {
	_ = req
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedServerDAO.CountAllEnabledServers()
	if err != nil {
		return nil, err
	}

	return &pb.CountAllEnabledServersResponse{Count: count}, nil
}

// 列出单页服务
func (this *ServerService) ListEnabledServers(ctx context.Context, req *pb.ListEnabledServersRequest) (*pb.ListEnabledServersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	servers, err := models.SharedServerDAO.ListEnabledServers(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.Server{}
	for _, server := range servers {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(server.ClusterId))
		if err != nil {
			return nil, err
		}
		result = append(result, &pb.Server{
			Id:           int64(server.Id),
			Type:         server.Type,
			Name:         server.Name,
			Description:  server.Description,
			Config:       []byte(server.Config),
			IncludeNodes: []byte(server.IncludeNodes),
			ExcludeNodes: []byte(server.ExcludeNodes),
			CreatedAt:    int64(server.CreatedAt),
			Cluster: &pb.NodeCluster{
				Id:   int64(server.ClusterId),
				Name: clusterName,
			},
		})
	}

	return &pb.ListEnabledServersResponse{Servers: result}, nil
}

// 禁用某服务
func (this *ServerService) DisableServer(ctx context.Context, req *pb.DisableServerRequest) (*pb.DisableServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	// 查找服务
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find the server")
	}

	// 禁用服务
	err = models.SharedServerDAO.DisableServer(req.ServerId)
	if err != nil {
		return nil, err
	}

	// 更新节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.DisableServerResponse{}, nil
}

// 查找单个服务
func (this *ServerService) FindEnabledServer(ctx context.Context, req *pb.FindEnabledServerRequest) (*pb.FindEnabledServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}

	if server == nil {
		return &pb.FindEnabledServerResponse{}, nil
	}

	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerResponse{Server: &pb.Server{
		Id:           int64(server.Id),
		Type:         server.Type,
		Name:         server.Name,
		Description:  server.Description,
		Config:       []byte(server.Config),
		IncludeNodes: []byte(server.IncludeNodes),
		ExcludeNodes: []byte(server.ExcludeNodes),
		CreatedAt:    int64(server.CreatedAt),
		Cluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
	}}, nil
}
