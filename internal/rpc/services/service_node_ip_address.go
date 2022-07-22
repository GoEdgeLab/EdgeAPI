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
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	addressId, err := models.SharedNodeIPAddressDAO.CreateAddress(tx, adminId, req.NodeId, req.Role, req.Name, req.Ip, req.CanAccess, req.IsUp, 0)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeIPAddressResponse{NodeIPAddressId: addressId}, nil
}

// CreateNodeIPAddresses 批量创建IP地址
func (this *NodeIPAddressService) CreateNodeIPAddresses(ctx context.Context, req *pb.CreateNodeIPAddressesRequest) (*pb.CreateNodeIPAddressesResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 创建分组
	groupId, err := models.SharedNodeIPAddressGroupDAO.CreateGroup(tx, req.GroupValue, req.GroupValue)
	if err != nil {
		return nil, err
	}

	var result = []int64{}
	for _, ip := range req.IpList {
		addressId, err := models.SharedNodeIPAddressDAO.CreateAddress(tx, adminId, req.NodeId, req.Role, req.Name, ip, req.CanAccess, req.IsUp, groupId)
		if err != nil {
			return nil, err
		}
		result = append(result, addressId)
	}

	return &pb.CreateNodeIPAddressesResponse{NodeIPAddressIds: result}, nil
}

