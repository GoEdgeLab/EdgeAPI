package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// DNS相关服务
type DNSService struct {
	BaseService
}

// 查找问题
func (this *DNSService) FindAllDNSIssues(ctx context.Context, req *pb.FindAllDNSIssuesRequest) (*pb.FindAllDNSIssuesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	result := []*pb.DNSIssue{}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersHaveDNSDomain(tx)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		issues, err := models.SharedNodeClusterDAO.CheckClusterDNS(tx, cluster)
		if err != nil {
			return nil, err
		}
		if len(issues) > 0 {
			result = append(result, issues...)
		}
	}

	return &pb.FindAllDNSIssuesResponse{Issues: result}, nil
}
