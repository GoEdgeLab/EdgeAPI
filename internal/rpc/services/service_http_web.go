package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPWebService struct {
	BaseService
}

// 创建Web配置
func (this *HTTPWebService) CreateHTTPWeb(ctx context.Context, req *pb.CreateHTTPWebRequest) (*pb.CreateHTTPWebResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	webId, err := models.SharedHTTPWebDAO.CreateWeb(req.RootJSON)
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
	return &pb.FindEnabledHTTPWebResponse{Web: result}, nil
}

// 查找Web配置
func (this *HTTPWebService) FindEnabledHTTPWebConfig(ctx context.Context, req *pb.FindEnabledHTTPWebConfigRequest) (*pb.FindEnabledHTTPWebConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(req.WebId)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledHTTPWebConfigResponse{WebJSON: configJSON}, nil
}

// 修改Web配置
func (this *HTTPWebService) UpdateHTTPWeb(ctx context.Context, req *pb.UpdateHTTPWebRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWeb(req.WebId, req.RootJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改Gzip配置
func (this *HTTPWebService) UpdateHTTPWebGzip(ctx context.Context, req *pb.UpdateHTTPWebGzipRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebGzip(req.WebId, req.GzipJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改字符集配置
func (this *HTTPWebService) UpdateHTTPWebCharset(ctx context.Context, req *pb.UpdateHTTPWebCharsetRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebCharset(req.WebId, req.CharsetJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 更改请求Header策略
func (this *HTTPWebService) UpdateHTTPWebRequestHeader(ctx context.Context, req *pb.UpdateHTTPWebRequestHeaderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebRequestHeaderPolicy(req.WebId, req.HeaderJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 更改响应Header策略
func (this *HTTPWebService) UpdateHTTPWebResponseHeader(ctx context.Context, req *pb.UpdateHTTPWebResponseHeaderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebResponseHeaderPolicy(req.WebId, req.HeaderJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 更改Shutdown
func (this *HTTPWebService) UpdateHTTPWebShutdown(ctx context.Context, req *pb.UpdateHTTPWebShutdownRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebShutdown(req.WebId, req.ShutdownJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 更改Pages
func (this *HTTPWebService) UpdateHTTPWebPages(ctx context.Context, req *pb.UpdateHTTPWebPagesRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebPages(req.WebId, req.PagesJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 更改访问日志配置
func (this *HTTPWebService) UpdateHTTPWebAccessLog(ctx context.Context, req *pb.UpdateHTTPWebAccessLogRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebAccessLogConfig(req.WebId, req.AccessLogJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 更改统计配置
func (this *HTTPWebService) UpdateHTTPWebStat(ctx context.Context, req *pb.UpdateHTTPWebStatRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebStat(req.WebId, req.StatJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 更改缓存配置
func (this *HTTPWebService) UpdateHTTPWebCache(ctx context.Context, req *pb.UpdateHTTPWebCacheRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebCache(req.WebId, req.CacheJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 更改防火墙设置
func (this *HTTPWebService) UpdateHTTPWebFirewall(ctx context.Context, req *pb.UpdateHTTPWebFirewallRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebFirewall(req.WebId, req.FirewallJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 更改路径规则设置
func (this *HTTPWebService) UpdateHTTPWebLocations(ctx context.Context, req *pb.UpdateHTTPWebLocationsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebLocations(req.WebId, req.LocationsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 更改跳转到HTTPS设置
func (this *HTTPWebService) UpdateHTTPWebRedirectToHTTPS(ctx context.Context, req *pb.UpdateHTTPWebRedirectToHTTPSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebRedirectToHTTPS(req.WebId, req.RedirectToHTTPSJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 更改Websocket设置
func (this *HTTPWebService) UpdateHTTPWebWebsocket(ctx context.Context, req *pb.UpdateHTTPWebWebsocketRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebsocket(req.WebId, req.WebsocketJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 更改重写规则设置
func (this *HTTPWebService) UpdateHTTPWebRewriteRules(ctx context.Context, req *pb.UpdateHTTPWebRewriteRulesRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebDAO.UpdateWebRewriteRules(req.WebId, req.RewriteRulesJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
