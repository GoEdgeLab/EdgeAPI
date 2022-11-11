package setup

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

type Setup struct {
	config *Config

	// 要返回的数据
	AdminNodeId     string
	AdminNodeSecret string

	logFp *os.File
}

func NewSetup(config *Config) *Setup {
	return &Setup{
		config: config,
	}
}

func NewSetupFromCmd() *Setup {
	var args = cmd.ParseArgs(strings.Join(os.Args[1:], " "))

	var config = &Config{}
	for _, arg := range args {
		var index = strings.Index(arg, "=")
		if index <= 0 {
			continue
		}
		var value = arg[index+1:]
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

	var setup = NewSetup(config)

	// log writer
	var tmpDir = os.TempDir()
	if len(tmpDir) > 0 {
		fp, err := os.OpenFile(tmpDir+"/edge-install.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			setup.logFp = fp
		}
	}

	return setup
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
	var config = &dbs.Config{}
	configData, err := os.ReadFile(Tea.ConfigFile("db.yaml"))
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

	var executor = NewSQLExecutor(dbConfig)
	if this.logFp != nil {
		executor.SetLogWriter(this.logFp)

		defer func() {
			_ = this.logFp.Close()
			_ = os.Remove(this.logFp.Name())
		}()
	}
	err = executor.Run(false)
	if err != nil {
		return err
	}

	// Admin节点信息
	var apiTokenDAO = models.NewApiTokenDAO()
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
	var dao = models.NewAPINodeDAO()
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

	apiNode, err := dao.FindEnabledAPINode(nil, apiNodeId, nil)
	if err != nil {
		return err
	}
	if apiNode == nil {
		return errors.New("apiNode should not be nil")
	}

	// 保存配置
	var apiConfig = &configs.APIConfig{
		NodeId: apiNode.UniqueId,
		Secret: apiNode.Secret,
	}
	err = apiConfig.WriteFile(Tea.ConfigFile("api.yaml"))
	if err != nil {
		return errors.New("save config failed: " + err.Error())
	}

	return nil
}
