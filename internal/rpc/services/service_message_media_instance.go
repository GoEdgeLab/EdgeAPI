package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

// MessageMediaInstanceService 消息媒介实例服务
type MessageMediaInstanceService struct {
	BaseService
}

// CreateMessageMediaInstance 创建消息媒介实例
func (this *MessageMediaInstanceService) CreateMessageMediaInstance(ctx context.Context, req *pb.CreateMessageMediaInstanceRequest) (*pb.CreateMessageMediaInstanceResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	params := maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}

	instanceId, err := models.SharedMessageMediaInstanceDAO.CreateMediaInstance(tx, req.Name, req.MediaType, params, req.Description, req.RateJSON, req.HashLife)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMessageMediaInstanceResponse{MessageMediaInstanceId: instanceId}, nil
}

// UpdateMessageMediaInstance 修改消息实例
func (this *MessageMediaInstanceService) UpdateMessageMediaInstance(ctx context.Context, req *pb.UpdateMessageMediaInstanceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	params := maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()
	err = models.SharedMessageMediaInstanceDAO.UpdateMediaInstance(tx, req.MessageMediaInstanceId, req.Name, req.MediaType, params, req.Description, req.RateJSON, req.HashLife, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteMessageMediaInstance 删除媒介实例
func (this *MessageMediaInstanceService) DeleteMessageMediaInstance(ctx context.Context, req *pb.DeleteMessageMediaInstanceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMessageMediaInstanceDAO.DisableMessageMediaInstance(tx, req.MessageMediaInstanceId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledMessageMediaInstances 计算媒介实例数量
func (this *MessageMediaInstanceService) CountAllEnabledMessageMediaInstances(ctx context.Context, req *pb.CountAllEnabledMessageMediaInstancesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedMessageMediaInstanceDAO.CountAllEnabledMediaInstances(tx, req.MediaType, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledMessageMediaInstances 列出单页媒介实例
func (this *MessageMediaInstanceService) ListEnabledMessageMediaInstances(ctx context.Context, req *pb.ListEnabledMessageMediaInstancesRequest) (*pb.ListEnabledMessageMediaInstancesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	instances, err := models.SharedMessageMediaInstanceDAO.ListAllEnabledMediaInstances(tx, req.MediaType, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbInstances := []*pb.MessageMediaInstance{}
	for _, instance := range instances {
		// 媒介
		media, err := models.SharedMessageMediaDAO.FindEnabledMediaWithType(tx, instance.MediaType)
		if err != nil {
			return nil, err
		}
		if media == nil {
			continue
		}
		pbMedia := &pb.MessageMedia{
			Id:              int64(media.Id),
			Type:            media.Type,
			Name:            media.Name,
			Description:     media.Description,
			UserDescription: media.UserDescription,
			IsOn:            media.IsOn == 1,
		}

		pbInstances = append(pbInstances, &pb.MessageMediaInstance{
			Id:           int64(instance.Id),
			Name:         instance.Name,
			IsOn:         instance.IsOn == 1,
			MessageMedia: pbMedia,
			ParamsJSON:   []byte(instance.Params),
			Description:  instance.Description,
			RateJSON:     []byte(instance.Rate),
		})
	}

	return &pb.ListEnabledMessageMediaInstancesResponse{MessageMediaInstances: pbInstances}, nil
}

// FindEnabledMessageMediaInstance 查找单个媒介实例信息
func (this *MessageMediaInstanceService) FindEnabledMessageMediaInstance(ctx context.Context, req *pb.FindEnabledMessageMediaInstanceRequest) (*pb.FindEnabledMessageMediaInstanceResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var cacheMap = maps.Map{}
	instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, req.MessageMediaInstanceId, cacheMap)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return &pb.FindEnabledMessageMediaInstanceResponse{MessageMediaInstance: nil}, nil
	}

	// 媒介
	media, err := models.SharedMessageMediaDAO.FindEnabledMediaWithType(tx, instance.MediaType)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return &pb.FindEnabledMessageMediaInstanceResponse{MessageMediaInstance: nil}, nil
	}
	pbMedia := &pb.MessageMedia{
		Id:              int64(media.Id),
		Type:            media.Type,
		Name:            media.Name,
		Description:     media.Description,
		UserDescription: media.UserDescription,
		IsOn:            media.IsOn == 1,
	}

	return &pb.FindEnabledMessageMediaInstanceResponse{MessageMediaInstance: &pb.MessageMediaInstance{
		Id:           int64(instance.Id),
		Name:         instance.Name,
		IsOn:         instance.IsOn == 1,
		MessageMedia: pbMedia,
		ParamsJSON:   []byte(instance.Params),
		Description:  instance.Description,
		RateJSON:     []byte(instance.Rate),
		HashLife:     types.Int32(instance.HashLife),
	}}, nil
}
