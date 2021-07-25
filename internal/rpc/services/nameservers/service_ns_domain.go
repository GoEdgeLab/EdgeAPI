// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NSDomainService 域名相关服务
type NSDomainService struct {
	services.BaseService
}

// CreateNSDomain 创建域名
func (this *NSDomainService) CreateNSDomain(ctx context.Context, req *pb.CreateNSDomainRequest) (*pb.CreateNSDomainResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	domainId, err := nameservers.SharedNSDomainDAO.CreateDomain(tx, req.NsClusterId, req.UserId, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSDomainResponse{NsDomainId: domainId}, nil
}

// UpdateNSDomain 修改域名
func (this *NSDomainService) UpdateNSDomain(ctx context.Context, req *pb.UpdateNSDomainRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSDomainDAO.UpdateDomain(tx, req.NsDomainId, req.NsClusterId, req.UserId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSDomain 删除域名
func (this *NSDomainService) DeleteNSDomain(ctx context.Context, req *pb.DeleteNSDomainRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSDomainDAO.DisableNSDomain(tx, req.NsDomainId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledNSDomain 查找单个域名
func (this *NSDomainService) FindEnabledNSDomain(ctx context.Context, req *pb.FindEnabledNSDomainRequest) (*pb.FindEnabledNSDomainResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	domain, err := nameservers.SharedNSDomainDAO.FindEnabledNSDomain(tx, req.NsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindEnabledNSDomainResponse{NsDomain: nil}, nil
	}

	// 集群
	cluster, err := nameservers.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(domain.ClusterId))
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return &pb.FindEnabledNSDomainResponse{NsDomain: nil}, nil
	}

	// 用户
	var pbUser *pb.User
	if domain.UserId > 0 {
		user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(domain.UserId))
		if err != nil {
			return nil, err
		}
		if user == nil {
			return &pb.FindEnabledNSDomainResponse{NsDomain: nil}, nil
		}
		pbUser = &pb.User{
			Id:       int64(user.Id),
			Username: user.Username,
			Fullname: user.Fullname,
		}
	}

	return &pb.FindEnabledNSDomainResponse{
		NsDomain: &pb.NSDomain{
			Id:        int64(domain.Id),
			Name:      domain.Name,
			IsOn:      domain.IsOn == 1,
			TsigJSON:  []byte(domain.Tsig),
			CreatedAt: int64(domain.CreatedAt),
			NsCluster: &pb.NSCluster{
				Id:   int64(cluster.Id),
				IsOn: cluster.IsOn == 1,
				Name: cluster.Name,
			},
			User: pbUser,
		},
	}, nil
}

// CountAllEnabledNSDomains 计算域名数量
func (this *NSDomainService) CountAllEnabledNSDomains(ctx context.Context, req *pb.CountAllEnabledNSDomainsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := nameservers.SharedNSDomainDAO.CountAllEnabledDomains(tx, req.NsClusterId, req.UserId, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledNSDomains 列出单页域名
func (this *NSDomainService) ListEnabledNSDomains(ctx context.Context, req *pb.ListEnabledNSDomainsRequest) (*pb.ListEnabledNSDomainsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	domains, err := nameservers.SharedNSDomainDAO.ListEnabledDomains(tx, req.NsClusterId, req.UserId, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbDomains := []*pb.NSDomain{}
	for _, domain := range domains {
		// 集群
		cluster, err := nameservers.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(domain.ClusterId))
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			continue
		}

		// 用户
		var pbUser *pb.User
		if domain.UserId > 0 {
			user, err := models.SharedUserDAO.FindEnabledUser(tx, int64(domain.UserId))
			if err != nil {
				return nil, err
			}
			if user == nil {
				continue
			}
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}

		pbDomains = append(pbDomains, &pb.NSDomain{
			Id:        int64(domain.Id),
			Name:      domain.Name,
			IsOn:      domain.IsOn == 1,
			CreatedAt: int64(domain.CreatedAt),
			TsigJSON:  []byte(domain.Tsig),
			NsCluster: &pb.NSCluster{
				Id:   int64(cluster.Id),
				IsOn: cluster.IsOn == 1,
				Name: cluster.Name,
			},
			User: pbUser,
		})
	}

	return &pb.ListEnabledNSDomainsResponse{NsDomains: pbDomains}, nil
}

// ListNSDomainsAfterVersion 根据版本列出一组域名
func (this *NSDomainService) ListNSDomainsAfterVersion(ctx context.Context, req *pb.ListNSDomainsAfterVersionRequest) (*pb.ListNSDomainsAfterVersionResponse, error) {
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	// 集群ID
	var tx = this.NullTx()
	if req.Size <= 0 {
		req.Size = 2000
	}
	domains, err := nameservers.SharedNSDomainDAO.ListDomainsAfterVersion(tx, req.Version, req.Size)
	if err != nil {
		return nil, err
	}

	var pbDomains []*pb.NSDomain
	for _, domain := range domains {
		pbDomains = append(pbDomains, &pb.NSDomain{
			Id:        int64(domain.Id),
			Name:      domain.Name,
			IsOn:      domain.IsOn == 1,
			IsDeleted: domain.State == nameservers.NSDomainStateDisabled,
			Version:   int64(domain.Version),
			TsigJSON:  []byte(domain.Tsig),
			NsCluster: &pb.NSCluster{Id: int64(domain.ClusterId)},
			User:      nil,
		})
	}
	return &pb.ListNSDomainsAfterVersionResponse{NsDomains: pbDomains}, nil
}

// FindEnabledNSDomainTSIG 查找TSIG配置
func (this *NSDomainService) FindEnabledNSDomainTSIG(ctx context.Context, req *pb.FindEnabledNSDomainTSIGRequest) (*pb.FindEnabledNSDomainTSIGResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	tsig, err := nameservers.SharedNSDomainDAO.FindEnabledDomainTSIG(tx, req.NsDomainId)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledNSDomainTSIGResponse{TsigJSON: tsig}, nil
}

// UpdateNSDomainTSIG 修改TSIG配置
func (this *NSDomainService) UpdateNSDomainTSIG(ctx context.Context, req *pb.UpdateNSDomainTSIGRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSDomainDAO.UpdateDomainTSIG(tx, req.NsDomainId, req.TsigJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
