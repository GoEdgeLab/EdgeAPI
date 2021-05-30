// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NSRouteService 线路相关服务
type NSRouteService struct {
	services.BaseService
}

// CreateNSRoute 创建线路
func (this *NSRouteService) CreateNSRoute(ctx context.Context, req *pb.CreateNSRouteRequest) (*pb.CreateNSRouteResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	routeId, err := nameservers.SharedNSRouteDAO.CreateRoute(tx, req.NsClusterId, req.NsDomainId, req.UserId, req.Name, req.RangesJSON)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSRouteResponse{NsRouteId: routeId}, nil
}

// UpdateNSRoute 修改线路
func (this *NSRouteService) UpdateNSRoute(ctx context.Context, req *pb.UpdateNSRouteRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = nameservers.SharedNSRouteDAO.UpdateRoute(tx, req.NsRouteId, req.Name, req.RangesJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSRoute 删除线路
func (this *NSRouteService) DeleteNSRoute(ctx context.Context, req *pb.DeleteNSRouteRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = nameservers.SharedNSRouteDAO.DisableNSRoute(tx, req.NsRouteId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledNSRoute 获取单个路线信息
func (this *NSRouteService) FindEnabledNSRoute(ctx context.Context, req *pb.FindEnabledNSRouteRequest) (*pb.FindEnabledNSRouteResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	route, err := nameservers.SharedNSRouteDAO.FindEnabledNSRoute(tx, req.NsRouteId)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return &pb.FindEnabledNSRouteResponse{NsRoute: nil}, nil
	}

	// 集群
	var pbCluster *pb.NSCluster
	if route.ClusterId > 0 {
		cluster, err := nameservers.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(route.ClusterId))
		if err != nil {
			return nil, err
		}
		if cluster != nil {
			pbCluster = &pb.NSCluster{
				Id:   int64(cluster.Id),
				IsOn: cluster.IsOn == 1,
				Name: cluster.Name,
			}
		}
	}

	// 域名
	var pbDomain *pb.NSDomain
	if route.DomainId > 0 {
		domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, int64(route.DomainId))
		if err != nil {
			return nil, err
		}
		if domain != nil {
			pbDomain = &pb.NSDomain{
				Id:   int64(domain.Id),
				Name: domain.Name,
				IsOn: domain.IsOn == 1,
			}
		}
	}

	return &pb.FindEnabledNSRouteResponse{NsRoute: &pb.NSRoute{
		Id:         int64(route.Id),
		IsOn:       route.IsOn == 1,
		Name:       route.Name,
		RangesJSON: []byte(route.Ranges),
		NsCluster:  pbCluster,
		NsDomain:   pbDomain,
	}}, nil
}

// FindAllEnabledNSRoutes 读取所有线路
func (this *NSRouteService) FindAllEnabledNSRoutes(ctx context.Context, req *pb.FindAllEnabledNSRoutesRequest) (*pb.FindAllEnabledNSRoutesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	routes, err := nameservers.SharedNSRouteDAO.FindAllEnabledRoutes(tx, req.NsClusterId, req.NsDomainId, req.UserId)
	if err != nil {
		return nil, err
	}
	var pbRoutes = []*pb.NSRoute{}
	for _, route := range routes {
		// 集群
		var pbCluster *pb.NSCluster
		if route.ClusterId > 0 {
			cluster, err := nameservers.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(route.ClusterId))
			if err != nil {
				return nil, err
			}
			if cluster != nil {
				pbCluster = &pb.NSCluster{
					Id:   int64(cluster.Id),
					IsOn: cluster.IsOn == 1,
					Name: cluster.Name,
				}
			}
		}

		// 域名
		var pbDomain *pb.NSDomain
		if route.DomainId > 0 {
			domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, int64(route.DomainId))
			if err != nil {
				return nil, err
			}
			if domain != nil {
				pbDomain = &pb.NSDomain{
					Id:   int64(domain.Id),
					Name: domain.Name,
					IsOn: domain.IsOn == 1,
				}
			}
		}

		pbRoutes = append(pbRoutes, &pb.NSRoute{
			Id:         int64(route.Id),
			IsOn:       route.IsOn == 1,
			Name:       route.Name,
			RangesJSON: []byte(route.Ranges),
			NsCluster:  pbCluster,
			NsDomain:   pbDomain,
		})
	}
	return &pb.FindAllEnabledNSRoutesResponse{NsRoutes: pbRoutes}, nil
}

// UpdateNSRouteOrders 设置线路排序
func (this *NSRouteService) UpdateNSRouteOrders(ctx context.Context, req *pb.UpdateNSRouteOrdersRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = nameservers.SharedNSRouteDAO.UpdateRouteOrders(tx, req.NsRouteIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
