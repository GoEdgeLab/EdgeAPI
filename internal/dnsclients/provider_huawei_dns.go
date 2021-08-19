// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/huaweidns"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const HuaweiDNSEndpoint = "https://dns.cn-north-1.myhuaweicloud.com/"

var huaweiDNSHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type HuaweiDNSProvider struct {
	BaseProvider

	accessKeyId     string
	accessKeySecret string
}

// Auth 认证
func (this *HuaweiDNSProvider) Auth(params maps.Map) error {
	this.accessKeyId = params.GetString("accessKeyId")
	this.accessKeySecret = params.GetString("accessKeySecret")
	if len(this.accessKeyId) == 0 {
		return errors.New("'accessKeyId' should not be empty")
	}
	if len(this.accessKeySecret) == 0 {
		return errors.New("'accessKeySecret' should not be empty")
	}
	return nil
}

// GetDomains 获取所有域名列表
func (this *HuaweiDNSProvider) GetDomains() (domains []string, err error) {
	var resp = new(huaweidns.ZonesResponse)
	err = this.doAPI(http.MethodGet, "/v2/zones", map[string]string{}, nil, resp)
	if err != nil {
		return nil, err
	}

	for _, zone := range resp.Zones {
		zone.Name = strings.TrimSuffix(zone.Name, ".")
		domains = append(domains, zone.Name)
	}
	return
}

