// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nodes

import (
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/clients"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/users"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc"
	"reflect"
	"strings"
)

// 注册服务
func (this *APINode) registerServices(server *grpc.Server) {
	{
		var instance = this.serviceInstance(&services.PingService{}).(*services.PingService)
		pb.RegisterPingServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.APITokenService{}).(*services.APITokenService)
		pb.RegisterAPITokenServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.AdminService{}).(*services.AdminService)
		pb.RegisterAdminServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeGrantService{}).(*services.NodeGrantService)
		pb.RegisterNodeGrantServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerService{}).(*services.ServerService)
		pb.RegisterServerServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeService{}).(*services.NodeService)
		pb.RegisterNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeClusterService{}).(*services.NodeClusterService)
		pb.RegisterNodeClusterServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeIPAddressService{}).(*services.NodeIPAddressService)
		pb.RegisterNodeIPAddressServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeIPAddressLogService{}).(*services.NodeIPAddressLogService)
		pb.RegisterNodeIPAddressLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeIPAddressThresholdService{}).(*services.NodeIPAddressThresholdService)
		pb.RegisterNodeIPAddressThresholdServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.APINodeService{}).(*services.APINodeService)
		pb.RegisterAPINodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.APIMethodStatService{}).(*services.APIMethodStatService)
		pb.RegisterAPIMethodStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.OriginService{}).(*services.OriginService)
		pb.RegisterOriginServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPWebService{}).(*services.HTTPWebService)
		pb.RegisterHTTPWebServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ReverseProxyService{}).(*services.ReverseProxyService)
		pb.RegisterReverseProxyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPGzipService{}).(*services.HTTPGzipService)
		pb.RegisterHTTPGzipServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPHeaderPolicyService{}).(*services.HTTPHeaderPolicyService)
		pb.RegisterHTTPHeaderPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPHeaderService{}).(*services.HTTPHeaderService)
		pb.RegisterHTTPHeaderServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPPageService{}).(*services.HTTPPageService)
		pb.RegisterHTTPPageServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPCachePolicyService{}).(*services.HTTPCachePolicyService)
		pb.RegisterHTTPCachePolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPFirewallPolicyService{}).(*services.HTTPFirewallPolicyService)
		pb.RegisterHTTPFirewallPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.FirewallService{}).(*services.FirewallService)
		pb.RegisterFirewallServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPLocationService{}).(*services.HTTPLocationService)
		pb.RegisterHTTPLocationServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPWebsocketService{}).(*services.HTTPWebsocketService)
		pb.RegisterHTTPWebsocketServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPRewriteRuleService{}).(*services.HTTPRewriteRuleService)
		pb.RegisterHTTPRewriteRuleServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.SSLCertService{}).(*services.SSLCertService)
		pb.RegisterSSLCertServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.SSLPolicyService{}).(*services.SSLPolicyService)
		pb.RegisterSSLPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.SysSettingService{}).(*services.SysSettingService)
		pb.RegisterSysSettingServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPFirewallRuleGroupService{}).(*services.HTTPFirewallRuleGroupService)
		pb.RegisterHTTPFirewallRuleGroupServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPFirewallRuleSetService{}).(*services.HTTPFirewallRuleSetService)
		pb.RegisterHTTPFirewallRuleSetServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.DBNodeService{}).(*services.DBNodeService)
		pb.RegisterDBNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeLogService{}).(*services.NodeLogService)
		pb.RegisterNodeLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeLoginService{}).(*services.NodeLoginService)
		pb.RegisterNodeLoginServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPAccessLogService{}).(*services.HTTPAccessLogService)
		pb.RegisterHTTPAccessLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.MessageService{}).(*services.MessageService)
		pb.RegisterMessageServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeGroupService{}).(*services.NodeGroupService)
		pb.RegisterNodeGroupServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeRegionService{}).(*services.NodeRegionService)
		pb.RegisterNodeRegionServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerGroupService{}).(*services.ServerGroupService)
		pb.RegisterServerGroupServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.IPLibraryService{}).(*services.IPLibraryService)
		pb.RegisterIPLibraryServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.IPLibraryFileService{}).(*services.IPLibraryFileService)
		pb.RegisterIPLibraryFileServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.IPLibraryArtifactService{}).(*services.IPLibraryArtifactService)
		pb.RegisterIPLibraryArtifactServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.FileChunkService{}).(*services.FileChunkService)
		pb.RegisterFileChunkServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.FileService{}).(*services.FileService)
		pb.RegisterFileServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.RegionCountryService{}).(*services.RegionCountryService)
		pb.RegisterRegionCountryServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.RegionProvinceService{}).(*services.RegionProvinceService)
		pb.RegisterRegionProvinceServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.RegionCityService{}).(*services.RegionCityService)
		pb.RegisterRegionCityServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.RegionTownService{}).(*services.RegionTownService)
		pb.RegisterRegionTownServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.RegionProviderService{}).(*services.RegionProviderService)
		pb.RegisterRegionProviderServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.IPListService{}).(*services.IPListService)
		pb.RegisterIPListServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.IPItemService{}).(*services.IPItemService)
		pb.RegisterIPItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.LogService{}).(*services.LogService)
		pb.RegisterLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.DNSProviderService{}).(*services.DNSProviderService)
		pb.RegisterDNSProviderServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.DNSDomainService{}).(*services.DNSDomainService)
		pb.RegisterDNSDomainServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.DNSService{}).(*services.DNSService)
		pb.RegisterDNSServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ACMEUserService{}).(*services.ACMEUserService)
		pb.RegisterACMEUserServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ACMETaskService{}).(*services.ACMETaskService)
		pb.RegisterACMETaskServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ACMEAuthenticationService{}).(*services.ACMEAuthenticationService)
		pb.RegisterACMEAuthenticationServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ACMEProviderService{}).(*services.ACMEProviderService)
		pb.RegisterACMEProviderServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ACMEProviderAccountService{}).(*services.ACMEProviderAccountService)
		pb.RegisterACMEProviderAccountServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&users.UserService{}).(*users.UserService)
		pb.RegisterUserServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.UserIdentityService{}).(*services.UserIdentityService)
		pb.RegisterUserIdentityServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerDailyStatService{}).(*services.ServerDailyStatService)
		pb.RegisterServerDailyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.LoginService{}).(*services.LoginService)
		pb.RegisterLoginServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.LoginSessionService{}).(*services.LoginSessionService)
		pb.RegisterLoginSessionServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.UserAccessKeyService{}).(*services.UserAccessKeyService)
		pb.RegisterUserAccessKeyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.SysLockerService{}).(*services.SysLockerService)
		pb.RegisterSysLockerServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeTaskService{}).(*services.NodeTaskService)
		pb.RegisterNodeTaskServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeValueService{}).(*services.NodeValueService)
		pb.RegisterNodeValueServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.DBService{}).(*services.DBService)
		pb.RegisterDBServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerRegionCityMonthlyStatService{}).(*services.ServerRegionCityMonthlyStatService)
		pb.RegisterServerRegionCityMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerRegionCountryMonthlyStatService{}).(*services.ServerRegionCountryMonthlyStatService)
		pb.RegisterServerRegionCountryMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerRegionProvinceMonthlyStatService{}).(*services.ServerRegionProvinceMonthlyStatService)
		pb.RegisterServerRegionProvinceMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerRegionProviderMonthlyStatService{}).(*services.ServerRegionProviderMonthlyStatService)
		pb.RegisterServerRegionProviderMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&clients.FormalClientSystemService{}).(*clients.FormalClientSystemService)
		pb.RegisterFormalClientSystemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&clients.FormalClientBrowserService{}).(*clients.FormalClientBrowserService)
		pb.RegisterFormalClientBrowserServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&clients.ClientAgentIPService{}).(*clients.ClientAgentIPService)
		pb.RegisterClientAgentIPServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&clients.ClientAgentService{}).(*clients.ClientAgentService)
		pb.RegisterClientAgentServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerClientSystemMonthlyStatService{}).(*services.ServerClientSystemMonthlyStatService)
		pb.RegisterServerClientSystemMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerClientBrowserMonthlyStatService{}).(*services.ServerClientBrowserMonthlyStatService)
		pb.RegisterServerClientBrowserMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.ServerHTTPFirewallDailyStatService{}).(*services.ServerHTTPFirewallDailyStatService)
		pb.RegisterServerHTTPFirewallDailyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.DNSTaskService{}).(*services.DNSTaskService)
		pb.RegisterDNSTaskServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeClusterFirewallActionService{}).(*services.NodeClusterFirewallActionService)
		pb.RegisterNodeClusterFirewallActionServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.AuthorityNodeService{}).(*services.AuthorityNodeService)
		pb.RegisterAuthorityNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.LatestItemService{}).(*services.LatestItemService)
		pb.RegisterLatestItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeThresholdService{}).(*services.NodeThresholdService)
		pb.RegisterNodeThresholdServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPFastcgiService{}).(*services.HTTPFastcgiService)
		pb.RegisterHTTPFastcgiServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.HTTPAuthPolicyService{}).(*services.HTTPAuthPolicyService)
		pb.RegisterHTTPAuthPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.MetricItemService{}).(*services.MetricItemService)
		pb.RegisterMetricItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.NodeClusterMetricItemService{}).(*services.NodeClusterMetricItemService)
		pb.RegisterNodeClusterMetricItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.MetricStatService{}).(*services.MetricStatService)
		pb.RegisterMetricStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		var instance = this.serviceInstance(&services.MetricChartService{}).(*services.MetricChartService)
		pb.RegisterMetricChartServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.ServerStatBoardService{}).(*services.ServerStatBoardService)
		pb.RegisterServerStatBoardServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.ServerStatBoardChartService{}).(*services.ServerStatBoardChartService)
		pb.RegisterServerStatBoardChartServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.PlanService{}).(*services.PlanService)
		pb.RegisterPlanServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.UserPlanService{}).(*services.UserPlanService)
		pb.RegisterUserPlanServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.ServerDomainHourlyStatService{}).(*services.ServerDomainHourlyStatService)
		pb.RegisterServerDomainHourlyStatServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.TrafficDailyStatService{}).(*services.TrafficDailyStatService)
		pb.RegisterTrafficDailyStatServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.HTTPCacheTaskKeyService{}).(*services.HTTPCacheTaskKeyService)
		pb.RegisterHTTPCacheTaskKeyServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.HTTPCacheTaskService{}).(*services.HTTPCacheTaskService)
		pb.RegisterHTTPCacheTaskServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.ServerBandwidthStatService{}).(*services.ServerBandwidthStatService)
		pb.RegisterServerBandwidthStatServiceServer(server, instance)
		this.rest(instance)
	}

	{
		var instance = this.serviceInstance(&services.UpdatingServerListService{}).(*services.UpdatingServerListService)
		pb.RegisterUpdatingServerListServiceServer(server, instance)
		this.rest(instance)
	}

	APINodeServicesRegister(this, server)

	// TODO check service names
	for serviceName := range server.GetServiceInfo() {
		var index = strings.LastIndex(serviceName, ".")
		if index >= 0 {
			serviceName = serviceName[index+1:]
		}
		_, ok := restServicesMap[serviceName]
		if !ok {
			panic("can not find service '" + serviceName + "' in rest")
		}
	}
}

func (this *APINode) rest(instance interface{}) {
	this.serviceInstanceLocker.Lock()
	defer this.serviceInstanceLocker.Unlock()

	var name = reflect.TypeOf(instance).String()
	var index = strings.LastIndex(name, ".")
	if index >= 0 {
		name = name[index+1:]
	}

	_, ok := restServicesMap[name]
	if ok {
		return
	}
	restServicesMap[name] = reflect.ValueOf(instance)
}

func (this *APINode) serviceInstance(instance interface{}) interface{} {
	this.serviceInstanceLocker.Lock()
	defer this.serviceInstanceLocker.Unlock()

	var typeName = reflect.TypeOf(instance).String()
	result, ok := this.serviceInstanceMap[typeName]
	if ok {
		return result
	}

	this.serviceInstanceMap[typeName] = instance
	return instance
}
