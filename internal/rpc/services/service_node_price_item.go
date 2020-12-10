package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 节点区域价格相关服务
type NodePriceItemService struct {
	BaseService
}

// 创建区域价格
func (this *NodePriceItemService) CreateNodePriceItem(ctx context.Context, req *pb.CreateNodePriceItemRequest) (*pb.CreateNodePriceItemResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	itemId, err := models.SharedNodePriceItemDAO.CreateItem(req.Name, req.Type, req.BitsFrom, req.BitsTo)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodePriceItemResponse{NodePriceItemId: itemId}, nil
}

// 修改区域价格
func (this *NodePriceItemService) UpdateNodePriceItem(ctx context.Context, req *pb.UpdateNodePriceItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodePriceItemDAO.UpdateItem(req.NodePriceItemId, req.Name, req.BitsFrom, req.BitsTo)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除区域价格
func (this *NodePriceItemService) DeleteNodePriceItem(ctx context.Context, req *pb.DeleteNodePriceItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodePriceItemDAO.DisableNodePriceItem(req.NodePriceItemId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查找所有区域价格
func (this *NodePriceItemService) FindAllEnabledNodePriceItems(ctx context.Context, req *pb.FindAllEnabledNodePriceItemsRequest) (*pb.FindAllEnabledNodePriceItemsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	prices, err := models.SharedNodePriceItemDAO.FindAllEnabledRegionPrices(req.Type)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodePriceItem{}
	for _, price := range prices {
		result = append(result, &pb.NodePriceItem{
			Id:       int64(price.Id),
			IsOn:     price.IsOn == 1,
			Name:     price.Name,
			Type:     price.Type,
			BitsFrom: int64(price.BitsFrom),
			BitsTo:   int64(price.BitsTo),
		})
	}

	return &pb.FindAllEnabledNodePriceItemsResponse{NodePriceItems: result}, nil
}

// 查找所有启用的区域价格
func (this *NodePriceItemService) FindAllEnabledAndOnNodePriceItems(ctx context.Context, req *pb.FindAllEnabledAndOnNodePriceItemsRequest) (*pb.FindAllEnabledAndOnNodePriceItemsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	prices, err := models.SharedNodePriceItemDAO.FindAllEnabledAndOnRegionPrices(req.Type)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodePriceItem{}
	for _, price := range prices {
		result = append(result, &pb.NodePriceItem{
			Id:       int64(price.Id),
			IsOn:     price.IsOn == 1,
			Name:     price.Name,
			Type:     price.Type,
			BitsFrom: int64(price.BitsFrom),
			BitsTo:   int64(price.BitsTo),
		})
	}

	return &pb.FindAllEnabledAndOnNodePriceItemsResponse{NodePriceItems: result}, nil
}

// 查找单个区域信息
func (this *NodePriceItemService) FindEnabledNodePriceItem(ctx context.Context, req *pb.FindEnabledNodePriceItemRequest) (*pb.FindEnabledNodePriceItemResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	price, err := models.SharedNodePriceItemDAO.FindEnabledNodePriceItem(req.NodePriceItemId)
	if err != nil {
		return nil, err
	}
	if price == nil {
		return &pb.FindEnabledNodePriceItemResponse{NodePriceItem: nil}, nil
	}
	return &pb.FindEnabledNodePriceItemResponse{NodePriceItem: &pb.NodePriceItem{
		Id:       int64(price.Id),
		IsOn:     price.IsOn == 1,
		Name:     price.Name,
		Type:     price.Type,
		BitsFrom: int64(price.BitsFrom),
		BitsTo:   int64(price.BitsTo),
	}}, nil
}