// GetRecords 获取域名解析记录列表
func (this *HuaweiDNSProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return nil, err
	}

	var limit = 500
	for i := 0; i <= 100; i++ {
		var resp = new(huaweidns.ZoneRecordSetsResponse)
		err := this.doAPI(http.MethodGet, "/v2.1/zones/"+zoneId+"/recordsets", map[string]string{
			"offset": types.String(i * limit),
			"limit":  types.String(limit),
		}, map[string]interface{}{}, resp)
		if err != nil {
			return nil, err
		}
		if len(resp.RecordSets) == 0 {
			break
		}
		for _, recordSet := range resp.RecordSets {
			for _, value := range recordSet.Records {
				name := strings.TrimSuffix(recordSet.Name, "."+domain+".")

				records = append(records, &dnstypes.Record{
					Id:    recordSet.Id + "@" + value,
					Name:  name,
					Type:  recordSet.Type,
					Value: value,
					Route: recordSet.Line,
				})
			}
		}
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *HuaweiDNSProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	// 自定义线路
	var resp = new(huaweidns.CustomLinesResponse)
	err = this.doAPI(http.MethodGet, "/v2.1/customlines", map[string]string{}, maps.Map{}, resp)
	if err != nil {
		return nil, err
	}
	for _, line := range resp.Lines {
		routes = append(routes, &dnstypes.Route{
			Name: line.Name,
			Code: line.LineId,
		})
	}

	// 官方线路列表
	// 参考：https://support.huaweicloud.com/api-dns/zh-cn_topic_0085546214.html

	// 运营商线路
	routes = append(routes, []*dnstypes.Route{
		{
			Code: "default_view",
			Name: "全网默认",
		},

		{
			Code: "Dianxin",
			Name: "电信",
		},

		{
			Code: "Liantong",
			Name: "联通",
		},

		{
			Code: "Yidong",
			Name: "移动",
		},

		{
			Code: "Jiaoyuwang",
			Name: "教育网",
		},

		{
			Code: "Tietong",
			Name: "铁通",
		},

		{
			Code: "Pengboshi",

			Name: "鹏博士",
		},

		{
			Code: "CN",
			Name: "中国",
		},

		{
			Code: "Abroad",
			Name: "全球",
		},
	}...)

	// 地域线路细分（全球）
	routes = append(routes, []*dnstypes.Route{
		{
			Code: "AP",
			Name: "全球_亚太地区",
		},
		{
			Code: "AE",

			Name: "全球_阿联酋",
		},
		{
			Code: "AM",
			Name: "全球_亚美尼亚",
		},
		{
			Code: "AZ",
			Name: "全球_阿塞拜疆",
		},
		{
			Code: "BD",
			Name: "全球_孟加拉",
		},
		{
			Code: "BH",
			Name: "全球_巴林",
		},
		{
			Code: "BN",
			Name: "全球_文莱",
		},
		{
			Code: "BT",
			Name: "全球_不丹",
		},
		{
			Code: "CX",
			Name: "全球_圣诞岛",
		},
		{
			Code: "GE",
			Name: "全球_格鲁吉亚",
		},
		{
			Code: "HK",
			Name: "全球_中国香港",
		},
		{
			Code: "ID",
			Name: "全球_印度尼西亚",
		},
		{
			Code: "IN",
			Name: "全球_印度",
		},

		{
			Code: "IQ",
			Name: "全球_伊拉克",
		},

		{
			Code: "JO",
			Name: "全球_约旦",
		},

		{
			Code: "KG",
			Name: "全球_吉尔吉斯斯坦",
		},

		{
			Code: "KH",
			Name: "全球_柬埔寨",
		},

		{
			Code: "KW",
			Name: "全球_科威特",
		},

		{
			Code: "KZ",
			Name: "全球_哈萨克斯坦",
		},
		{
			Code: "LB",
			Name: "全球_黎巴嫩",
		},
		{
			Code: "LK",
			Name: "全球_斯里兰卡",
		},
		{
			Code: "MM",
			Name: "全球_缅甸",
		},
		{
			Code: "MN",
			Name: "全球_蒙古",
		},
		{
			Code: "MO",
			Name: "全球_中国澳门",
		},
		{
			Code: "MV",
			Name: "全球_马尔代夫",
		},
		{
			Code: "MY",
			Name: "全球_马来西亚",
		},
		{
			Code: "NP",
			Name: "全球_尼泊尔",
		},
		{
			Code: "OM",
			Name: "全球_阿曼",
		},
		{
			Code: "PH",
			Name: "全球_菲律宾",
		},
		{
			Code: "PK",
			Name: "全球_巴基斯坦",
		},
		{
			Code: "PS",
			Name: "全球_巴勒斯坦",
		},
		{
			Code: "QA",
			Name: "全球_卡塔尔",
		},
		{
			Code: "SA",
			Name: "全球_沙特阿拉伯",
		},
		{
			Code: "SG",
			Name: "全球_新加坡",
		},
		{
			Code: "TH",
			Name: "全球_泰国",
		},
		{
			Code: "TJ",
			Name: "全球_塔吉克斯坦",
		},
		{
			Code: "TL",
			Name: "全球_东帝汶",
		},
		{
			Code: "TM",
			Name: "全球_土库曼斯坦",
		},
		{
			Code: "TW",
			Name: "全球_中国台湾",
		},
		{
			Code: "UZ",
			Name: "全球_乌兹别克斯坦",
		},
		{
			Code: "VN",
			Name: "全球_越南",
		},
		{
			Code: "YE",
			Name: "全球_也门",
		},
		{
			Code: "AS",
			Name: "全球_美属萨摩亚",
		},
		{
			Code: "CK",
			Name: "全球_库克群岛",
		},
		{
			Code: "FM",
			Name: "全球_密克罗尼西亚",
		},
		{
			Code: "GU",
			Name: "全球_关岛",
		},
		{
			Code: "KI",
			Name: "全球_基里巴斯",
		},
		{
			Code: "MH",
			Name: "全球_马绍尔群岛",
		},
		{
			Code: "MP",
			Name: "全球_北马里亚纳群岛",
		},
		{
			Code: "NC",
			Name: "全球_新喀里多尼亚",
		},
		{
			Code: "NF",
			Name: "全球_诺福克岛",
		},
		{
			Code: "NR",
			Name: "全球_瑙鲁",
		},
		{
			Code: "PF",
			Name: "全球_法属波利尼西亚",
		},
		{
			Code: "PG",
			Name: "全球_巴布亚新几内亚",
		},
		{
			Code: "PW",
			Name: "全球_帕劳",
		},
		{
			Code: "SB",
			Name: "全球_所罗门群岛",
		},
		{
			Code: "TK",
			Name: "全球_托克劳群岛",
		},
		{
			Code: "TO",
			Name: "全球_汤加",
		},
		{
			Code: "TV",
			Name: "全球_图瓦卢",
		},
		{
			Code: "VU",
			Name: "全球_瓦努阿图",
		},
		{
			Code: "WS",
			Name: "全球_萨摩亚",
		},
		{
			Code: "CY",
			Name: "全球_塞浦路斯",
		},
		{
			Code: "IL",
			Name: "全球_以色列",
		},
		{
			Code: "JP",
			Name: "全球_日本",
		},
		{
			Code: "KR",
			Name: "全球_韩国",
		},
		{
			Code: "TR",
			Name: "全球_土耳其",
		},
		{
			Code: "IR",
			Name: "全球_伊朗",
		},
		{
			Code: "SY",
			Name: "全球_叙利亚",
		},
		{
			Code: "CN",
			Name: "全球_中国",
		},
		{
			Code: "OA",
			Name: "全球_大洋洲",
		},
		{
			Code: "AU",
			Name: "全球_澳大利亚",
		},
		{
			Code: "NZ",
			Name: "全球_新西兰",
		},
		{
			Code: "EU",
			Name: "全球_欧洲",
		},
		{
			Code: "IO",
			Name: "全球_英属印度洋领地",
		},
		{
			Code: "BY",
			Name: "全球_白俄罗斯",
		},
		{
			Code: "UA",
			Name: "全球_乌克兰",
		},
		{
			Code: "RU",
			Name: "全球_俄罗斯",
		},
		{
			Code: "AD",
			Name: "全球_安道尔",
		},
		{
			Code: "AL",
			Name: "全球_阿尔巴尼亚",
		},
		{
			Code: "AT",
			Name: "全球_奥地利",
		},
		{
			Code: "AX",
			Name: "全球_奥兰群岛",
		},
		{
			Code: "BE",
			Name: "全球_比利时",
		},
		{
			Code: "BG",
			Name: "全球_保加利亚",
		},
		{
			Code: "CH",
			Name: "全球_瑞士",
		},
		{
			Code: "CZ",
			Name: "全球_捷克",
		},
		{
			Code: "DE",
			Name: "全球_德国",
		},
		{
			Code: "DK",
			Name: "全球_丹麦",
		},
		{
			Code: "EE",
			Name: "全球_爱沙尼亚",
		},
		{
			Code: "ES",
			Name: "全球_西班牙",
		},
		{
			Code: "FI",
			Name: "全球_芬兰",
		},
		{
			Code: "FO",
			Name: "全球_法罗群岛",
		},
		{
			Code: "FR",
			Name: "全球_法国",
		},
		{
			Code: "GB",
			Name: "全球_英国",
		},
		{
			Code: "GG",
			Name: "全球_根西岛",
		},
		{
			Code: "GI",
			Name: "全球_直布罗陀",
		},
		{
			Code: "GR",
			Name: "全球_希腊",
		},
		{
			Code: "HR",
			Name: "全球_克罗地亚",
		},
		{
			Code: "HU",
			Name: "全球_匈牙利",
		},
		{
			Code: "IE",
			Name: "全球_爱尔兰",
		},
		{
			Code: "IM",
			Name: "全球_马恩岛",
		},
		{
			Code: "IS",
			Name: "全球_冰岛",
		},
		{
			Code: "IT",
			Name: "全球_意大利",
		},
		{
			Code: "JE",
			Name: "全球_泽西岛",
		},
		{
			Code: "LI",
			Name: "全球_列支敦士登",
		},
		{
			Code: "LT",
			Name: "全球_立陶宛",
		},
		{
			Code: "LU",
			Name: "全球_卢森堡",
		},
		{
			Code: "LV",
			Name: "全球_拉脱维亚",
		},
		{
			Code: "MC",
			Name: "全球_摩纳哥",
		},

		{
			Code: "MD",
			Name: "全球_摩尔多瓦",
		},
		{
			Code: "ME",
			Name: "全球_黑山",
		},
		{
			Code: "MK",
			Name: "全球_北马其顿",
		},
		{
			Code: "MT",
			Name: "全球_马耳他",
		},
		{
			Code: "NL",
			Name: "全球_荷兰",
		},
		{
			Code: "NO",
			Name: "全球_挪威",
		},
		{
			Code: "PL",
			Name: "全球_波兰",
		},
		{
			Code: "PT",
			Name: "全球_葡萄牙",
		},
		{
			Code: "RO",
			Name: "全球_罗马尼亚",
		},
		{
			Code: "RS",
			Name: "全球_塞尔维亚",
		},
		{
			Code: "SE",
			Name: "全球_瑞典",
		},
		{
			Code: "SI",
			Name: "全球_斯洛文尼亚",
		},
		{
			Code: "SK",
			Name: "全球_斯洛伐克",
		},
		{
			Code: "SM",
			Name: "全球_圣马力诺",
		},
		{
			Code: "VA",
			Name: "全球_梵蒂冈",
		},
		{
			Code: "XK",
			Name: "全球_科索沃",
		},
		{
			Code: "GL",
			Name: "全球_格陵兰",
		},
		{
			Code: "NA",
			Name: "全球_北美洲",
		},
		{
			Code: "AG",
			Name: "全球_安提瓜和巴布达",
		},
		{
			Code: "BB",
			Name: "全球_巴巴多斯",
		},
		{
			Code: "BS",
			Name: "全球_巴哈马",
		},
		{
			Code: "BZ",
			Name: "全球_伯利兹",
		},
		{
			Code: "CR",
			Name: "全球_哥斯达黎加",
		},
		{
			Code: "DM",
			Name: "全球_多米尼克",
		},
		{
			Code: "DO",
			Name: "全球_多米尼加",
		},
		{
			Code: "GD",
			Name: "全球_格林纳达",
		},
		{
			Code: "GT",
			Name: "全球_危地马拉",
		},
		{
			Code: "HN",
			Name: "全球_洪都拉斯",
		},
		{
			Code: "HT",
			Name: "全球_海地",
		},
		{
			Code: "JM",
			Name: "全球_牙买加",
		},
		{
			Code: "KN",
			Name: "全球_圣基茨和尼维斯",
		},
		{
			Code: "KY",
			Name: "全球_开曼群岛",
		},
		{
			Code: "LC",
			Name: "全球_圣卢西亚",
		},
		{
			Code: "MX",
			Name: "全球_墨西哥",
		},
		{
			Code: "NI",
			Name: "全球_尼加拉瓜",
		},
		{
			Code: "PA",
			Name: "全球_巴拿马",
		},
		{
			Code: "PR",
			Name: "全球_波多黎各",
		},
		{
			Code: "SV",
			Name: "全球_萨尔瓦多",
		},
		{
			Code: "TC",
			Name: "全球_特克斯和凯科斯群岛",
		},
		{
			Code: "TT",
			Name: "全球_特立尼达和多巴哥",
		},
		{
			Code: "VG",
			Name: "全球_英属维尔京群岛",
		},
		{
			Code: "VI",
			Name: "全球_美属维尔京群岛",
		},
		{
			Code: "CA",
			Name: "全球_加拿大",
		},
		{
			Code: "US",
			Name: "全球_美国",
		},
		{
			Code: "VC",
			Name: "全球_圣文森特和格林纳丁斯",
		},
		{
			Code: "PM",
			Name: "全球_圣皮埃尔和密克隆群岛",
		},
		{
			Code: "AN",
			Name: "全球_荷属安的列斯群岛",
		},
		{
			Code: "CU",
			Name: "全球_古巴",
		},
		{
			Code: "GL",
			Name: "全球_格陵兰岛",
		},
		{
			Code: "LA",
			Name: "全球_南美洲",
		},
		{
			Code: "AI",
			Name: "全球_安圭拉",
		},
		{
			Code: "AW",
			Name: "全球_阿鲁巴",
		},
		{
			Code: "BL",
			Name: "全球_圣巴泰勒米岛",
		},
		{
			Code: "BM",
			Name: "全球_百慕大",
		},
		{
			Code: "GP",
			Name: "全球_瓜德罗普",
		},
		{
			Code: "MS",
			Name: "全球_蒙特塞拉特岛",
		},
		{
			Code: "AR",
			Name: "全球_阿根廷",
		},
		{
			Code: "BO",
			Name: "全球_玻利维亚",
		},
		{
			Code: "BR",
			Name: "全球_巴西",
		},
		{
			Code: "CL",
			Name: "全球_智利",
		},
		{
			Code: "CO",
			Name: "全球_哥伦比亚",
		},
		{
			Code: "CW",
			Name: "全球_库拉索",
		},
		{
			Code: "EC",
			Name: "全球_厄瓜多尔",
		},
		{
			Code: "GF",
			Name: "全球_法属圭亚那",
		},
		{
			Code: "GY",
			Name: "全球_圭亚那",
		},
		{
			Code: "PE",
			Name: "全球_秘鲁",
		},
		{
			Code: "PY",
			Name: "全球_巴拉圭",
		},
		{
			Code: "SR",
			Name: "全球_苏里南",
		},
		{
			Code: "UY",
			Name: "全球_乌拉圭",
		},
		{
			Code: "VE",
			Name: "全球_委内瑞拉",
		},
		{
			Code: "MF",
			Name: "全球_法属圣马丁",
		},
		{
			Code: "SX",
			Name: "全球_荷属圣马丁",
		},
		{
			Code: "AF",
			Name: "全球_非洲",
		},
		{
			Code: "AO",
			Name: "全球_安哥拉",
		},
		{
			Code: "BF",
			Name: "全球_布基纳法索",
		},
		{
			Code: "BI",
			Name: "全球_布隆迪",
		},
		{
			Code: "BJ",
			Name: "全球_贝宁",
		},
		{
			Code: "BW",
			Name: "全球_博茨瓦纳",
		},
		{
			Code: "CD",
			Name: "全球_刚果金",
		},
		{
			Code: "CF",
			Name: "全球_中非",
		},
		{
			Code: "CG",
			Name: "全球_刚果布",
		},
		{
			Code: "CI",
			Name: "全球_科特迪瓦",
		},
		{
			Code: "CM",
			Name: "全球_喀麦隆",
		},
		{
			Code: "CV",
			Name: "全球_佛得角",
		},
		{
			Code: "DJ",
			Name: "全球_吉布提",
		},
		{
			Code: "DZ",
			Name: "全球_阿尔及利亚",
		},
		{
			Code: "EG",
			Name: "全球_埃及",
		},
		{
			Code: "EH",
			Name: "全球_西撒哈拉",
		},
		{
			Code: "ER",
			Name: "全球_厄立特里亚",
		},
		{
			Code: "ET",
			Name: "全球_埃塞俄比亚",
		},
		{
			Code: "GA",
			Name: "全球_加蓬",
		},
		{
			Code: "GH",
			Name: "全球_加纳",
		},
		{
			Code: "GM",
			Name: "全球_冈比亚",
		},
		{
			Code: "GN",
			Name: "全球_几内亚",
		},
		{
			Code: "GQ",
			Name: "全球_赤道几内亚",
		},
		{
			Code: "GW",
			Name: "全球_几内亚比绍",
		},
		{
			Code: "KE",
			Name: "全球_肯尼亚",
		},
		{
			Code: "KM",
			Name: "全球_科摩罗",
		},
		{
			Code: "LR",
			Name: "全球_利比里亚",
		},
		{
			Code: "LS",
			Name: "全球_莱索托",
		},
		{
			Code: "LY",
			Name: "全球_利比亚",
		},
		{
			Code: "MA",
			Name: "全球_摩洛哥",
		},
		{
			Code: "MG",
			Name: "全球_马达加斯加",
		},
		{
			Code: "ML",
			Name: "全球_马里",
		},
		{
			Code: "MR",
			Name: "全球_毛里塔尼亚",
		},
		{
			Code: "MU",
			Name: "全球_毛里求斯",
		},
		{
			Code: "MW",
			Name: "全球_马拉维",
		},
		{
			Code: "MZ",
			Name: "全球_莫桑比克",
		},
		{
			Code: "NE",
			Name: "全球_尼日尔",
		},
		{
			Code: "NG",
			Name: "全球_尼日利亚",
		},
		{
			Code: "RE",
			Name: "全球_留尼汪",
		},
		{
			Code: "RW",
			Name: "全球_卢旺达",
		},
		{
			Code: "SC",
			Name: "全球_塞舌尔",
		},
		{
			Code: "SL",
			Name: "全球_塞拉利昂",
		},
		{
			Code: "SN",
			Name: "全球_塞内加尔",
		},
		{
			Code: "SO",
			Name: "全球_索马里",
		},
		{
			Code: "SS",
			Name: "全球_南苏丹",
		},
		{
			Code: "ST",
			Name: "全球_圣多美和普林西比",
		},
		{
			Code: "SZ",
			Name: "全球_斯威士兰",
		},
		{
			Code: "TD",
			Name: "全球_乍得",
		},
		{
			Code: "TG",
			Name: "全球_多哥",
		},
		{
			Code: "TN",
			Name: "全球_突尼斯",
		},
		{
			Code: "TZ",
			Name: "全球_坦桑尼亚",
		},
		{
			Code: "UG",
			Name: "全球_乌干达",
		},
		{
			Code: "YT",
			Name: "全球_马约特",
		},
		{
			Code: "ZA",
			Name: "全球_南非",
		},
		{
			Code: "ZM",
			Name: "全球_赞比亚",
		},
		{
			Code: "ZW",
			Name: "全球_津巴布韦",
		},
		{
			Code: "AQ",
			Name: "全球_南极洲",
		},
	}...)

	// 地域线路
	routes = append(routes, []*dnstypes.Route{
		{
			Code: "Beijing",
			Name: "中国_北京",
		},
		{
			Code: "Hebei",
			Name: "中国_河北",
		},
		{
			Code: "Tianjin",
			Name: "中国_天津",
		},
		{
			Code: "Shanxi",
			Name: "中国_山西",
		},
		{
			Code: "Neimenggu",
			Name: "中国_内蒙古",
		},
		{
			Code: "Heilongjiang",
			Name: "中国_黑龙江",
		},
		{
			Code: "Jilin",
			Name: "中国_吉林",
		},
		{
			Code: "Liaoning",
			Name: "中国_辽宁",
		},
		{
			Code: "Jiangsu",
			Name: "中国_江苏",
		},
		{
			Code: "Shanghai",
			Name: "中国_上海",
		},
		{
			Code: "Zhejiang",
			Name: "中国_浙江",
		},
		{
			Code: "Anhui",
			Name: "中国_安徽",
		},
		{
			Code: "Fujian",
			Name: "中国_福建",
		},
		{
			Code: "Jiangxi",
			Name: "中国_江西",
		},
		{
			Code: "Shandong",
			Name: "中国_山东",
		},
		{
			Code: "Hubei",
			Name: "中国_湖北",
		},
		{
			Code: "Hunan",
			Name: "中国_湖南",
		},
		{
			Code: "Henan",
			Name: "中国_河南",
		},
		{
			Code: "Guangdong",
			Name: "中国_广东",
		},
		{
			Code: "Guangxi",
			Name: "中国_广西",
		},
		{
			Code: "Hainan",
			Name: "中国_海南",
		},
		{
			Code: "Sichuan",
			Name: "中国_四川",
		},
		{
			Code: "Xizang",
			Name: "中国_西藏",
		},
		{
			Code: "Chongqing",
			Name: "中国_重庆",
		},
		{
			Code: "Yunnan",
			Name: "中国_云南",
		},
		{
			Code: "Guizhou",
			Name: "中国_贵州",
		},
		{
			Code: "Gansu",
			Name: "中国_甘肃",
		},
		{
			Code: "Xinjiang",
			Name: "中国_新疆",
		},
		{
			Code: "Shaanxi",
			Name: "中国_陕西",
		},
		{
			Code: "Qinghai",
			Name: "中国_青海",
		},
		{
			Code: "Ningxia",
			Name: "中国_宁夏",
		},
		{
			Code: "Huabei",
			Name: "中国_华北地区",
		},
		{
			Code: "Dongbei",
			Name: "中国_东北地区",
		},
		{
			Code: "Huadong",
			Name: "中国_华东地区",
		},
		{
			Code: "Huazhong",
			Name: "中国_华中地区",
		},
		{
			Code: "Huanan",
			Name: "中国_华南地区",
		},
		{
			Code: "Xinan",
			Name: "中国_西南地区",
		},
		{
			Code: "Xibei",
			Name: "中国_西北地区",
		},
	}...)

	return
}

