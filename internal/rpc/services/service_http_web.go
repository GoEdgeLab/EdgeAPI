package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
)

type HTTPWebService struct {
	BaseService
}

// CreateHTTPWeb 创建Web配置
func (this *HTTPWebService) CreateHTTPWeb(ctx context.Context, req *pb.CreateHTTPWebRequest) (*pb.CreateHTTPWebResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	webId, err := models.SharedHTTPWebDAO.CreateWeb(tx, adminId, userId, req.RootJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPWebResponse{WebId: webId}, nil
}

// FindEnabledHTTPWeb 查找Web配置
func (this *HTTPWebService) FindEnabledHTTPWeb(ctx context.Context, req *pb.FindEnabledHTTPWebRequest) (*pb.FindEnabledHTTPWebResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	web, err := models.SharedHTTPWebDAO.FindEnabledHTTPWeb(tx, req.WebId)
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

// FindEnabledHTTPWebConfig 查找Web配置
func (this *HTTPWebService) FindEnabledHTTPWebConfig(ctx context.Context, req *pb.FindEnabledHTTPWebConfigRequest) (*pb.FindEnabledHTTPWebConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, req.WebId)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledHTTPWebConfigResponse{WebJSON: configJSON}, nil
}

// UpdateHTTPWeb 修改Web配置
func (this *HTTPWebService) UpdateHTTPWeb(ctx context.Context, req *pb.UpdateHTTPWebRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWeb(tx, req.WebId, req.RootJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPWebGzip 修改Gzip配置
func (this *HTTPWebService) UpdateHTTPWebGzip(ctx context.Context, req *pb.UpdateHTTPWebGzipRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebGzip(tx, req.WebId, req.GzipJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPWebCharset 修改字符集配置
func (this *HTTPWebService) UpdateHTTPWebCharset(ctx context.Context, req *pb.UpdateHTTPWebCharsetRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebCharset(tx, req.WebId, req.CharsetJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebRequestHeader 更改请求Header策略
func (this *HTTPWebService) UpdateHTTPWebRequestHeader(ctx context.Context, req *pb.UpdateHTTPWebRequestHeaderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebRequestHeaderPolicy(tx, req.WebId, req.HeaderJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPWebResponseHeader 更改响应Header策略
func (this *HTTPWebService) UpdateHTTPWebResponseHeader(ctx context.Context, req *pb.UpdateHTTPWebResponseHeaderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebResponseHeaderPolicy(tx, req.WebId, req.HeaderJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPWebShutdown 更改Shutdown
func (this *HTTPWebService) UpdateHTTPWebShutdown(ctx context.Context, req *pb.UpdateHTTPWebShutdownRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebShutdown(tx, req.WebId, req.ShutdownJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebPages 更改Pages
func (this *HTTPWebService) UpdateHTTPWebPages(ctx context.Context, req *pb.UpdateHTTPWebPagesRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebPages(tx, req.WebId, req.PagesJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebAccessLog 更改访问日志配置
func (this *HTTPWebService) UpdateHTTPWebAccessLog(ctx context.Context, req *pb.UpdateHTTPWebAccessLogRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebAccessLogConfig(tx, req.WebId, req.AccessLogJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebStat 更改统计配置
func (this *HTTPWebService) UpdateHTTPWebStat(ctx context.Context, req *pb.UpdateHTTPWebStatRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebStat(tx, req.WebId, req.StatJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebCache 更改缓存配置
func (this *HTTPWebService) UpdateHTTPWebCache(ctx context.Context, req *pb.UpdateHTTPWebCacheRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebCache(tx, req.WebId, req.CacheJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPWebFirewall 更改防火墙设置
func (this *HTTPWebService) UpdateHTTPWebFirewall(ctx context.Context, req *pb.UpdateHTTPWebFirewallRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebFirewall(tx, req.WebId, req.FirewallJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPWebLocations 更改路径规则设置
func (this *HTTPWebService) UpdateHTTPWebLocations(ctx context.Context, req *pb.UpdateHTTPWebLocationsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 检查用户权限
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebLocations(tx, req.WebId, req.LocationsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPWebRedirectToHTTPS 更改跳转到HTTPS设置
func (this *HTTPWebService) UpdateHTTPWebRedirectToHTTPS(ctx context.Context, req *pb.UpdateHTTPWebRedirectToHTTPSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebRedirectToHTTPS(tx, req.WebId, req.RedirectToHTTPSJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebWebsocket 更改Websocket设置
func (this *HTTPWebService) UpdateHTTPWebWebsocket(ctx context.Context, req *pb.UpdateHTTPWebWebsocketRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebsocket(tx, req.WebId, req.WebsocketJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebFastcgi 更改Fastcgi设置
func (this *HTTPWebService) UpdateHTTPWebFastcgi(ctx context.Context, req *pb.UpdateHTTPWebFastcgiRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebFastcgi(tx, req.WebId, req.FastcgiJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebRewriteRules 更改重写规则设置
func (this *HTTPWebService) UpdateHTTPWebRewriteRules(ctx context.Context, req *pb.UpdateHTTPWebRewriteRulesRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPWebDAO.UpdateWebRewriteRules(tx, req.WebId, req.RewriteRulesJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateHTTPWebHostRedirects 更改主机跳转设置
func (this *HTTPWebService) UpdateHTTPWebHostRedirects(ctx context.Context, req *pb.UpdateHTTPWebHostRedirectsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	hostRedirects := []*serverconfigs.HTTPHostRedirectConfig{}
	if len(req.HostRedirectsJSON) == 0 {
		return nil, errors.New("'hostRedirectsJSON' should not be empty")
	}
	err = json.Unmarshal(req.HostRedirectsJSON, &hostRedirects)
	if err != nil {
		return nil, err
	}

	// 校验
	for _, redirect := range hostRedirects {
		err := redirect.Init()
		if err != nil {
			return nil, err
		}
	}

	var tx *dbs.Tx
	err = models.SharedHTTPWebDAO.UpdateWebHostRedirects(tx, req.WebId, hostRedirects)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindHTTPWebHostRedirects 查找主机跳转设置
func (this *HTTPWebService) FindHTTPWebHostRedirects(ctx context.Context, req *pb.FindHTTPWebHostRedirectsRequest) (*pb.FindHTTPWebHostRedirectsResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPWebDAO.CheckUserWeb(nil, userId, req.WebId)
		if err != nil {
			return nil, err
		}
	}

	var tx *dbs.Tx
	redirectsJSON, err := models.SharedHTTPWebDAO.FindWebHostRedirects(tx, req.WebId)
	if err != nil {
		return nil, err
	}
	return &pb.FindHTTPWebHostRedirectsResponse{HostRedirectsJSON: redirectsJSON}, nil
}
