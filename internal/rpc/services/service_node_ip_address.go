package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

type NodeIPAddressService struct {
	BaseService
}

// CreateNodeIPAddress 创建IP地址
func (this *NodeIPAddressService) CreateNodeIPAddress(ctx context.Context, req *pb.CreateNodeIPAddressRequest) (*pb.CreateNodeIPAddressResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	addressId, err := models.SharedNodeIPAddressDAO.CreateAddress(tx, adminId, req.NodeId, req.Role, req.Name, req.Ip, req.CanAccess, req.ThresholdsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeIPAddressResponse{NodeIPAddressId: addressId}, nil
}

// UpdateNodeIPAddress 修改IP地址
func (this *NodeIPAddressService) UpdateNodeIPAddress(ctx context.Context, req *pb.UpdateNodeIPAddressRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeIPAddressDAO.UpdateAddress(tx, adminId, req.NodeIPAddressId, req.Name, req.Ip, req.CanAccess, req.IsOn, req.ThresholdsJSON)
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

	err = models.SharedNodeIPAddressDAO.UpdateAddressNodeId(tx, req.NodeIPAddressId, req.NodeId)
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

	err = models.SharedNodeIPAddressDAO.DisableAddress(tx, req.NodeIPAddressId)
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

	address, err := models.SharedNodeIPAddressDAO.FindEnabledAddress(tx, req.NodeIPAddressId)
	if err != nil {
		return nil, err
	}

	var result *pb.NodeIPAddress = nil
	if address != nil {
		result = &pb.NodeIPAddress{
			Id:             int64(address.Id),
			NodeId:         int64(address.NodeId),
			Role:           address.Role,
			Name:           address.Name,
			Ip:             address.Ip,
			Description:    address.Description,
			State:          int64(address.State),
			Order:          int64(address.Order),
			CanAccess:      address.CanAccess == 1,
			IsOn:           address.IsOn == 1,
			IsUp:           address.IsUp == 1,
			ThresholdsJSON: []byte(address.Thresholds),
		}
	}

	return &pb.FindEnabledNodeIPAddressResponse{NodeIPAddress: result}, nil
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
			Id:             int64(address.Id),
			NodeId:         int64(address.NodeId),
			Role:           address.Role,
			Name:           address.Name,
			Ip:             address.Ip,
			Description:    address.Description,
			State:          int64(address.State),
			Order:          int64(address.Order),
			CanAccess:      address.CanAccess == 1,
			IsOn:           address.IsOn == 1,
			IsUp:           address.IsUp == 1,
			ThresholdsJSON: []byte(address.Thresholds),
		})
	}

	return &pb.FindAllEnabledIPAddressesWithNodeIdResponse{Addresses: result}, nil
}

// CountAllEnabledIPAddresses 计算IP地址数量
func (this *NodeIPAddressService) CountAllEnabledIPAddresses(ctx context.Context, req *pb.CountAllEnabledIPAddressesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNodeIPAddressDAO.CountAllEnabledIPAddresses(tx, req.Role, req.NodeClusterId, types.Int8(req.UpState), req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledIPAddresses 列出单页IP地址
func (this *NodeIPAddressService) ListEnabledIPAddresses(ctx context.Context, req *pb.ListEnabledIPAddressesRequest) (*pb.ListEnabledIPAddressesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	addresses, err := models.SharedNodeIPAddressDAO.ListEnabledIPAddresses(tx, req.Role, req.NodeClusterId, types.Int8(req.UpState), req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbAddrs = []*pb.NodeIPAddress{}
	for _, addr := range addresses {
		pbAddrs = append(pbAddrs, &pb.NodeIPAddress{
			Id:             int64(addr.Id),
			NodeId:         int64(addr.NodeId),
			Role:           addr.Role,
			Name:           addr.Name,
			Ip:             addr.Ip,
			Description:    addr.Description,
			CanAccess:      addr.CanAccess == 1,
			IsOn:           addr.IsOn == 1,
			IsUp:           addr.IsUp == 1,
			ThresholdsJSON: []byte(addr.Thresholds),
		})
	}
	return &pb.ListEnabledIPAddressesResponse{NodeIPAddresses: pbAddrs}, nil
}
