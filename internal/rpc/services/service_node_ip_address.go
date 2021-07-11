package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type NodeIPAddressService struct {
	BaseService
}

// CreateNodeIPAddress 创建IP地址
func (this *NodeIPAddressService) CreateNodeIPAddress(ctx context.Context, req *pb.CreateNodeIPAddressRequest) (*pb.CreateNodeIPAddressResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	addressId, err := models.SharedNodeIPAddressDAO.CreateAddress(tx, req.NodeId, req.Role, req.Name, req.Ip, req.CanAccess)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeIPAddressResponse{AddressId: addressId}, nil
}

// UpdateNodeIPAddress 修改IP地址
func (this *NodeIPAddressService) UpdateNodeIPAddress(ctx context.Context, req *pb.UpdateNodeIPAddressRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeIPAddressDAO.UpdateAddress(tx, req.AddressId, req.Name, req.Ip, req.CanAccess)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNodeIPAddressNodeId 修改IP地址所属节点
func (this *NodeIPAddressService) UpdateNodeIPAddressNodeId(ctx context.Context, req *pb.UpdateNodeIPAddressNodeIdRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeIPAddressDAO.UpdateAddressNodeId(tx, req.AddressId, req.NodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DisableNodeIPAddress 禁用单个IP地址
func (this *NodeIPAddressService) DisableNodeIPAddress(ctx context.Context, req *pb.DisableNodeIPAddressRequest) (*pb.DisableNodeIPAddressResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeIPAddressDAO.DisableAddress(tx, req.AddressId)
	if err != nil {
		return nil, err
	}

	return &pb.DisableNodeIPAddressResponse{}, nil
}

// DisableAllIPAddressesWithNodeId 禁用某个节点的IP地址
func (this *NodeIPAddressService) DisableAllIPAddressesWithNodeId(ctx context.Context, req *pb.DisableAllIPAddressesWithNodeIdRequest) (*pb.DisableAllIPAddressesWithNodeIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeIPAddressDAO.DisableAllAddressesWithNodeId(tx, req.NodeId, req.Role)
	if err != nil {
		return nil, err
	}

	return &pb.DisableAllIPAddressesWithNodeIdResponse{}, nil
}

// FindEnabledNodeIPAddress 查找单个IP地址
func (this *NodeIPAddressService) FindEnabledNodeIPAddress(ctx context.Context, req *pb.FindEnabledNodeIPAddressRequest) (*pb.FindEnabledNodeIPAddressResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	address, err := models.SharedNodeIPAddressDAO.FindEnabledAddress(tx, req.AddressId)
	if err != nil {
		return nil, err
	}

	var result *pb.NodeIPAddress = nil
	if address != nil {
		result = &pb.NodeIPAddress{
			Id:          int64(address.Id),
			NodeId:      int64(address.NodeId),
			Name:        address.Name,
			Ip:          address.Ip,
			Description: address.Description,
			State:       int64(address.State),
			Order:       int64(address.Order),
			CanAccess:   address.CanAccess == 1,
		}
	}

	return &pb.FindEnabledNodeIPAddressResponse{IpAddress: result}, nil
}

// FindAllEnabledIPAddressesWithNodeId 查找节点的所有地址
func (this *NodeIPAddressService) FindAllEnabledIPAddressesWithNodeId(ctx context.Context, req *pb.FindAllEnabledIPAddressesWithNodeIdRequest) (*pb.FindAllEnabledIPAddressesWithNodeIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	addresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(tx, req.NodeId, req.Role)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeIPAddress{}
	for _, address := range addresses {
		result = append(result, &pb.NodeIPAddress{
			Id:          int64(address.Id),
			NodeId:      int64(address.NodeId),
			Name:        address.Name,
			Ip:          address.Ip,
			Description: address.Description,
			State:       int64(address.State),
			Order:       int64(address.Order),
			CanAccess:   address.CanAccess == 1,
		})
	}

	return &pb.FindAllEnabledIPAddressesWithNodeIdResponse{Addresses: result}, nil
}
