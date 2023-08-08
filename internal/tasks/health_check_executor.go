package tasks

import (
	"context"
	"crypto/tls"
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HealthCheckExecutor struct {
	BaseTask

	clusterId int64
}

func NewHealthCheckExecutor(clusterId int64) *HealthCheckExecutor {
	return &HealthCheckExecutor{clusterId: clusterId}
}

func (this *HealthCheckExecutor) Run() ([]*HealthCheckResult, error) {
	cluster, err := models.NewNodeClusterDAO().FindEnabledNodeCluster(nil, this.clusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil || !cluster.IsOn {
		// 如果节点已经被删除，则不提示错误
		return nil, nil
	}
	if !cluster.HealthCheck.IsNotNull() {
		return nil, errors.New("health check config is not found")
	}

	var healthCheckConfig = &serverconfigs.HealthCheckConfig{}
	err = json.Unmarshal(cluster.HealthCheck, healthCheckConfig)
	if err != nil {
		return nil, err
	}

	var results = []*HealthCheckResult{}

	// 查询集群下的节点
	nodes, err := models.NewNodeDAO().FindAllEnabledNodesWithClusterId(nil, this.clusterId, false)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return results, nil
	}

	var tx *dbs.Tx
	for _, node := range nodes {
		if !node.IsOn || node.IsBackupForCluster || node.IsBackupForGroup || (len(node.OfflineDay) > 0 && node.OfflineDay < timeutil.Format("Ymd")) {
			continue
		}

		ipAddrs, err := models.SharedNodeIPAddressDAO.FindNodeAccessIPAddresses(tx, int64(node.Id), nodeconfigs.NodeRoleNode)
		if err != nil {
			return nil, err
		}
		for _, ipAddr := range ipAddrs {
			var ipClusterIds = ipAddr.DecodeClusterIds()
			if len(ipClusterIds) > 0 && !lists.ContainsInt64(ipClusterIds, this.clusterId) {
				continue
			}

			// TODO 支持备用IP
			var result = &HealthCheckResult{
				Node:       node,
				NodeAddrId: int64(ipAddr.Id),
				NodeAddr:   ipAddr.Ip,
			}

			results = append(results, result)
		}
	}

	// 并行检查
	var preparedResults = []*HealthCheckResult{}
	for _, result := range results {
		if len(result.NodeAddr) > 0 {
			preparedResults = append(preparedResults, result)
		}
	}

	if len(preparedResults) == 0 {
		return results, nil
	}

	var countResults = len(preparedResults)
	var queue = make(chan *HealthCheckResult, countResults)
	for _, result := range preparedResults {
		queue <- result
	}

	var countTries = types.Int(healthCheckConfig.CountTries)
	if countTries > 10 { // 限定最多尝试10次 TODO 应该在管理界面提示用户
		countTries = 10
	}
	if countTries < 1 {
		countTries = 3
	}

	var tryDelay = 1 * time.Second
	if healthCheckConfig.TryDelay != nil {
		tryDelay = healthCheckConfig.TryDelay.Duration()

		if tryDelay > 1*time.Minute { // 最多不能超过1分钟 TODO 应该在管理界面提示用户
			tryDelay = 1 * time.Minute
		}
	}

	var concurrent = 128

	var wg = sync.WaitGroup{}
	wg.Add(countResults)
	for i := 0; i < concurrent; i++ {
		go func() {
			for {
				select {
				case result := <-queue:
					this.runNode(healthCheckConfig, result, countTries, tryDelay)
					wg.Done()
				default:
					return
				}
			}
		}()
	}
	wg.Wait()

	return results, nil
}

func (this *HealthCheckExecutor) runNode(healthCheckConfig *serverconfigs.HealthCheckConfig, result *HealthCheckResult, countTries int, tryDelay time.Duration) {
	for i := 1; i <= countTries; i++ {
		var before = time.Now()
		err := this.runNodeOnce(healthCheckConfig, result)
		result.CostMs = time.Since(before).Seconds() * 1000
		if err != nil {
			result.Error = err.Error()
		}
		if result.IsOk {
			break
		}
		if tryDelay > 0 {
			time.Sleep(tryDelay)
		}
	}

	// 修改节点IP状态
	if teaconst.IsPlus {
		isChanged, err := models.SharedNodeIPAddressDAO.UpdateAddressHealthCount(nil, result.NodeAddrId, result.IsOk, healthCheckConfig.CountUp, healthCheckConfig.CountDown, healthCheckConfig.AutoDown)
		if err != nil {
			this.logErr("HealthCheckExecutor", err.Error())
			return
		}

		if isChanged {
			// 在线状态发生变化
			if healthCheckConfig.AutoDown {
				// 发送消息
				var message string
				var messageType string
				var messageLevel string
				if result.IsOk {
					message = "健康检查成功，节点\"" + result.Node.Name + "\"，IP\"" + result.NodeAddr + "\"已恢复上线"
					messageType = models.MessageTypeHealthCheckNodeUp
					messageLevel = models.MessageLevelSuccess
				} else {
					message = "健康检查失败，节点\"" + result.Node.Name + "\"，IP\"" + result.NodeAddr + "\"已自动下线"
					messageType = models.MessageTypeHealthCheckNodeDown
					messageLevel = models.MessageLevelError
				}

				err = models.NewMessageDAO().CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, int64(result.Node.Id), messageType, messageLevel, message, message, nil, false)
				if err != nil {
					this.logErr("HealthCheckExecutor", err.Error())
					return
				}

				// 触发节点动作
				if !result.IsOk {
					err := this.fireNodeActions(int64(result.Node.Id))
					if err != nil {
						this.logErr("HealthCheckExecutor", err.Error())
					}
					return
				}
			}

			// 触发阈值
			err = models.SharedNodeIPAddressDAO.FireThresholds(nil, nodeconfigs.NodeRoleNode, int64(result.Node.Id))
			if err != nil {
				this.logErr("HealthCheckExecutor", err.Error())
				return
			}
		}

		// 结束处理 ，因为我们只处理IP的上下线，不修改节点的状态
		if healthCheckConfig.AutoDown {
			return
		}
	}

	// 修改节点状态
	if healthCheckConfig.AutoDown {
		isChanged, err := models.SharedNodeDAO.UpdateNodeUpCount(nil, int64(result.Node.Id), result.IsOk, healthCheckConfig.CountUp, healthCheckConfig.CountDown)
		if err != nil {
			this.logErr("HealthCheckExecutor", err.Error())
		} else if isChanged {
			// 通知恢复或下线
			if result.IsOk {
				var message = "健康检查成功，节点\"" + result.Node.Name + "\"已恢复上线"
				err = models.NewMessageDAO().CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, int64(result.Node.Id), models.MessageTypeHealthCheckNodeUp, models.MessageLevelSuccess, message, message, nil, false)
			} else {
				var message = "健康检查失败，节点\"" + result.Node.Name + "\"已自动下线"
				err = models.NewMessageDAO().CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, int64(result.Node.Id), models.MessageTypeHealthCheckNodeDown, models.MessageLevelError, message, message, nil, false)
			}
			if err != nil {
				this.logErr("HealthCheckExecutor", err.Error())
				return
			}
		}
	} else {
		// 通知健康检查结果
		var err error
		if result.IsOk {
			message := "节点\"" + result.Node.Name + "\"健康检查成功"
			err = models.NewMessageDAO().CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, int64(result.Node.Id), models.MessageTypeHealthCheckNodeUp, models.MessageLevelSuccess, message, message, nil, false)
		} else {
			message := "节点\"" + result.Node.Name + "\"健康检查失败"
			err = models.NewMessageDAO().CreateNodeMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, int64(result.Node.Id), models.MessageTypeHealthCheckNodeDown, models.MessageLevelError, message, message, nil, false)
		}
		if err != nil {
			this.logErr("HealthCheckExecutor", err.Error())
			return
		}
	}
}

