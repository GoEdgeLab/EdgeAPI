package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 节点区域相关服务
type NodeRegionService struct {
	BaseService
}

// 创建区域
func (this *NodeRegionService) CreateNodeRegion(ctx context.Context, req *pb.CreateNodeRegionRequest) (*pb.CreateNodeRegionResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	regionId, err := models.SharedNodeRegionDAO.CreateRegion(adminId, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeRegionResponse{NodeRegionId: regionId}, nil
}

// 修改区域
func (this *NodeRegionService) UpdateNodeRegion(ctx context.Context, req *pb.UpdateNodeRegionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	err = models.SharedNodeRegionDAO.UpdateRegion(req.NodeRegionId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除区域
func (this *NodeRegionService) DeleteNodeRegion(ctx context.Context, req *pb.DeleteNodeRegionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	err = models.SharedNodeRegionDAO.DisableNodeRegion(req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查找所有区域
func (this *NodeRegionService) FindAllEnabledNodeRegions(ctx context.Context, req *pb.FindAllEnabledNodeRegionsRequest) (*pb.FindAllEnabledNodeRegionsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	regions, err := models.SharedNodeRegionDAO.FindAllEnabledRegions()
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeRegion{}
	for _, region := range regions {
		result = append(result, &pb.NodeRegion{
			Id:   int64(region.Id),
			IsOn: region.IsOn == 1,
			Name: region.Name,
		})
	}
	return &pb.FindAllEnabledNodeRegionsResponse{NodeRegions: result}, nil
}

// 查找所有启用的区域
func (this *NodeRegionService) FindAllEnabledAndOnNodeRegions(ctx context.Context, req *pb.FindAllEnabledAndOnNodeRegionsRequest) (*pb.FindAllEnabledAndOnNodeRegionsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	regions, err := models.SharedNodeRegionDAO.FindAllEnabledAndOnRegions()
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeRegion{}
	for _, region := range regions {
		result = append(result, &pb.NodeRegion{
			Id:   int64(region.Id),
			IsOn: region.IsOn == 1,
			Name: region.Name,
		})
	}
	return &pb.FindAllEnabledAndOnNodeRegionsResponse{NodeRegions: result}, nil
}

// 排序
func (this *NodeRegionService) UpdateNodeRegionOrders(ctx context.Context, req *pb.UpdateNodeRegionOrdersRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	err = models.SharedNodeRegionDAO.UpdateRegionOrders(req.NodeRegionIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查找单个区域信息
func (this *NodeRegionService) FindEnabledNodeRegion(ctx context.Context, req *pb.FindEnabledNodeRegionRequest) (*pb.FindEnabledNodeRegionResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	if region == nil {
		return &pb.FindEnabledNodeRegionResponse{NodeRegion: nil}, nil
	}
	return &pb.FindEnabledNodeRegionResponse{NodeRegion: &pb.NodeRegion{
		Id:   int64(region.Id),
		IsOn: region.IsOn == 1,
		Name: region.Name,
	}}, nil
}
