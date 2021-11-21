package setup

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Setup struct {
	config *Config

	// 要返回的数据
	AdminNodeId     string
	AdminNodeSecret string
}

func NewSetup(config *Config) *Setup {
	return &Setup{
		config: config,
	}
}

func NewSetupFromCmd() *Setup {
	args := cmd.ParseArgs(strings.Join(os.Args[1:], " "))

	config := &Config{}
	for _, arg := range args {
		index := strings.Index(arg, "=")
		if index <= 0 {
			continue
		}
		value := arg[index+1:]
		value = strings.Trim(value, "\"'")
		switch arg[:index] {
		case "-api-node-protocol":
			config.APINodeProtocol = value
		case "-api-node-host":
			config.APINodeHost = value
		case "-api-node-port":
			config.APINodePort = types.Int(value)
		}
	}

	return NewSetup(config)
}

func (this *Setup) Run() error {
	if this.config == nil {
		return errors.New("config should not be nil")
	}

	if len(this.config.APINodeProtocol) == 0 {
		return errors.New("api node protocol should not be empty")
	}
	if this.config.APINodeProtocol != "http" && this.config.APINodeProtocol != "https" {
		return errors.New("invalid api node protocol: " + this.config.APINodeProtocol)
	}
	if len(this.config.APINodeHost) == 0 {
		return errors.New("api node host should not be empty")
	}
	if this.config.APINodePort <= 0 {
		return errors.New("api node port should not be less than 1")
	}

	// 执行SQL
	config := &dbs.Config{}
	configData, err := ioutil.ReadFile(Tea.ConfigFile("db.yaml"))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return err
	}
	for _, db := range config.DBs {
		// 可以同时运行多条语句
		db.Dsn += "&multiStatements=true"
	}
	dbConfig, ok := config.DBs[Tea.Env]
	if !ok {
		return errors.New("can not find database config for env '" + Tea.Env + "'")
	}

	executor := NewSQLExecutor(dbConfig)
	err = executor.Run(false)
	if err != nil {
		return err
	}

	// Admin节点信息
	apiTokenDAO := models.NewApiTokenDAO()
	token, err := apiTokenDAO.FindEnabledTokenWithRole(nil, "admin")
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("can not find admin node token, please run the setup again")
	}
	this.AdminNodeId = token.NodeId
	this.AdminNodeSecret = token.Secret

	// 检查API节点
	dao := models.NewAPINodeDAO()
	apiNodeId, err := dao.FindEnabledAPINodeIdWithAddr(nil, this.config.APINodeProtocol, this.config.APINodeHost, this.config.APINodePort)
	if err != nil {
		return err
	}
	if apiNodeId == 0 {
		addr := &serverconfigs.NetworkAddressConfig{
			Protocol:  serverconfigs.Protocol(this.config.APINodeProtocol),
			Host:      this.config.APINodeHost,
			PortRange: strconv.Itoa(this.config.APINodePort),
		}
		addrsJSON, err := json.Marshal([]*serverconfigs.NetworkAddressConfig{addr})
		if err != nil {
			return errors.New("json encode api node addr failed: " + err.Error())
		}

		var httpJSON []byte = nil
		var httpsJSON []byte = nil
		if this.config.APINodeProtocol == "http" {
			httpConfig := &serverconfigs.HTTPProtocolConfig{}
			httpConfig.IsOn = true
			httpConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "http",
					PortRange: strconv.Itoa(this.config.APINodePort),
				},
			}
			httpJSON, err = json.Marshal(httpConfig)
			if err != nil {
				return errors.New("json encode api node http config failed: " + err.Error())
			}
		}
		if this.config.APINodeProtocol == "https" {
			// TODO 如果在安装过程中开启了HTTPS，需要同时上传SSL证书
			httpsConfig := &serverconfigs.HTTPSProtocolConfig{}
			httpsConfig.IsOn = true
			httpsConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "https",
					PortRange: strconv.Itoa(this.config.APINodePort),
				},
			}
			httpsJSON, err = json.Marshal(httpsConfig)
			if err != nil {
				return errors.New("json encode api node https config failed: " + err.Error())
			}
		}

		// 创建API节点
		nodeId, err := dao.CreateAPINode(nil, "默认API节点", "这是默认创建的第一个API节点", httpJSON, httpsJSON, false, nil, nil, addrsJSON, true)
		if err != nil {
			return errors.New("create api node in database failed: " + err.Error())
		}
		apiNodeId = nodeId
	}

	apiNode, err := dao.FindEnabledAPINode(nil, apiNodeId)
	if err != nil {
		return err
	}
	if apiNode == nil {
		return errors.New("apiNode should not be nil")
	}

	// 保存配置
	apiConfig := &configs.APIConfig{
		NodeId: apiNode.UniqueId,
		Secret: apiNode.Secret,
	}
	err = apiConfig.WriteFile(Tea.ConfigFile("api.yaml"))
	if err != nil {
		return errors.New("save config failed: " + err.Error())
	}

	return nil
}
