package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
)

type NodeIPAddressService struct {
}

// 创建IP地址
func (this *NodeIPAddressService) CreateNodeIPAddress(ctx context.Context, req *pb.CreateNodeIPAddressRequest) (*pb.CreateNodeIPAddressResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	addressId, err := models.SharedNodeIPAddressDAO.CreateAddress(req.NodeId, req.Name, req.Ip)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeIPAddressResponse{AddressId: addressId}, nil
}

// 修改IP地址
func (this *NodeIPAddressService) UpdateNodeIPAddress(ctx context.Context, req *pb.UpdateNodeIPAddressRequest) (*pb.UpdateNodeIPAddressResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeIPAddressDAO.UpdateAddress(req.AddressId, req.Name, req.Ip)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateNodeIPAddressResponse{}, nil
}

// 修改IP地址所属节点
func (this *NodeIPAddressService) UpdateNodeIPAddressNodeId(ctx context.Context, req *pb.UpdateNodeIPAddressNodeIdRequest) (*pb.UpdateNodeIPAddressNodeIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeIPAddressDAO.UpdateAddressNodeId(req.AddressId, req.NodeId)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateNodeIPAddressNodeIdResponse{}, nil
}

// 禁用单个IP地址
func (this *NodeIPAddressService) DisableNodeIPAddress(ctx context.Context, req *pb.DisableNodeIPAddressRequest) (*pb.DisableNodeIPAddressResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeIPAddressDAO.DisableAddress(req.AddressId)
	if err != nil {
		return nil, err
	}

	return &pb.DisableNodeIPAddressResponse{}, nil
}

// 禁用某个节点的IP地址
func (this *NodeIPAddressService) DisableAllIPAddressesWithNodeId(ctx context.Context, req *pb.DisableAllIPAddressesWithNodeIdRequest) (*pb.DisableAllIPAddressesWithNodeIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeIPAddressDAO.DisableAllAddressesWithNodeId(req.NodeId)
	if err != nil {
		return nil, err
	}

	return &pb.DisableAllIPAddressesWithNodeIdResponse{}, nil
}

// 查找单个IP地址
func (this *NodeIPAddressService) FindEnabledNodeIPAddress(ctx context.Context, req *pb.FindEnabledNodeIPAddressRequest) (*pb.FindEnabledNodeIPAddressResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	address, err := models.SharedNodeIPAddressDAO.FindEnabledAddress(req.AddressId)
	if err != nil {
		return nil, err
	}

	var result *pb.NodeIPAddress = nil
	if address != nil {
		result = &pb.NodeIPAddress{
			Id:          int64(address.Id),
			NodeId:      int64(address.NodeId),
			Name:        address.Name,
			Ip:          address.IP,
			Description: address.Description,
			State:       int64(address.State),
			Order:       int64(address.Order),
		}
	}

	return &pb.FindEnabledNodeIPAddressResponse{IpAddress: result}, nil
}

// 查找节点的所有地址
func (this *NodeIPAddressService) FindAllEnabledIPAddressesWithNodeId(ctx context.Context, req *pb.FindAllEnabledIPAddressesWithNodeIdRequest) (*pb.FindAllEnabledIPAddressesWithNodeIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	addresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(req.NodeId)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeIPAddress{}
	for _, address := range addresses {
		result = append(result, &pb.NodeIPAddress{
			Id:          int64(address.Id),
			NodeId:      int64(address.NodeId),
			Name:        address.Name,
			Ip:          address.IP,
			Description: address.Description,
			State:       int64(address.State),
			Order:       int64(address.Order),
		})
	}

	return &pb.FindAllEnabledIPAddressesWithNodeIdResponse{Addresses: result}, nil
}
