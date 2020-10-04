package installers

import (
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/logs"
	"strconv"
	"time"
)

var sharedQueue = NewQueue()

type Queue struct {
}

func NewQueue() *Queue {
	return &Queue{}
}

func SharedQueue() *Queue {
	return sharedQueue
}

// 安装边缘节点流程控制
func (this *Queue) InstallNodeProcess(nodeId int64) error {
	installStatus := models.NewNodeInstallStatus()
	installStatus.IsRunning = true
	installStatus.UpdatedAt = time.Now().Unix()

	err := models.SharedNodeDAO.UpdateNodeInstallStatus(nodeId, installStatus)
	if err != nil {
		return err
	}

	// 更新时间
	ticker := utils.NewTicker(3 * time.Second)
	go func() {
		for ticker.Wait() {
			installStatus.UpdatedAt = time.Now().Unix()
			err := models.SharedNodeDAO.UpdateNodeInstallStatus(nodeId, installStatus)
			if err != nil {
				logs.Println("[INSTALL]" + err.Error())
				continue
			}
		}
	}()
	defer func() {
		ticker.Stop()
	}()

	// 开始安装
	err = this.InstallNode(nodeId)

	// 安装结束
	installStatus.IsRunning = false
	installStatus.IsFinished = true
	if err != nil {
		installStatus.Error = err.Error()
	} else {
		installStatus.IsOk = true
	}
	err = models.SharedNodeDAO.UpdateNodeInstallStatus(nodeId, installStatus)
	if err != nil {
		return err
	}

	// 修改为已安装
	if installStatus.IsOk {
		err = models.SharedNodeDAO.UpdateNodeIsInstalled(nodeId, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// 安装边缘节点
func (this *Queue) InstallNode(nodeId int64) error {
	node, err := models.SharedNodeDAO.FindEnabledNode(nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("can not find node, ID：'" + strconv.FormatInt(nodeId, 10) + "'")
	}

	// 登录信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(nodeId)
	if err != nil {
		return err
	}
	if login == nil {
		return errors.New("can not find node login information")
	}
	loginParams, err := login.DecodeSSHParams()
	if err != nil {
		return err
	}

	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(loginParams.GrantId)
	if err != nil {
		return err
	}
	if grant == nil {
		return errors.New("can not find user grant with id '" + strconv.FormatInt(loginParams.GrantId, 10) + "'")
	}

	// 安装目录
	installDir := node.InstallDir
	if len(installDir) == 0 {
		clusterId := node.ClusterId
		cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(int64(clusterId))
		if err != nil {
			return err
		}
		if cluster == nil {
			return errors.New("can not find cluster, ID：'" + fmt.Sprintf("%d", clusterId) + "'")
		}
		installDir = cluster.InstallDir
		if len(installDir) == 0 {
			return errors.New("unable to find installation dir")
		}
	}

	// API终端
	apiNodes, err := models.SharedAPINodeDAO.FindAllEnabledAPINodes()
	if err != nil {
		return err
	}
	if len(apiNodes) == 0 {
		return errors.New("no available api nodes")
	}

	apiEndpoints := []string{}
	for _, apiNode := range apiNodes {
		addrConfigs, err := apiNode.DecodeAccessAddrs()
		if err != nil {
			return errors.New("decode api node access addresses failed: " + err.Error())
		}
		for _, addrConfig := range addrConfigs {
			apiEndpoints = append(apiEndpoints, addrConfig.FullAddresses()...)
		}
	}

	params := &NodeParams{
		Endpoints: apiEndpoints,
		NodeId:    node.UniqueId,
		Secret:    node.Secret,
	}

	installer := &NodeInstaller{}
	err = installer.Login(&Credentials{
		Host:       loginParams.Host,
		Port:       loginParams.Port,
		Username:   grant.Username,
		Password:   grant.Password,
		PrivateKey: grant.PrivateKey,
	})
	if err != nil {
		return err
	}

	err = installer.Install(installDir, params)
	return err
}
