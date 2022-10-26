package installers

import (
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

var sharedNodeQueue = NewNodeQueue()

type NodeQueue struct {
}

func NewNodeQueue() *NodeQueue {
	return &NodeQueue{}
}

func SharedNodeQueue() *NodeQueue {
	return sharedNodeQueue
}

// InstallNodeProcess 安装边缘节点流程控制
func (this *NodeQueue) InstallNodeProcess(nodeId int64, isUpgrading bool) error {
	var installStatus = models.NewNodeInstallStatus()
	installStatus.IsRunning = true
	installStatus.UpdatedAt = time.Now().Unix()

	err := models.SharedNodeDAO.UpdateNodeInstallStatus(nil, nodeId, installStatus)
	if err != nil {
		return err
	}

	// 更新时间
	var ticker = utils.NewTicker(3 * time.Second)
	goman.New(func() {
		for ticker.Wait() {
			installStatus.UpdatedAt = time.Now().Unix()
			err := models.SharedNodeDAO.UpdateNodeInstallStatus(nil, nodeId, installStatus)
			if err != nil {
				logs.Println("[INSTALL]" + err.Error())
				continue
			}
		}
	})
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
	err = models.SharedNodeDAO.UpdateNodeInstallStatus(nil, nodeId, installStatus)
	if err != nil {
		return err
	}

	// 修改为已安装
	if installStatus.IsOk {
		err = models.SharedNodeDAO.UpdateNodeIsInstalled(nil, nodeId, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// InstallNode 安装边缘节点
func (this *NodeQueue) InstallNode(nodeId int64, installStatus *models.NodeInstallStatus, isUpgrading bool) error {
	node, err := models.SharedNodeDAO.FindEnabledNode(nil, nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("can not find node, ID：'" + numberutils.FormatInt64(nodeId) + "'")
	}

	// 登录信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(nil, nodeconfigs.NodeRoleNode, nodeId)
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
		// 查询节点IP
		ip, _, err := models.SharedNodeIPAddressDAO.FindFirstNodeAccessIPAddress(nil, nodeId, false, nodeconfigs.NodeRoleNode)
		if err != nil {
			return err
		}
		if len(ip) > 0 {
			loginParams.Host = ip
		} else {
			installStatus.ErrorCode = "EMPTY_SSH_HOST"
			return errors.New("ssh host should not be empty")
		}
	}

	if loginParams.Port <= 0 {
		// 从集群中读取
		sshParams, err := models.SharedNodeClusterDAO.FindClusterSSHParams(nil, int64(node.ClusterId))
		if err != nil {
			return err
		}
		if sshParams != nil && sshParams.Port > 0 {
			loginParams.Port = sshParams.Port
		} else {
			installStatus.ErrorCode = "EMPTY_SSH_PORT"
			return errors.New("ssh port is invalid")
		}
	}

	if loginParams.GrantId == 0 {
		// 从集群中读取
		grantId, err := models.SharedNodeClusterDAO.FindClusterGrantId(nil, int64(node.ClusterId))
		if err != nil {
			return err
		}
		if grantId == 0 {
			installStatus.ErrorCode = "EMPTY_GRANT"
			return errors.New("can not find node grant")
		}
		loginParams.GrantId = grantId
	}
	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(nil, loginParams.GrantId)
	if err != nil {
		return err
	}
	if grant == nil {
		installStatus.ErrorCode = "EMPTY_GRANT"
		return errors.New("can not find user grant with id '" + numberutils.FormatInt64(loginParams.GrantId) + "'")
	}

	// API终端
	apiNodes, err := models.SharedAPINodeDAO.FindAllEnabledAndOnAPINodes(nil)
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

	var installer = &NodeInstaller{}
	err = installer.Login(&Credentials{
		Host:       loginParams.Host,
		Port:       loginParams.Port,
		Username:   grant.Username,
		Password:   grant.Password,
		PrivateKey: grant.PrivateKey,
		Passphrase: grant.Passphrase,
		Method:     grant.Method,
		Sudo:       grant.Su == 1,
	})
	if err != nil {
		installStatus.ErrorCode = "SSH_LOGIN_FAILED"
		return err
	}
	defer func() {
		_ = installer.Close()
	}()

	// 安装目录
	installDir := node.InstallDir
	if len(installDir) == 0 {
		clusterId := node.ClusterId
		cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(nil, int64(clusterId))
		if err != nil {
			return err
		}
		if cluster == nil {
			return errors.New("can not find cluster, ID：'" + fmt.Sprintf("%d", clusterId) + "'")
		}
		installDir = cluster.InstallDir
		if len(installDir) == 0 {
			// 默认是 $登录用户/edge-node
			installDir = installer.client.UserHome() + "/edge-node"
		}
	}

	err = installer.Install(installDir, params, installStatus)
	return err
}

// StartNode 启动边缘节点
func (this *NodeQueue) StartNode(nodeId int64) error {
	node, err := models.SharedNodeDAO.FindEnabledNode(nil, nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("can not find node, ID：'" + numberutils.FormatInt64(nodeId) + "'")
	}

	// 登录信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(nil, nodeconfigs.NodeRoleNode, nodeId)
	if err != nil {
		return err
	}
	if login == nil {
		return newGrantError("can not find node login information")
	}
	loginParams, err := login.DecodeSSHParams()
	if err != nil {
		return newGrantError(err.Error())
	}

	if len(loginParams.Host) == 0 {
		// 查询节点IP
		ip, _, err := models.SharedNodeIPAddressDAO.FindFirstNodeAccessIPAddress(nil, nodeId, false, nodeconfigs.NodeRoleNode)
		if err != nil {
			return err
		}
		if len(ip) > 0 {
			loginParams.Host = ip
		} else {
			return newGrantError("ssh host should not be empty")
		}
	}

	if loginParams.Port <= 0 {
		// 从集群中读取
		sshParams, err := models.SharedNodeClusterDAO.FindClusterSSHParams(nil, int64(node.ClusterId))
		if err != nil {
			return err
		}
		if sshParams != nil && sshParams.Port > 0 {
			loginParams.Port = sshParams.Port
		} else {
			return newGrantError("ssh port is invalid")
		}
	}

	if loginParams.GrantId == 0 {
		// 从集群中读取
		grantId, err := models.SharedNodeClusterDAO.FindClusterGrantId(nil, int64(node.ClusterId))
		if err != nil {
			return err
		}
		if grantId == 0 {
			return newGrantError("can not find node grant")
		}
		loginParams.GrantId = grantId
	}
	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(nil, loginParams.GrantId)
	if err != nil {
		return err
	}
	if grant == nil {
		return newGrantError("can not find user grant with id '" + numberutils.FormatInt64(loginParams.GrantId) + "'")
	}

	var installer = &NodeInstaller{}
	err = installer.Login(&Credentials{
		Host:       loginParams.Host,
		Port:       loginParams.Port,
		Username:   grant.Username,
		Password:   grant.Password,
		PrivateKey: grant.PrivateKey,
		Passphrase: grant.Passphrase,
		Method:     grant.Method,
		Sudo:       grant.Su == 1,
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = installer.Close()
	}()

	// 检查命令是否存在
	exe, err := this.lookupNodeExe(node, installer.client)
	if err != nil {
		return errors.New("edge node was not installed correctly, can not find executable file")
	}

	// 我们先尝试Systemd启动
	_, _, _ = installer.client.Exec("systemctl start edge-node")

	// 执行start
	_, stderr, err := installer.client.Exec("sudo " + exe + " start")
	if err != nil {
		return errors.New("start failed: " + err.Error())
	}
	if len(stderr) > 0 {
		return errors.New("start failed: " + stderr)
	}

	return nil
}

// StopNode 停止节点
func (this *NodeQueue) StopNode(nodeId int64) error {
	node, err := models.SharedNodeDAO.FindEnabledNode(nil, nodeId)
	if err != nil {
		return err
	}
	if node == nil {
		return errors.New("can not find node, ID：'" + numberutils.FormatInt64(nodeId) + "'")
	}

	// 登录信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(nil, nodeconfigs.NodeRoleNode, nodeId)
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
		// 查询节点IP
		ip, _, err := models.SharedNodeIPAddressDAO.FindFirstNodeAccessIPAddress(nil, nodeId, false, nodeconfigs.NodeRoleNode)
		if err != nil {
			return err
		}
		if len(ip) > 0 {
			loginParams.Host = ip
		} else {
			return errors.New("ssh host should not be empty")
		}
	}

	if loginParams.Port <= 0 {
		// 从集群中读取
		sshParams, err := models.SharedNodeClusterDAO.FindClusterSSHParams(nil, int64(node.ClusterId))
		if err != nil {
			return err
		}
		if sshParams != nil && sshParams.Port > 0 {
			loginParams.Port = sshParams.Port
		} else {
			return errors.New("ssh port is invalid")
		}
	}

	if loginParams.GrantId == 0 {
		// 从集群中读取
		grantId, err := models.SharedNodeClusterDAO.FindClusterGrantId(nil, int64(node.ClusterId))
		if err != nil {
			return err
		}
		if grantId == 0 {
			return errors.New("can not find node grant")
		}
		loginParams.GrantId = grantId
	}
	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(nil, loginParams.GrantId)
	if err != nil {
		return err
	}
	if grant == nil {
		return errors.New("can not find user grant with id '" + numberutils.FormatInt64(loginParams.GrantId) + "'")
	}

	var installer = &NodeInstaller{}
	err = installer.Login(&Credentials{
		Host:       loginParams.Host,
		Port:       loginParams.Port,
		Username:   grant.Username,
		Password:   grant.Password,
		PrivateKey: grant.PrivateKey,
		Passphrase: grant.Passphrase,
		Method:     grant.Method,
		Sudo:       grant.Su == 1,
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = installer.Close()
	}()

	// 检查命令是否存在
	exe, err := this.lookupNodeExe(node, installer.client)
	if err != nil {
		return errors.New("edge node was not installed correctly, can not find executable file")
	}

	// 我们先尝试Systemd停止
	_, _, _ = installer.client.Exec("/usr/bin/systemctl stop edge-node")

	// 执行stop
	_, stderr, err := installer.client.Exec(exe + " stop")
	if err != nil {
		return errors.New("stop failed: " + err.Error())
	}
	if len(stderr) > 0 {
		return errors.New("stop failed: " + stderr)
	}

	return nil
}

func (this *NodeQueue) lookupNodeExe(node *models.Node, client *SSHClient) (string, error) {
	// 安装目录
	var nodeDirs = []string{}
	if len(node.InstallDir) > 0 {
		nodeDirs = append(nodeDirs, node.InstallDir)
	}
	var clusterId = node.ClusterId
	cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(nil, int64(clusterId))
	if err != nil {
		return "", err
	}
	if cluster == nil {
		return "", errors.New("can not find cluster, ID：'" + fmt.Sprintf("%d", clusterId) + "'")
	}
	if len(cluster.InstallDir) > 0 {
		nodeDirs = append(nodeDirs, cluster.InstallDir)
	}

	// 默认是 $登录用户/edge-node
	nodeDirs = append(nodeDirs, client.UserHome()+"/edge-node")

	// edge-boot安装目录
	nodeDirs = append(nodeDirs, "/usr/local/goedge")

	for _, dir := range nodeDirs {
		var path = dir + "/edge-node/bin/edge-node"
		_, err := client.sftp.Stat(path)
		if err == nil {
			return path, nil
		}
	}

	return "", nil
}
