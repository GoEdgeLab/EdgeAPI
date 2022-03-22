package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NodeRegionService 节点区域相关服务
type NodeRegionService struct {
	BaseService
}

// CreateNodeRegion 创建区域
func (this *NodeRegionService) CreateNodeRegion(ctx context.Context, req *pb.CreateNodeRegionRequest) (*pb.CreateNodeRegionResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	regionId, err := models.SharedNodeRegionDAO.CreateRegion(tx, adminId, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeRegionResponse{NodeRegionId: regionId}, nil
}

// UpdateNodeRegion 修改区域
func (this *NodeRegionService) UpdateNodeRegion(ctx context.Context, req *pb.UpdateNodeRegionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeRegionDAO.UpdateRegion(tx, req.NodeRegionId, req.Name, req.Description, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNodeRegion 删除区域
func (this *NodeRegionService) DeleteNodeRegion(ctx context.Context, req *pb.DeleteNodeRegionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeRegionDAO.DisableNodeRegion(tx, req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllEnabledNodeRegions 查找所有区域
func (this *NodeRegionService) FindAllEnabledNodeRegions(ctx context.Context, req *pb.FindAllEnabledNodeRegionsRequest) (*pb.FindAllEnabledNodeRegionsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	regions, err := models.SharedNodeRegionDAO.FindAllEnabledRegions(tx)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeRegion{}
	for _, region := range regions {
		result = append(result, &pb.NodeRegion{
			Id:          int64(region.Id),
			IsOn:        region.IsOn == 1,
			Name:        region.Name,
			Description: region.Description,
			PricesJSON:  region.Prices,
		})
	}
	return &pb.FindAllEnabledNodeRegionsResponse{NodeRegions: result}, nil
}

// FindAllEnabledAndOnNodeRegions 查找所有启用的区域
func (this *NodeRegionService) FindAllEnabledAndOnNodeRegions(ctx context.Context, req *pb.FindAllEnabledAndOnNodeRegionsRequest) (*pb.FindAllEnabledAndOnNodeRegionsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	regions, err := models.SharedNodeRegionDAO.FindAllEnabledAndOnRegions(tx)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeRegion{}
	for _, region := range regions {
		result = append(result, &pb.NodeRegion{
			Id:          int64(region.Id),
			IsOn:        region.IsOn == 1,
			Name:        region.Name,
			Description: region.Description,
			PricesJSON:  region.Prices,
		})
	}
	return &pb.FindAllEnabledAndOnNodeRegionsResponse{NodeRegions: result}, nil
}

// UpdateNodeRegionOrders 排序
func (this *NodeRegionService) UpdateNodeRegionOrders(ctx context.Context, req *pb.UpdateNodeRegionOrdersRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeRegionDAO.UpdateRegionOrders(tx, req.NodeRegionIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledNodeRegion 查找单个区域信息
func (this *NodeRegionService) FindEnabledNodeRegion(ctx context.Context, req *pb.FindEnabledNodeRegionRequest) (*pb.FindEnabledNodeRegionResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(tx, req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	if region == nil {
		return &pb.FindEnabledNodeRegionResponse{NodeRegion: nil}, nil
	}
	return &pb.FindEnabledNodeRegionResponse{NodeRegion: &pb.NodeRegion{
		Id:          int64(region.Id),
		IsOn:        region.IsOn == 1,
		Name:        region.Name,
		Description: region.Description,
		PricesJSON:  region.Prices,
	}}, nil
}

// UpdateNodeRegionPrice 修改价格项价格
func (this *NodeRegionService) UpdateNodeRegionPrice(ctx context.Context, req *pb.UpdateNodeRegionPriceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeRegionDAO.UpdateRegionItemPrice(tx, req.NodeRegionId, req.NodeItemId, req.Price)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
