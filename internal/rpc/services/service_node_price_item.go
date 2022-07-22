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
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	itemId, err := models.SharedNodePriceItemDAO.CreateItem(tx, req.Name, req.Type, req.BitsFrom, req.BitsTo)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodePriceItemResponse{NodePriceItemId: itemId}, nil
}

// 修改区域价格
func (this *NodePriceItemService) UpdateNodePriceItem(ctx context.Context, req *pb.UpdateNodePriceItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodePriceItemDAO.UpdateItem(tx, req.NodePriceItemId, req.Name, req.BitsFrom, req.BitsTo)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除区域价格
func (this *NodePriceItemService) DeleteNodePriceItem(ctx context.Context, req *pb.DeleteNodePriceItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodePriceItemDAO.DisableNodePriceItem(tx, req.NodePriceItemId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查找所有区域价格
func (this *NodePriceItemService) FindAllEnabledNodePriceItems(ctx context.Context, req *pb.FindAllEnabledNodePriceItemsRequest) (*pb.FindAllEnabledNodePriceItemsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	prices, err := models.SharedNodePriceItemDAO.FindAllEnabledRegionPrices(tx, req.Type)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodePriceItem{}
	for _, price := range prices {
		result = append(result, &pb.NodePriceItem{
			Id:       int64(price.Id),
			IsOn:     price.IsOn,
			Name:     price.Name,
			Type:     price.Type,
			BitsFrom: int64(price.BitsFrom),
			BitsTo:   int64(price.BitsTo),
		})
	}

	return &pb.FindAllEnabledNodePriceItemsResponse{NodePriceItems: result}, nil
}

// FindAllEnabledAndOnNodePriceItems 查找所有启用的区域价格
func (this *NodePriceItemService) FindAllEnabledAndOnNodePriceItems(ctx context.Context, req *pb.FindAllEnabledAndOnNodePriceItemsRequest) (*pb.FindAllEnabledAndOnNodePriceItemsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	prices, err := models.SharedNodePriceItemDAO.FindAllEnabledAndOnRegionPrices(tx, req.Type)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodePriceItem{}
	for _, price := range prices {
		result = append(result, &pb.NodePriceItem{
			Id:       int64(price.Id),
			IsOn:     price.IsOn,
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
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	price, err := models.SharedNodePriceItemDAO.FindEnabledNodePriceItem(tx, req.NodePriceItemId)
	if err != nil {
		return nil, err
	}
	if price == nil {
		return &pb.FindEnabledNodePriceItemResponse{NodePriceItem: nil}, nil
	}
	return &pb.FindEnabledNodePriceItemResponse{NodePriceItem: &pb.NodePriceItem{
		Id:       int64(price.Id),
		IsOn:     price.IsOn,
		Name:     price.Name,
		Type:     price.Type,
		BitsFrom: int64(price.BitsFrom),
		BitsTo:   int64(price.BitsTo),
	}}, nil
}