// QueryRecord 查询单个记录
func (this *HuaweiDNSProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	var resp = new(huaweidns.RecordSetsResponse)
	err := this.doAPI(http.MethodGet, "/v2.1/recordsets", map[string]string{
		"name": name + "." + domain + ".",
		"type": recordType,
	}, maps.Map{}, resp)
	if err != nil {
		return nil, err
	}

	if len(resp.RecordSets) == 0 {
		return nil, nil
	}

	var recordSet = resp.RecordSets[0]
	if len(recordSet.Records) == 0 {
		return nil, nil
	}

	return &dnstypes.Record{
		Id:    recordSet.Id + "@" + recordSet.Records[0],
		Name:  name,
		Type:  recordType,
		Value: recordSet.Records[0],
		Route: recordSet.Line,
	}, nil
}

// AddRecord 设置记录
func (this *HuaweiDNSProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return err
	}

	var resp = new(huaweidns.ZonesCreateRecordSetResponse)
	err = this.doAPI(http.MethodPost, "/v2.1/zones/"+zoneId+"/recordsets", map[string]string{}, maps.Map{
		"name":        newRecord.Name + "." + domain + ".",
		"description": "CDN系统自动创建",
		"type":        newRecord.Type,
		"records":     []string{newRecord.Value},
		"line":        newRecord.Route,
	}, resp)
	if err != nil {
		return err
	}

	newRecord.Id = resp.Id + "@" + newRecord.Value

	return nil
}

