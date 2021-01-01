package services

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type NodeGrantService struct {
	BaseService
}

func (this *NodeGrantService) CreateNodeGrant(ctx context.Context, req *pb.CreateNodeGrantRequest) (*pb.CreateNodeGrantResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	grantId, err := models.SharedNodeGrantDAO.CreateGrant(tx, adminId, req.Name, req.Method, req.Username, req.Password, req.PrivateKey, req.Description, req.NodeId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeGrantResponse{
		GrantId: grantId,
	}, err
}

func (this *NodeGrantService) UpdateNodeGrant(ctx context.Context, req *pb.UpdateNodeGrantRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.GrantId <= 0 {
		return nil, errors.New("wrong grantId")
	}

	tx := this.NullTx()

	err = models.SharedNodeGrantDAO.UpdateGrant(tx, req.GrantId, req.Name, req.Method, req.Username, req.Password, req.PrivateKey, req.Description, req.NodeId)
	return this.Success()
}

func (this *NodeGrantService) DisableNodeGrant(ctx context.Context, req *pb.DisableNodeGrantRequest) (*pb.DisableNodeGrantResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeGrantDAO.DisableNodeGrant(tx, req.GrantId)
	return &pb.DisableNodeGrantResponse{}, err
}

func (this *NodeGrantService) CountAllEnabledNodeGrants(ctx context.Context, req *pb.CountAllEnabledNodeGrantsRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeGrantDAO.CountAllEnabledGrants(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

func (this *NodeGrantService) ListEnabledNodeGrants(ctx context.Context, req *pb.ListEnabledNodeGrantsRequest) (*pb.ListEnabledNodeGrantsResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	grants, err := models.SharedNodeGrantDAO.ListEnabledGrants(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeGrant{}
	for _, grant := range grants {
		result = append(result, &pb.NodeGrant{
			Id:          int64(grant.Id),
			Name:        grant.Name,
			Method:      grant.Method,
			Password:    grant.Password,
			Su:          grant.Su == 1,
			PrivateKey:  grant.PrivateKey,
			Description: grant.Description,
			NodeId:      int64(grant.NodeId),
		})
	}

	return &pb.ListEnabledNodeGrantsResponse{Grants: result}, nil
}

// 列出所有认证信息
func (this *NodeGrantService) FindAllEnabledNodeGrants(ctx context.Context, req *pb.FindAllEnabledNodeGrantsRequest) (*pb.FindAllEnabledNodeGrantsResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	grants, err := models.SharedNodeGrantDAO.FindAllEnabledGrants(this.NullTx())
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeGrant{}
	for _, grant := range grants {
		result = append(result, &pb.NodeGrant{
			Id:          int64(grant.Id),
			Name:        grant.Name,
			Method:      grant.Method,
			Password:    grant.Password,
			Su:          grant.Su == 1,
			PrivateKey:  grant.PrivateKey,
			Description: grant.Description,
			NodeId:      int64(grant.NodeId),
		})
	}

	return &pb.FindAllEnabledNodeGrantsResponse{Grants: result}, nil
}

func (this *NodeGrantService) FindEnabledGrant(ctx context.Context, req *pb.FindEnabledGrantRequest) (*pb.FindEnabledGrantResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(this.NullTx(), req.GrantId)
	if err != nil {
		return nil, err
	}
	if grant == nil {
		return &pb.FindEnabledGrantResponse{}, nil
	}
	return &pb.FindEnabledGrantResponse{Grant: &pb.NodeGrant{
		Id:          int64(grant.Id),
		Name:        grant.Name,
		Method:      grant.Method,
		Username:    grant.Username,
		Password:    grant.Password,
		Su:          grant.Su == 1,
		PrivateKey:  grant.PrivateKey,
		Description: grant.Description,
		NodeId:      int64(grant.NodeId),
	}}, nil
}
