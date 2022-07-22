package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// MessageMediaService 消息媒介服务
type MessageMediaService struct {
	BaseService
}

// FindAllMessageMedias 获取所有支持的媒介
func (this *MessageMediaService) FindAllMessageMedias(ctx context.Context, req *pb.FindAllMessageMediasRequest) (*pb.FindAllMessageMediasResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	medias, err := models.SharedMessageMediaDAO.FindAllEnabledMessageMedias(tx)
	if err != nil {
		return nil, err
	}
	pbMedias := []*pb.MessageMedia{}
	for _, media := range medias {
		pbMedias = append(pbMedias, &pb.MessageMedia{
			Id:              int64(media.Id),
			Type:            media.Type,
			Name:            media.Name,
			Description:     media.Description,
			UserDescription: media.UserDescription,
			IsOn:            media.IsOn,
		})
	}
	return &pb.FindAllMessageMediasResponse{MessageMedias: pbMedias}, nil
}

// UpdateMessageMedias 设置所有支持的媒介
func (this *MessageMediaService) UpdateMessageMedias(ctx context.Context, req *pb.UpdateMessageMediasRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateMonitorNode(ctx)
	if err != nil {
		return nil, err
	}

	mediaMaps := []maps.Map{}
	for _, media := range req.MessageMedias {
		mediaMaps = append(mediaMaps, maps.Map{
			"name":            media.Name,
			"type":            media.Type,
			"description":     media.Description,
			"userDescription": media.UserDescription,
			"isOn":            media.IsOn,
		})
	}

	var tx = this.NullTx()
	err = models.SharedMessageMediaDAO.UpdateMessageMedias(tx, mediaMaps)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
