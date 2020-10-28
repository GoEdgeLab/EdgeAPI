package installers

import (
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/iwind/TeaGo/logs"
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
func (this *Queue) InstallNodeProcess(nodeId int64, isUpgrading bool) error {
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
	err = this.InstallNode(nodeId, installStatus, isUpgrading)

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
func (this *Queue) InstallNode(nodeId int64, installStatus *models.NodeInstallStatus, isUpgrading bool) error {
	node, err := models.SharedNodeDAO.FindEnabledNode(nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("can not find node, ID：'" + numberutils.FormatInt64(nodeId) + "'")
	}

	// 登录信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(nodeId)
	if err != nil {
		return err
	}
	if login == nil {
		installStatus.ErrorCode = "EMPTY_LOGIN"
		return errors.New("can not find node login information")
	}
	loginParams, err := login.DecodeSSHParams()
	if err != nil {
		return err
	}

	if len(loginParams.Host) == 0 {
		installStatus.ErrorCode = "EMPTY_SSH_HOST"
		return errors.New("ssh host should not be empty")
	}

	if loginParams.Port <= 0 {
		installStatus.ErrorCode = "EMPTY_SSH_PORT"
		return errors.New("ssh port is invalid")
	}

	if loginParams.GrantId == 0 {
		// 从集群中读取
		grantId, err := models.SharedNodeClusterDAO.FindClusterGrantId(int64(node.ClusterId))
		if err != nil {
			return err
		}
		if grantId == 0 {
			installStatus.ErrorCode = "EMPTY_GRANT"
			return errors.New("can not find node grant")
		}
		loginParams.GrantId = grantId
	}
	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(loginParams.GrantId)
	if err != nil {
		return err
	}
	if grant == nil {
		installStatus.ErrorCode = "EMPTY_GRANT"
		return errors.New("can not find user grant with id '" + numberutils.FormatInt64(loginParams.GrantId) + "'")
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
			// 默认是 $登录用户/edge-node
			installDir = "/" + grant.Username + "/edge-node"
		}
	}

	// API终端
	apiNodes, err := models.SharedAPINodeDAO.FindAllEnabledAndOnAPINodes()
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
		Endpoints:   apiEndpoints,
		NodeId:      node.UniqueId,
		Secret:      node.Secret,
		IsUpgrading: isUpgrading,
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
	defer func() {
		_ = installer.Close()
	}()

	err = installer.Install(installDir, params, installStatus)
	return err
}

// 启动边缘节点
func (this *Queue) StartNode(nodeId int64) error {
	node, err := models.SharedNodeDAO.FindEnabledNode(nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("can not find node, ID：'" + numberutils.FormatInt64(nodeId) + "'")
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

	if len(loginParams.Host) == 0 {
		return errors.New("ssh host should not be empty")
	}

	if loginParams.Port <= 0 {
		return errors.New("ssh port is invalid")
	}

	if loginParams.GrantId == 0 {
		// 从集群中读取
		grantId, err := models.SharedNodeClusterDAO.FindClusterGrantId(int64(node.ClusterId))
		if err != nil {
			return err
		}
		if grantId == 0 {
			return errors.New("can not find node grant")
		}
		loginParams.GrantId = grantId
	}
	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(loginParams.GrantId)
	if err != nil {
		return err
	}
	if grant == nil {
		return errors.New("can not find user grant with id '" + numberutils.FormatInt64(loginParams.GrantId) + "'")
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
			// 默认是 $登录用户/edge-node
			installDir = "/" + grant.Username + "/edge-node"
		}
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
	defer func() {
		_ = installer.Close()
	}()

	// 检查命令是否存在
	exeFile := installDir + "/edge-node/bin/edge-node"
	_, err = installer.client.Stat(exeFile)
	if err != nil {
		return errors.New("edge node is not installed correctly, can not find executable file: " + exeFile)
	}

	_, stderr, err := installer.client.Exec(exeFile + " start")
	if err != nil {
		return errors.New("start failed: " + err.Error())
	}
	if len(stderr) > 0 {
		return errors.New("start failed: " + stderr)
	}

	return nil
}

// 停止节点
func (this *Queue) StopNode(nodeId int64) error {
	node, err := models.SharedNodeDAO.FindEnabledNode(nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("can not find node, ID：'" + numberutils.FormatInt64(nodeId) + "'")
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

	if len(loginParams.Host) == 0 {
		return errors.New("ssh host should not be empty")
	}

	if loginParams.Port <= 0 {
		return errors.New("ssh port is invalid")
	}

	if loginParams.GrantId == 0 {
		// 从集群中读取
		grantId, err := models.SharedNodeClusterDAO.FindClusterGrantId(int64(node.ClusterId))
		if err != nil {
			return err
		}
		if grantId == 0 {
			return errors.New("can not find node grant")
		}
		loginParams.GrantId = grantId
	}
	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(loginParams.GrantId)
	if err != nil {
		return err
	}
	if grant == nil {
		return errors.New("can not find user grant with id '" + numberutils.FormatInt64(loginParams.GrantId) + "'")
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
			// 默认是 $登录用户/edge-node
			installDir = "/" + grant.Username + "/edge-node"
		}
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
	defer func() {
		_ = installer.Close()
	}()

	// 检查命令是否存在
	exeFile := installDir + "/edge-node/bin/edge-node"
	_, err = installer.client.Stat(exeFile)
	if err != nil {
		return errors.New("edge node is not installed correctly, can not find executable file: " + exeFile)
	}

	_, stderr, err := installer.client.Exec(exeFile + " stop")
	if err != nil {
		return errors.New("start failed: " + err.Error())
	}
	if len(stderr) > 0 {
		return errors.New("start failed: " + stderr)
	}

	return nil
}