// UpdateRecord 修改记录
func (this *HuaweiDNSProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return err
	}

	var recordId string
	var atIndex = strings.Index(record.Id, "@")
	if atIndex > 0 {
		recordId = record.Id[:atIndex]
	} else {
		recordId = record.Id
	}

	var resp = new(huaweidns.ZonesUpdateRecordSetResponse)
	err = this.doAPI(http.MethodPut, "/v2.1/zones/"+zoneId+"/recordsets/"+recordId, map[string]string{}, maps.Map{
		"name":        newRecord.Name + "." + domain + ".",
		"description": "CDN系统自动创建",
		"type":        newRecord.Type,
		"records":     []string{newRecord.Value},
		"line":        newRecord.Route, // TODO 华为云此API无法修改线路，API地址：https://support.huaweicloud.com/api-dns/dns_api_65006.html
	}, resp)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecord 删除记录
func (this *HuaweiDNSProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return err
	}

	var recordId string
	var atIndex = strings.Index(record.Id, "@")
	if atIndex > 0 {
		recordId = record.Id[:atIndex]
	} else {
		recordId = record.Id
	}

	var resp = new(huaweidns.ZonesDeleteRecordSetResponse)
	err = this.doAPI(http.MethodDelete, "/v2.1/zones/"+zoneId+"/recordsets/"+recordId, map[string]string{}, maps.Map{}, resp)
	if err != nil {
		return err
	}

	return nil
}

