package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
)

type ServerService struct {
}

func (this *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.CreateServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	serverId, err := models.SharedServerDAO.CreateServer(req.AdminId, req.UserId, req.ClusterId, string(req.Config), string(req.IncludeNodesJSON), string(req.ExcludeNodesJSON))
	if err != nil {
		return nil, err
	}
	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

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
