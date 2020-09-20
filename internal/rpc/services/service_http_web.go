package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPWebService struct {
}

// 创建Web配置
func (this *HTTPWebService) CreateHTTPWeb(ctx context.Context, req *pb.CreateHTTPWebRequest) (*pb.CreateHTTPWebResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	webId, err := models.SharedHTTPWebDAO.CreateWeb(req.Root)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPWebResponse{WebId: webId}, nil
}

// 查找Web配置
func (this *HTTPWebService) FindEnabledHTTPWeb(ctx context.Context, req *pb.FindEnabledHTTPWebRequest) (*pb.FindEnabledHTTPWebResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	web, err := models.SharedHTTPWebDAO.FindEnabledHTTPWeb(req.WebId)
	if err != nil {
		return nil, err
	}

	if web == nil {
		return &pb.FindEnabledHTTPWebResponse{Web: nil}, nil
	}

	result := &pb.HTTPWeb{}
	result.Id = int64(web.Id)
	result.IsOn = web.IsOn == 1
	result.Root = web.Root
	result.Charset = web.Charset
	result.RequestHeaderPolicyId = int64(web.RequestHeaderPolicyId)
	result.ResponseHeaderPolicyId = int64(web.ResponseHeaderPolicyId)
	return &pb.FindEnabledHTTPWebResponse{Web: result}, nil
}

// 修改Web配置
func (this *HTTPWebService) UpdateHTTPWeb(ctx context.Context, req *pb.UpdateHTTPWebRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWeb(req.WebId, req.Root)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改Gzip配置
func (this *HTTPWebService) UpdateHTTPWebGzip(ctx context.Context, req *pb.UpdateHTTPWebGzipRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebGzip(req.WebId, req.GzipJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改字符集配置
func (this *HTTPWebService) UpdateHTTPWebCharset(ctx context.Context, req *pb.UpdateHTTPWebCharsetRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebCharset(req.WebId, req.Charset)
	if err != nil {
		return nil, err
	}
	return &pb.RPCUpdateSuccess{}, nil
}

// 更改请求Header策略
func (this *HTTPWebService) UpdateHTTPWebRequestHeaderPolicy(ctx context.Context, req *pb.UpdateHTTPWebRequestHeaderPolicyRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebRequestHeaderPolicy(req.WebId, req.HeaderPolicyId)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 更改响应Header策略
func (this *HTTPWebService) UpdateHTTPWebResponseHeaderPolicy(ctx context.Context, req *pb.UpdateHTTPWebResponseHeaderPolicyRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebResponseHeaderPolicy(req.WebId, req.HeaderPolicyId)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 更改Shutdown
func (this *HTTPWebService) UpdateHTTPWebShutdown(ctx context.Context, req *pb.UpdateHTTPWebShutdownRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebShutdown(req.WebId, req.ShutdownJSON)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 更改Pages
func (this *HTTPWebService) UpdateHTTPWebPages(ctx context.Context, req *pb.UpdateHTTPWebPagesRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebPages(req.WebId, req.PagesJSON)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 更改访问日志配置
func (this *HTTPWebService) UpdateHTTPAccessLog(ctx context.Context, req *pb.UpdateHTTPAccessLogRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebAccessLogConfig(req.WebId, req.AccessLogJSON)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 更改统计配置
func (this *HTTPWebService) UpdateHTTPStat(ctx context.Context, req *pb.UpdateHTTPStatRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebStat(req.WebId, req.StatJSON)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 更改缓存配置
func (this *HTTPWebService) UpdateHTTPCache(ctx context.Context, req *pb.UpdateHTTPCacheRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebCache(req.WebId, req.CacheJSON)
	if err != nil {
		return nil, err
	}

	return rpcutils.RPCUpdateSuccess()
}
