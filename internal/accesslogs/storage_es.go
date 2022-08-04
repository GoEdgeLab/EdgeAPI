package accesslogs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ESStorage ElasticSearch存储策略
type ESStorage struct {
	BaseStorage

	config *serverconfigs.AccessLogESStorageConfig
}

func NewESStorage(config *serverconfigs.AccessLogESStorageConfig) *ESStorage {
	return &ESStorage{config: config}
}

func (this *ESStorage) Config() interface{} {
	return this.config
}

// Start 开启
func (this *ESStorage) Start() error {
	if len(this.config.Endpoint) == 0 {
		return errors.New("'endpoint' should not be nil")
	}
	if !regexp.MustCompile(`(?i)^(http|https)://`).MatchString(this.config.Endpoint) {
		this.config.Endpoint = "http://" + this.config.Endpoint
	}
	if len(this.config.Index) == 0 {
		return errors.New("'index' should not be nil")
	}
	if len(this.config.MappingType) == 0 {
		return errors.New("'mappingType' should not be nil")
	}
	return nil
}

// 写入日志
func (this *ESStorage) Write(accessLogs []*pb.HTTPAccessLog) error {
	if len(accessLogs) == 0 {
		return nil
	}

	bulk := &strings.Builder{}
	indexName := this.FormatVariables(this.config.Index)
	typeName := this.FormatVariables(this.config.MappingType)
	for _, accessLog := range accessLogs {
		if this.firewallOnly && accessLog.FirewallPolicyId == 0 {
			continue
		}

		if len(accessLog.RequestId) == 0 {
			continue
		}

		opData, err := json.Marshal(map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_type":  typeName,
				"_id":    accessLog.RequestId,
			},
		})
		if err != nil {
			remotelogs.Error("ACCESS_LOG_ES_STORAGE", "write failed: "+err.Error())
			continue
		}

		data, err := this.Marshal(accessLog)
		if err != nil {
			remotelogs.Error("ACCESS_LOG_ES_STORAGE", "marshal data failed: "+err.Error())
			continue
		}

		bulk.Write(opData)
		bulk.WriteString("\n")
		bulk.Write(data)
		bulk.WriteString("\n")
	}

	if bulk.Len() == 0 {
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, this.config.Endpoint+"/_bulk", strings.NewReader(bulk.String()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", strings.ReplaceAll(teaconst.ProductName, " ", "-")+"/"+teaconst.Version)
	if len(this.config.Username) > 0 || len(this.config.Password) > 0 {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(this.config.Username+":"+this.config.Password)))
	}
	client := utils.SharedHttpClient(10 * time.Second)
	defer func() {
		_ = req.Body.Close()
	}()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyData, _ := io.ReadAll(resp.Body)
		return errors.New("ElasticSearch response status code: " + fmt.Sprintf("%d", resp.StatusCode) + " content: " + string(bodyData))
	}

	return nil
}

// Close 关闭
func (this *ESStorage) Close() error {
	return nil
}
