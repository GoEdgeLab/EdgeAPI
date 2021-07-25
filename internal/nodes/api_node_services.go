// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nodes

import (
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services/nameservers"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc"
	"reflect"
	"strings"
)

// 注册服务
func (this *APINode) registerServices(server *grpc.Server) {
	{
		instance := this.serviceInstance(&services.APITokenService{}).(*services.APITokenService)
		pb.RegisterAPITokenServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.AdminService{}).(*services.AdminService)
		pb.RegisterAdminServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeGrantService{}).(*services.NodeGrantService)
		pb.RegisterNodeGrantServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerService{}).(*services.ServerService)
		pb.RegisterServerServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeService{}).(*services.NodeService)
		pb.RegisterNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeClusterService{}).(*services.NodeClusterService)
		pb.RegisterNodeClusterServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeIPAddressService{}).(*services.NodeIPAddressService)
		pb.RegisterNodeIPAddressServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.APINodeService{}).(*services.APINodeService)
		pb.RegisterAPINodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.OriginService{}).(*services.OriginService)
		pb.RegisterOriginServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPWebService{}).(*services.HTTPWebService)
		pb.RegisterHTTPWebServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ReverseProxyService{}).(*services.ReverseProxyService)
		pb.RegisterReverseProxyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPGzipService{}).(*services.HTTPGzipService)
		pb.RegisterHTTPGzipServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPHeaderPolicyService{}).(*services.HTTPHeaderPolicyService)
		pb.RegisterHTTPHeaderPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPHeaderService{}).(*services.HTTPHeaderService)
		pb.RegisterHTTPHeaderServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPPageService{}).(*services.HTTPPageService)
		pb.RegisterHTTPPageServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPAccessLogPolicyService{}).(*services.HTTPAccessLogPolicyService)
		pb.RegisterHTTPAccessLogPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPCachePolicyService{}).(*services.HTTPCachePolicyService)
		pb.RegisterHTTPCachePolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPFirewallPolicyService{}).(*services.HTTPFirewallPolicyService)
		pb.RegisterHTTPFirewallPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.FirewallService{}).(*services.FirewallService)
		pb.RegisterFirewallServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPLocationService{}).(*services.HTTPLocationService)
		pb.RegisterHTTPLocationServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPWebsocketService{}).(*services.HTTPWebsocketService)
		pb.RegisterHTTPWebsocketServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPRewriteRuleService{}).(*services.HTTPRewriteRuleService)
		pb.RegisterHTTPRewriteRuleServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.SSLCertService{}).(*services.SSLCertService)
		pb.RegisterSSLCertServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.SSLPolicyService{}).(*services.SSLPolicyService)
		pb.RegisterSSLPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.SysSettingService{}).(*services.SysSettingService)
		pb.RegisterSysSettingServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPFirewallRuleGroupService{}).(*services.HTTPFirewallRuleGroupService)
		pb.RegisterHTTPFirewallRuleGroupServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPFirewallRuleSetService{}).(*services.HTTPFirewallRuleSetService)
		pb.RegisterHTTPFirewallRuleSetServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.DBNodeService{}).(*services.DBNodeService)
		pb.RegisterDBNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeLogService{}).(*services.NodeLogService)
		pb.RegisterNodeLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPAccessLogService{}).(*services.HTTPAccessLogService)
		pb.RegisterHTTPAccessLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageService{}).(*services.MessageService)
		pb.RegisterMessageServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageRecipientService{}).(*services.MessageRecipientService)
		pb.RegisterMessageRecipientServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageReceiverService{}).(*services.MessageReceiverService)
		pb.RegisterMessageReceiverServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageMediaService{}).(*services.MessageMediaService)
		pb.RegisterMessageMediaServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageRecipientGroupService{}).(*services.MessageRecipientGroupService)
		pb.RegisterMessageRecipientGroupServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageMediaInstanceService{}).(*services.MessageMediaInstanceService)
		pb.RegisterMessageMediaInstanceServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageTaskService{}).(*services.MessageTaskService)
		pb.RegisterMessageTaskServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MessageTaskLogService{}).(*services.MessageTaskLogService)
		pb.RegisterMessageTaskLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeGroupService{}).(*services.NodeGroupService)
		pb.RegisterNodeGroupServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeRegionService{}).(*services.NodeRegionService)
		pb.RegisterNodeRegionServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodePriceItemService{}).(*services.NodePriceItemService)
		pb.RegisterNodePriceItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerGroupService{}).(*services.ServerGroupService)
		pb.RegisterServerGroupServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.IPLibraryService{}).(*services.IPLibraryService)
		pb.RegisterIPLibraryServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.FileChunkService{}).(*services.FileChunkService)
		pb.RegisterFileChunkServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.FileService{}).(*services.FileService)
		pb.RegisterFileServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.RegionCountryService{}).(*services.RegionCountryService)
		pb.RegisterRegionCountryServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.RegionProvinceService{}).(*services.RegionProvinceService)
		pb.RegisterRegionProvinceServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.IPListService{}).(*services.IPListService)
		pb.RegisterIPListServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.IPItemService{}).(*services.IPItemService)
		pb.RegisterIPItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.LogService{}).(*services.LogService)
		pb.RegisterLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.DNSProviderService{}).(*services.DNSProviderService)
		pb.RegisterDNSProviderServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.DNSDomainService{}).(*services.DNSDomainService)
		pb.RegisterDNSDomainServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.DNSService{}).(*services.DNSService)
		pb.RegisterDNSServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ACMEUserService{}).(*services.ACMEUserService)
		pb.RegisterACMEUserServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ACMETaskService{}).(*services.ACMETaskService)
		pb.RegisterACMETaskServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ACMEAuthenticationService{}).(*services.ACMEAuthenticationService)
		pb.RegisterACMEAuthenticationServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.UserService{}).(*services.UserService)
		pb.RegisterUserServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerDailyStatService{}).(*services.ServerDailyStatService)
		pb.RegisterServerDailyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.UserBillService{}).(*services.UserBillService)
		pb.RegisterUserBillServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.UserNodeService{}).(*services.UserNodeService)
		pb.RegisterUserNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.LoginService{}).(*services.LoginService)
		pb.RegisterLoginServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.UserAccessKeyService{}).(*services.UserAccessKeyService)
		pb.RegisterUserAccessKeyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.SysLockerService{}).(*services.SysLockerService)
		pb.RegisterSysLockerServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeTaskService{}).(*services.NodeTaskService)
		pb.RegisterNodeTaskServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeValueService{}).(*services.NodeValueService)
		pb.RegisterNodeValueServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.DBService{}).(*services.DBService)
		pb.RegisterDBServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerRegionCityMonthlyStatService{}).(*services.ServerRegionCityMonthlyStatService)
		pb.RegisterServerRegionCityMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerRegionCountryMonthlyStatService{}).(*services.ServerRegionCountryMonthlyStatService)
		pb.RegisterServerRegionCountryMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerRegionProvinceMonthlyStatService{}).(*services.ServerRegionProvinceMonthlyStatService)
		pb.RegisterServerRegionProvinceMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerRegionProviderMonthlyStatService{}).(*services.ServerRegionProviderMonthlyStatService)
		pb.RegisterServerRegionProviderMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerClientSystemMonthlyStatService{}).(*services.ServerClientSystemMonthlyStatService)
		pb.RegisterServerClientSystemMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerClientBrowserMonthlyStatService{}).(*services.ServerClientBrowserMonthlyStatService)
		pb.RegisterServerClientBrowserMonthlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.ServerHTTPFirewallDailyStatService{}).(*services.ServerHTTPFirewallDailyStatService)
		pb.RegisterServerHTTPFirewallDailyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.DNSTaskService{}).(*services.DNSTaskService)
		pb.RegisterDNSTaskServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeClusterFirewallActionService{}).(*services.NodeClusterFirewallActionService)
		pb.RegisterNodeClusterFirewallActionServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MonitorNodeService{}).(*services.MonitorNodeService)
		pb.RegisterMonitorNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.AuthorityKeyService{}).(*services.AuthorityKeyService)
		pb.RegisterAuthorityKeyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.AuthorityNodeService{}).(*services.AuthorityNodeService)
		pb.RegisterAuthorityNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.LatestItemService{}).(*services.LatestItemService)
		pb.RegisterLatestItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeThresholdService{}).(*services.NodeThresholdService)
		pb.RegisterNodeThresholdServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPFastcgiService{}).(*services.HTTPFastcgiService)
		pb.RegisterHTTPFastcgiServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSClusterService{}).(*nameservers.NSClusterService)
		pb.RegisterNSClusterServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSNodeService{}).(*nameservers.NSNodeService)
		pb.RegisterNSNodeServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSDomainService{}).(*nameservers.NSDomainService)
		pb.RegisterNSDomainServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSRecordService{}).(*nameservers.NSRecordService)
		pb.RegisterNSRecordServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSRouteService{}).(*nameservers.NSRouteService)
		pb.RegisterNSRouteServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSKeyService{}).(*nameservers.NSKeyService)
		pb.RegisterNSKeyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSAccessLogService{}).(*nameservers.NSAccessLogService)
		pb.RegisterNSAccessLogServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSRecordHourlyStatService{}).(*nameservers.NSRecordHourlyStatService)
		pb.RegisterNSRecordHourlyStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&nameservers.NSService{}).(*nameservers.NSService)
		pb.RegisterNSServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.HTTPAuthPolicyService{}).(*services.HTTPAuthPolicyService)
		pb.RegisterHTTPAuthPolicyServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MetricItemService{}).(*services.MetricItemService)
		pb.RegisterMetricItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.NodeClusterMetricItemService{}).(*services.NodeClusterMetricItemService)
		pb.RegisterNodeClusterMetricItemServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MetricStatService{}).(*services.MetricStatService)
		pb.RegisterMetricStatServiceServer(server, instance)
		this.rest(instance)
	}
	{
		instance := this.serviceInstance(&services.MetricChartService{}).(*services.MetricChartService)
		pb.RegisterMetricChartServiceServer(server, instance)
		this.rest(instance)
	}

	{
		instance := this.serviceInstance(&services.ServerStatBoardService{}).(*services.ServerStatBoardService)
		pb.RegisterServerStatBoardServiceServer(server, instance)
		this.rest(instance)
	}

	{
		instance := this.serviceInstance(&services.ServerStatBoardChartService{}).(*services.ServerStatBoardChartService)
		pb.RegisterServerStatBoardChartServiceServer(server, instance)
		this.rest(instance)
	}

	// TODO check service names
	for serviceName := range server.GetServiceInfo() {
		index := strings.LastIndex(serviceName, ".")
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
	index := strings.LastIndex(name, ".")
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

	typeName := reflect.TypeOf(instance).String()
	result, ok := this.serviceInstanceMap[typeName]
	if ok {
		return result
	}

	this.serviceInstanceMap[typeName] = instance
	return instance
}
