package utils

import (
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/iwind/TeaGo/logs"
	"github.com/miekg/dns"
)

var sharedDNSClient *dns.Client
var sharedDNSConfig *dns.ClientConfig

func init() {
	if !teaconst.IsMain {
		return
	}

	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		logs.Println("ERROR: configure dns client failed: " + err.Error())
		return
	}

	sharedDNSConfig = config
	sharedDNSClient = &dns.Client{}

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
	for _, serverAddr := range sharedDNSConfig.Servers {
		r, _, err := sharedDNSClient.Exchange(m, configutils.QuoteIP(serverAddr)+":"+sharedDNSConfig.Port)
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
// TODO 可以设置使用的DNS主机地址
func LookupNS(host string) ([]string, error) {
	var m = new(dns.Msg)

	m.SetQuestion(host+".", dns.TypeNS)
	m.RecursionDesired = true

	var result = []string{}

	var lastErr error
	var hasValidServer = false
	for _, serverAddr := range sharedDNSConfig.Servers {
		r, _, err := sharedDNSClient.Exchange(m, configutils.QuoteIP(serverAddr)+":"+sharedDNSConfig.Port)
		if err != nil {
			lastErr = err
			continue
		}

		hasValidServer = true

		if len(r.Answer) == 0 {
			continue
		}

		for _, answer := range r.Answer {
			result = append(result, answer.(*dns.NS).Ns)
		}
		break
	}

	if hasValidServer {
		return result, nil
	}

	return nil, lastErr
}

// LookupTXT 获取CNAME
// TODO 可以设置使用的DNS主机地址
func LookupTXT(host string) ([]string, error) {
	var m = new(dns.Msg)

	m.SetQuestion(host+".", dns.TypeTXT)
	m.RecursionDesired = true

	var lastErr error
	var result = []string{}
	var hasValidServer = false
	for _, serverAddr := range sharedDNSConfig.Servers {
		r, _, err := sharedDNSClient.Exchange(m, configutils.QuoteIP(serverAddr)+":"+sharedDNSConfig.Port)
		if err != nil {
			lastErr = err
			continue
		}
		hasValidServer = true

		if len(r.Answer) == 0 {
			continue
		}

		for _, answer := range r.Answer {
			result = append(result, answer.(*dns.TXT).Txt...)
		}

		break
	}

	if hasValidServer {
		return result, nil
	}

	return nil, lastErr
}