// DefaultRoute 默认线路
func (this *HuaweiDNSProvider) DefaultRoute() string {
	return "default_view"
}

func (this *HuaweiDNSProvider) doAPI(method string, apiPath string, args map[string]string, bodyMap maps.Map, respPtr interface{}) error {
	apiURL := HuaweiDNSEndpoint + strings.TrimLeft(apiPath, "/")
	u, err := url.Parse(HuaweiDNSEndpoint)
	if err != nil {
		return err
	}
	apiHost := u.Host
	argStrings := []string{}
	if len(args) > 0 {
		apiURL += "?"
		for k, v := range args {
			argStrings = append(argStrings, k+"="+url.QueryEscape(v))
		}
		apiURL += strings.Join(argStrings, "&")
	}
	sort.Strings(argStrings)
	method = strings.ToUpper(method)

	var bodyReader io.Reader = nil
	var bodyData []byte
	if bodyMap != nil {
		bodyData, err = json.Marshal(bodyMap)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	req, err := http.NewRequest(method, apiURL, bodyReader)
	if err != nil {
		return err
	}

	contentType := "application/json"
	host := apiHost
	datetime := time.Now().UTC().Format("20060102T150405Z")
	if !strings.HasSuffix(apiPath, "/") {
		apiPath += "/"
	}
	canonicalRequest := method + "\n" + apiPath + "\n" + strings.Join(argStrings, "&") + "\ncontent-type:" + contentType + "\nhost:" + host + "\nx-sdk-date:" + datetime + "\n" + "\ncontent-type;host;x-sdk-date"

	h := sha256.New()
	_, err = h.Write(bodyData)
	if err != nil {
		return err
	}
	canonicalRequest += "\n" + fmt.Sprintf("%x", h.Sum(nil))

	h2 := sha256.New()
	_, err = h2.Write([]byte(canonicalRequest))
	if err != nil {
		return err
	}
	source := "SDK-HMAC-SHA256\n" + datetime + "\n" + fmt.Sprintf("%x", h2.Sum(nil))
	h3 := hmac.New(sha256.New, []byte(this.accessKeySecret))
	h3.Write([]byte(source))
	signString := fmt.Sprintf("%x", h3.Sum(nil))
	req.Header.Set("Host", host)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-sdk-date", datetime)
	req.Header.Set("Authorization", "SDK-HMAC-SHA256 Access="+this.accessKeyId+", SignedHeaders=content-type;host;x-sdk-date, Signature="+signString)

	resp, err := huaweiDNSHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 0 {
		return errors.New("invalid response status '" + strconv.Itoa(resp.StatusCode) + "', response '" + string(data) + "'")
	}

	err = json.Unmarshal(data, respPtr)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("response error: status code: " + types.String(resp.StatusCode) + ", response data: " + string(data))
	}

	return nil
}

func (this *HuaweiDNSProvider) findZoneIdWithDomain(domain string) (string, error) {
	var resp = new(huaweidns.ZonesResponse)
	err := this.doAPI(http.MethodGet, "/v2/zones", map[string]string{"name": domain}, nil, resp)
	if err != nil {
		return "", err
	}

	for _, zone := range resp.Zones {
		if zone.Name == domain+"." {
			return zone.Id, nil
		}
	}
	return "", errors.New("can not find zone id for '" + domain + "'")
}