// 检查单个节点
func (this *HealthCheckExecutor) runNodeOnce(healthCheckConfig *serverconfigs.HealthCheckConfig, result *HealthCheckResult) error {
	// 支持IPv6
	if utils.IsIPv6(result.NodeAddr) {
		result.NodeAddr = configutils.QuoteIP(result.NodeAddr)
	}

	if len(healthCheckConfig.URL) == 0 {
		healthCheckConfig.URL = "http://${host}/"
	}

	var url = strings.ReplaceAll(healthCheckConfig.URL, "${host}", result.NodeAddr)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if len(healthCheckConfig.UserAgent) > 0 {
		req.Header.Set("User-Agent", healthCheckConfig.UserAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	}

	key, err := nodeutils.Base64EncodeMap(maps.Map{
		"onlyBasicRequest": healthCheckConfig.OnlyBasicRequest,
		"accessLogIsOn":    healthCheckConfig.AccessLogIsOn,
	})
	if err != nil {
		return err
	}
	req.Header.Set(serverconfigs.HealthCheckHeaderName, key)

	var timeout = 5 * time.Second
	if healthCheckConfig.Timeout != nil {
		timeout = healthCheckConfig.Timeout.Duration()
	}

	var client = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				_, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				return net.DialTimeout(network, configutils.QuoteIP(result.NodeAddr)+":"+port, timeout)
			},
			MaxIdleConns:          1,
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       1,
			IdleConnTimeout:       10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSHandshakeTimeout:   0,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	defer func() {
		client.CloseIdleConnections()
	}()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if len(healthCheckConfig.StatusCodes) > 0 && !lists.ContainsInt(healthCheckConfig.StatusCodes, resp.StatusCode) {
		result.Error = "invalid response status code '" + strconv.Itoa(resp.StatusCode) + "'"
		return nil
	}

	result.IsOk = true

	return nil
}
