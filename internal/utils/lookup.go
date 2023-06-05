package utils

import (
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/taskutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/fsnotify/fsnotify"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/miekg/dns"
	"sync"
)

var sharedDNSClient *dns.Client
var sharedDNSConfig *dns.ClientConfig
var sharedDNSLocker = &sync.RWMutex{}

func init() {
	if !teaconst.IsMain {
		return
	}

	var resolvConfFile = "/etc/resolv.conf"
	config, err := dns.ClientConfigFromFile(resolvConfFile)
	if err != nil {
		logs.Println("ERROR: configure dns client failed: " + err.Error())
		return
	}

	sharedDNSConfig = config
	sharedDNSClient = &dns.Client{}

	// 监视文件变化，以便及时更新配置
	go func() {
		watcher, watcherErr := fsnotify.NewWatcher()
		if watcherErr == nil {
			err = watcher.Add(resolvConfFile)
			for range watcher.Events {
				newConfig, err := dns.ClientConfigFromFile(resolvConfFile)
				if err == nil && newConfig != nil {
					sharedDNSLocker.Lock()
					sharedDNSConfig = newConfig
					sharedDNSLocker.Unlock()
				}
			}
		}
	}()
}

// LookupCNAME 查询CNAME记录
// TODO 可以设置使用的DNS主机地址
func LookupCNAME(host string) (string, error) {
	if sharedDNSClient == nil {
		return "", errors.New("could not find dns client")
	}

	var m = new(dns.Msg)

	m.SetQuestion(host+".", dns.TypeCNAME)
	m.RecursionDesired = true

	var lastErr error
	var serverAddrs = composeDNSResolverAddrs(nil)

	for _, serverAddr := range serverAddrs {
		r, _, err := sharedDNSClient.Exchange(m, serverAddr)
		if err != nil {
			lastErr = err
			continue
		}
		if len(r.Answer) == 0 {
			continue
		}

		return r.Answer[0].(*dns.CNAME).Target, nil
	}
	return "", lastErr
}

// LookupNS 查询NS记录
func LookupNS(host string, extraResolvers []*dnsconfigs.DNSResolver) ([]string, error) {
	var m = new(dns.Msg)

	m.SetQuestion(host+".", dns.TypeNS)
	m.RecursionDesired = true

	var result = []string{}

	var lastErr error
	var hasValidServer = false
	var serverAddrs = composeDNSResolverAddrs(extraResolvers)
	if len(serverAddrs) == 0 {
		return nil, nil
	}

	taskErr := taskutils.RunConcurrent(serverAddrs, taskutils.DefaultConcurrent, func(task any, locker *sync.RWMutex) {
		var serverAddr = task.(string)
		r, _, err := sharedDNSClient.Exchange(m, serverAddr)
		if err != nil {
			lastErr = err
			return
		}

		hasValidServer = true

		if len(r.Answer) == 0 {
			return
		}

		for _, answer := range r.Answer {
			var value = answer.(*dns.NS).Ns
			locker.Lock()
			if len(value) > 0 && !lists.ContainsString(result, value) {
				result = append(result, value)
			}
			locker.Unlock()
		}
	})
	if taskErr != nil {
		return result, taskErr
	}

	if hasValidServer {
		return result, nil
	}

	return nil, lastErr
}

// LookupTXT 获取CNAME
func LookupTXT(host string, extraResolvers []*dnsconfigs.DNSResolver) ([]string, error) {
	var m = new(dns.Msg)

	m.SetQuestion(host+".", dns.TypeTXT)
	m.RecursionDesired = true

	var lastErr error
	var result = []string{}
	var hasValidServer = false
	var serverAddrs = composeDNSResolverAddrs(extraResolvers)
	if len(serverAddrs) == 0 {
		return nil, nil
	}

	taskErr := taskutils.RunConcurrent(serverAddrs, taskutils.DefaultConcurrent, func(task any, locker *sync.RWMutex) {
		var serverAddr = task.(string)
		r, _, err := sharedDNSClient.Exchange(m, serverAddr)
		if err != nil {
			lastErr = err
			return
		}
		hasValidServer = true

		if len(r.Answer) == 0 {
			return
		}

		for _, answer := range r.Answer {
			for _, txt := range answer.(*dns.TXT).Txt {
				locker.Lock()
				if len(txt) > 0 && !lists.ContainsString(result, txt) {
					result = append(result, txt)
				}
				locker.Unlock()
			}
		}
	})
	if taskErr != nil {
		return result, taskErr
	}

	if hasValidServer {
		return result, nil
	}

	return nil, lastErr
}

// 组合DNS解析服务器地址
func composeDNSResolverAddrs(extraResolvers []*dnsconfigs.DNSResolver) []string {
	sharedDNSLocker.RLock()
	defer sharedDNSLocker.RUnlock()

	// 这里不处理重复，方便我们可以多次重试
	var servers = sharedDNSConfig.Servers
	var port = sharedDNSConfig.Port

	var serverAddrs = []string{}
	for _, serverAddr := range servers {
		serverAddrs = append(serverAddrs, configutils.QuoteIP(serverAddr)+":"+port)
	}
	for _, resolver := range extraResolvers {
		serverAddrs = append(serverAddrs, resolver.Addr())
	}
	return serverAddrs
}
