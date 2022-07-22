package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns/dnsutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// DNSService DNS相关服务
type DNSService struct {
	BaseService
}

// FindAllDNSIssues 查找问题
func (this *DNSService) FindAllDNSIssues(ctx context.Context, req *pb.FindAllDNSIssuesRequest) (*pb.FindAllDNSIssuesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var result = []*pb.DNSIssue{}
	var tx = this.NullTx()
	var clusters []*models.NodeCluster

	if req.NodeClusterId <= 0 {
		clusters, err = models.SharedNodeClusterDAO.FindAllEnabledClustersHaveDNSDomain(tx)
		if err != nil {
			return nil, err
		}
	} else {
		cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, req.NodeClusterId)
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			return &pb.FindAllDNSIssuesResponse{Issues: nil}, nil
		}
		clusters = []*models.NodeCluster{cluster}
	}
	for _, cluster := range clusters {
		issues, err := dnsutils.CheckClusterDNS(tx, cluster)
		if err != nil {
			return nil, err
		}
		if len(issues) > 0 {
			result = append(result, issues...)
		}
	}

	return &pb.FindAllDNSIssuesResponse{Issues: result}, nil
}
