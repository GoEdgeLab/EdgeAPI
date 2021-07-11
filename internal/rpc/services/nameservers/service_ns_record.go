// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// NSRecordService 域名记录相关服务
type NSRecordService struct {
	services.BaseService
}

// CreateNSRecord 创建记录
func (this *NSRecordService) CreateNSRecord(ctx context.Context, req *pb.CreateNSRecordRequest) (*pb.CreateNSRecordResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	recordId, err := nameservers.SharedNSRecordDAO.CreateRecord(tx, req.NsDomainId, req.Description, req.Name, req.Type, req.Value, req.Ttl, req.NsRouteIds)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSRecordResponse{NsRecordId: recordId}, nil
}

// UpdateNSRecord 修改记录
func (this *NSRecordService) UpdateNSRecord(ctx context.Context, req *pb.UpdateNSRecordRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSRecordDAO.UpdateRecord(tx, req.NsRecordId, req.Description, req.Name, req.Type, req.Value, req.Ttl, req.NsRouteIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSRecord 删除记录
func (this *NSRecordService) DeleteNSRecord(ctx context.Context, req *pb.DeleteNSRecordRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSRecordDAO.DisableNSRecord(tx, req.NsRecordId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledNSRecords 计算记录数量
func (this *NSRecordService) CountAllEnabledNSRecords(ctx context.Context, req *pb.CountAllEnabledNSRecordsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := nameservers.SharedNSRecordDAO.CountAllEnabledDomainRecords(tx, req.NsDomainId, req.Type, req.Keyword, req.NsRouteId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledNSRecords 读取单页记录
func (this *NSRecordService) ListEnabledNSRecords(ctx context.Context, req *pb.ListEnabledNSRecordsRequest) (*pb.ListEnabledNSRecordsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	records, err := nameservers.SharedNSRecordDAO.ListEnabledRecords(tx, req.NsDomainId, req.Type, req.Keyword, req.NsRouteId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbRecords = []*pb.NSRecord{}
	for _, record := range records {
		// 线路
		var pbRoutes = []*pb.NSRoute{}
		for _, recordId := range record.DecodeRouteIds() {
			route, err := nameservers.SharedNSRouteDAO.FindEnabledNSRoute(tx, recordId)
			if err != nil {
				return nil, err
			}
			if route == nil {
				continue
			}
			pbRoutes = append(pbRoutes, &pb.NSRoute{
				Id:   int64(route.Id),
				Name: route.Name,
			})
		}

		pbRecords = append(pbRecords, &pb.NSRecord{
			Id:          int64(record.Id),
			Description: record.Description,
			Name:        record.Name,
			Type:        record.Type,
			Value:       record.Value,
			Ttl:         types.Int32(record.Ttl),
			Weight:      types.Int32(record.Weight),
			CreatedAt:   int64(record.CreatedAt),
			IsOn:        record.IsOn == 1,
			NsDomain:    nil,
			NsRoutes:    pbRoutes,
		})
	}
	return &pb.ListEnabledNSRecordsResponse{NsRecords: pbRecords}, nil
}

// FindEnabledNSRecord 查询单个记录信息
func (this *NSRecordService) FindEnabledNSRecord(ctx context.Context, req *pb.FindEnabledNSRecordRequest) (*pb.FindEnabledNSRecordResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	record, err := nameservers.SharedNSRecordDAO.FindEnabledNSRecord(tx, req.NsRecordId)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &pb.FindEnabledNSRecordResponse{NsRecord: nil}, nil
	}

	// 域名
	domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, int64(record.DomainId))
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindEnabledNSRecordResponse{NsRecord: nil}, nil
	}
	var pbDomain = &pb.NSDomain{
		Id:   int64(domain.Id),
		Name: domain.Name,
		IsOn: domain.IsOn == 1,
	}

	// 线路
	var pbRoutes = []*pb.NSRoute{}
	for _, recordId := range record.DecodeRouteIds() {
		route, err := nameservers.SharedNSRouteDAO.FindEnabledNSRoute(tx, recordId)
		if err != nil {
			return nil, err
		}
		if route == nil {
			continue
		}
		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Id:   int64(route.Id),
			Name: route.Name,
		})
	}

	return &pb.FindEnabledNSRecordResponse{NsRecord: &pb.NSRecord{
		Id:          int64(record.Id),
		Description: record.Description,
		Name:        record.Name,
		Type:        record.Type,
		Value:       record.Value,
		Ttl:         types.Int32(record.Ttl),
		Weight:      types.Int32(record.Weight),
		CreatedAt:   int64(record.CreatedAt),
		IsOn:        record.IsOn == 1,
		NsDomain:    pbDomain,
		NsRoutes:    pbRoutes,
	}}, nil
}

// ListNSRecordsAfterVersion 根据版本列出一组记录
func (this *NSRecordService) ListNSRecordsAfterVersion(ctx context.Context, req *pb.ListNSRecordsAfterVersionRequest) (*pb.ListNSRecordsAfterVersionResponse, error) {
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	// 集群ID
	var tx = this.NullTx()
	records, err := nameservers.SharedNSRecordDAO.ListRecordsAfterVersion(tx, req.Version, 2000)
	if err != nil {
		return nil, err
	}

	var pbRecords []*pb.NSRecord
	for _, record := range records {
		// 线路
		pbRoutes := []*pb.NSRoute{}
		routeIds := record.DecodeRouteIds()
		for _, routeId := range routeIds {
			pbRoutes = append(pbRoutes, &pb.NSRoute{Id: routeId})
		}

		pbRecords = append(pbRecords, &pb.NSRecord{
			Id:          int64(record.Id),
			Description: "",
			Name:        record.Name,
			Type:        record.Type,
			Value:       record.Value,
			Ttl:         types.Int32(record.Ttl),
			Weight:      types.Int32(record.Weight),
			IsDeleted:   record.State == nameservers.NSRecordStateDisabled,
			IsOn:        record.IsOn == 1,
			Version:     int64(record.Version),
			NsDomain:    &pb.NSDomain{Id: int64(record.DomainId)},
			NsRoutes:    pbRoutes,
		})
	}
	return &pb.ListNSRecordsAfterVersionResponse{NsRecords: pbRecords}, nil
}