// UpdateNodeIPAddress 修改IP地址
func (this *NodeIPAddressService) UpdateNodeIPAddress(ctx context.Context, req *pb.UpdateNodeIPAddressRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNodeIPAddressDAO.UpdateAddress(tx, adminId, req.NodeIPAddressId, req.Name, req.Ip, req.CanAccess, req.IsOn, req.IsUp)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNodeIPAddressNodeId 修改IP地址所属节点
func (this *NodeIPAddressService) UpdateNodeIPAddressNodeId(ctx context.Context, req *pb.UpdateNodeIPAddressNodeIdRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNodeIPAddressDAO.UpdateAddressNodeId(tx, req.NodeIPAddressId, req.NodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DisableNodeIPAddress 禁用单个IP地址
func (this *NodeIPAddressService) DisableNodeIPAddress(ctx context.Context, req *pb.DisableNodeIPAddressRequest) (*pb.DisableNodeIPAddressResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNodeIPAddressDAO.DisableAddress(tx, req.NodeIPAddressId)
	if err != nil {
		return nil, err
	}

	return &pb.DisableNodeIPAddressResponse{}, nil
}

// DisableAllNodeIPAddressesWithNodeId 禁用某个节点的IP地址
func (this *NodeIPAddressService) DisableAllNodeIPAddressesWithNodeId(ctx context.Context, req *pb.DisableAllNodeIPAddressesWithNodeIdRequest) (*pb.DisableAllNodeIPAddressesWithNodeIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNodeIPAddressDAO.DisableAllAddressesWithNodeId(tx, req.NodeId, req.Role)
	if err != nil {
		return nil, err
	}

	return &pb.DisableAllNodeIPAddressesWithNodeIdResponse{}, nil
}

// FindEnabledNodeIPAddress 查找单个IP地址
func (this *NodeIPAddressService) FindEnabledNodeIPAddress(ctx context.Context, req *pb.FindEnabledNodeIPAddressRequest) (*pb.FindEnabledNodeIPAddressResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	address, err := models.SharedNodeIPAddressDAO.FindEnabledAddress(tx, req.NodeIPAddressId)
	if err != nil {
		return nil, err
	}

	var result *pb.NodeIPAddress = nil
	if address != nil {
		result = &pb.NodeIPAddress{
			Id:          int64(address.Id),
			NodeId:      int64(address.NodeId),
			Role:        address.Role,
			Name:        address.Name,
			Ip:          address.Ip,
			Description: address.Description,
			State:       int64(address.State),
			Order:       int64(address.Order),
			CanAccess:   address.CanAccess,
			IsOn:        address.IsOn,
			IsUp:        address.IsUp,
			IsHealthy:   address.IsHealthy,
			BackupIP:    address.DecodeBackupIP(),
		}
	}

	return &pb.FindEnabledNodeIPAddressResponse{NodeIPAddress: result}, nil
}

// FindAllEnabledNodeIPAddressesWithNodeId 查找节点的所有地址
func (this *NodeIPAddressService) FindAllEnabledNodeIPAddressesWithNodeId(ctx context.Context, req *pb.FindAllEnabledNodeIPAddressesWithNodeIdRequest) (*pb.FindAllEnabledNodeIPAddressesWithNodeIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	addresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(tx, req.NodeId, req.Role)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeIPAddress{}
	for _, address := range addresses {
		result = append(result, &pb.NodeIPAddress{
			Id:          int64(address.Id),
			NodeId:      int64(address.NodeId),
			Role:        address.Role,
			Name:        address.Name,
			Ip:          address.Ip,
			Description: address.Description,
			State:       int64(address.State),
			Order:       int64(address.Order),
			CanAccess:   address.CanAccess,
			IsOn:        address.IsOn,
			IsUp:        address.IsUp,
			IsHealthy:   address.IsHealthy,
			BackupIP:    address.DecodeBackupIP(),
		})
	}

	return &pb.FindAllEnabledNodeIPAddressesWithNodeIdResponse{NodeIPAddresses: result}, nil
}

// CountAllEnabledNodeIPAddresses 计算IP地址数量
func (this *NodeIPAddressService) CountAllEnabledNodeIPAddresses(ctx context.Context, req *pb.CountAllEnabledNodeIPAddressesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
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

// ListEnabledNodeIPAddresses 列出单页IP地址
func (this *NodeIPAddressService) ListEnabledNodeIPAddresses(ctx context.Context, req *pb.ListEnabledNodeIPAddressesRequest) (*pb.ListEnabledNodeIPAddressesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
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
			Id:          int64(addr.Id),
			NodeId:      int64(addr.NodeId),
			Role:        addr.Role,
			Name:        addr.Name,
			Ip:          addr.Ip,
			Description: addr.Description,
			CanAccess:   addr.CanAccess,
			IsOn:        addr.IsOn,
			IsUp:        addr.IsUp,
			IsHealthy:   addr.IsHealthy,
			BackupIP:    addr.DecodeBackupIP(),
		})
	}
	return &pb.ListEnabledNodeIPAddressesResponse{NodeIPAddresses: pbAddrs}, nil
}

// UpdateNodeIPAddressIsUp 设置上下线状态
func (this *NodeIPAddressService) UpdateNodeIPAddressIsUp(ctx context.Context, req *pb.UpdateNodeIPAddressIsUpRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeIPAddressDAO.UpdateAddressIsUp(tx, req.NodeIPAddressId, req.IsUp)
	if err != nil {
		return nil, err
	}

	// 增加日志
	if req.IsUp {
		err = models.SharedNodeIPAddressLogDAO.CreateLog(tx, adminId, req.NodeIPAddressId, "手动上线")
	} else {
		err = models.SharedNodeIPAddressLogDAO.CreateLog(tx, adminId, req.NodeIPAddressId, "手动下线")
	}
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// RestoreNodeIPAddressBackupIP 还原备用IP状态
func (this *NodeIPAddressService) RestoreNodeIPAddressBackupIP(ctx context.Context, req *pb.RestoreNodeIPAddressBackupIPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeIPAddressDAO.UpdateAddressBackupIP(tx, req.NodeIPAddressId, 0, "")
	if err != nil {
		return nil, err
	}

	// 增加日志
	err = models.SharedNodeIPAddressLogDAO.CreateLog(tx, adminId, req.NodeIPAddressId, "恢复IP状态")
	if err != nil {
		return nil, err
	}

	return this.Success()
}
