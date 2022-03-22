// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/logs"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io"
	"path/filepath"
)

// NSNodeService 域名服务器节点服务
type NSNodeService struct {
	services.BaseService
}

// FindAllEnabledNSNodesWithNSClusterId 根据集群查找所有节点
func (this *NSNodeService) FindAllEnabledNSNodesWithNSClusterId(ctx context.Context, req *pb.FindAllEnabledNSNodesWithNSClusterIdRequest) (*pb.FindAllEnabledNSNodesWithNSClusterIdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodes, err := models.SharedNSNodeDAO.FindAllEnabledNodesWithClusterId(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	pbNodes := []*pb.NSNode{}
	for _, node := range nodes {
		pbNodes = append(pbNodes, &pb.NSNode{
			Id:          int64(node.Id),
			Name:        node.Name,
			IsOn:        node.IsOn,
			UniqueId:    node.UniqueId,
			Secret:      node.Secret,
			IsInstalled: node.IsInstalled,
			InstallDir:  node.InstallDir,
			IsUp:        node.IsUp,
			NsCluster:   nil,
		})
	}
	return &pb.FindAllEnabledNSNodesWithNSClusterIdResponse{NsNodes: pbNodes}, nil
}

// CountAllEnabledNSNodes 所有可用的节点数量
func (this *NSNodeService) CountAllEnabledNSNodes(ctx context.Context, req *pb.CountAllEnabledNSNodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNSNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountAllEnabledNSNodesMatch 计算匹配的节点数量
func (this *NSNodeService) CountAllEnabledNSNodesMatch(ctx context.Context, req *pb.CountAllEnabledNSNodesMatchRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNSNodeDAO.CountAllEnabledNodesMatch(tx, req.NsClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledNSNodesMatch 列出单页节点
func (this *NSNodeService) ListEnabledNSNodesMatch(ctx context.Context, req *pb.ListEnabledNSNodesMatchRequest) (*pb.ListEnabledNSNodesMatchResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	nodes, err := models.SharedNSNodeDAO.ListAllEnabledNodesMatch(tx, req.NsClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbNodes := []*pb.NSNode{}
	for _, node := range nodes {
		// 安装信息
		installStatus, err := node.DecodeInstallStatus()
		if err != nil {
			return nil, err
		}
		installStatusResult := &pb.NodeInstallStatus{}
		if installStatus != nil {
			installStatusResult = &pb.NodeInstallStatus{
				IsRunning:  installStatus.IsRunning,
				IsFinished: installStatus.IsFinished,
				IsOk:       installStatus.IsOk,
				Error:      installStatus.Error,
				ErrorCode:  installStatus.ErrorCode,
				UpdatedAt:  installStatus.UpdatedAt,
			}
		}

		pbNodes = append(pbNodes, &pb.NSNode{
			Id:            int64(node.Id),
			Name:          node.Name,
			IsOn:          node.IsOn,
			UniqueId:      node.UniqueId,
			Secret:        node.Secret,
			IsActive:      node.IsActive,
			IsInstalled:   node.IsInstalled,
			InstallDir:    node.InstallDir,
			IsUp:          node.IsUp,
			StatusJSON:    node.Status,
			InstallStatus: installStatusResult,
			NsCluster:     nil,
		})
	}
	return &pb.ListEnabledNSNodesMatchResponse{NsNodes: pbNodes}, nil
}

// CountAllUpgradeNSNodesWithNSClusterId 计算需要升级的节点数量
func (this *NSNodeService) CountAllUpgradeNSNodesWithNSClusterId(ctx context.Context, req *pb.CountAllUpgradeNSNodesWithNSClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	deployFiles := installers.SharedDeployManager.LoadNSNodeFiles()
	total := int64(0)
	for _, deployFile := range deployFiles {
		count, err := models.SharedNSNodeDAO.CountAllLowerVersionNodesWithClusterId(tx, req.NsClusterId, deployFile.OS, deployFile.Arch, deployFile.Version)
		if err != nil {
			return nil, err
		}
		total += count
	}

	return this.SuccessCount(total)
}

// CreateNSNode 创建节点
func (this *NSNodeService) CreateNSNode(ctx context.Context, req *pb.CreateNSNodeRequest) (*pb.CreateNSNodeResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedNSNodeDAO.CreateNode(tx, adminId, req.Name, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	// 增加认证相关
	if req.NodeLogin != nil {
		_, err = models.SharedNodeLoginDAO.CreateNodeLogin(tx, nodeconfigs.NodeRoleDNS, nodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
		if err != nil {
			return nil, err
		}
	}

	return &pb.CreateNSNodeResponse{
		NsNodeId: nodeId,
	}, nil
}

// DeleteNSNode 删除节点
func (this *NSNodeService) DeleteNSNode(ctx context.Context, req *pb.DeleteNSNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNSNodeDAO.DisableNSNode(tx, req.NsNodeId)
	if err != nil {
		return nil, err
	}

	// 删除任务
	err = models.SharedNodeTaskDAO.DeleteNodeTasks(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledNSNode 查询单个节点信息
func (this *NSNodeService) FindEnabledNSNode(ctx context.Context, req *pb.FindEnabledNSNodeRequest) (*pb.FindEnabledNSNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedNSNodeDAO.FindEnabledNSNode(tx, req.NsNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindEnabledNSNodeResponse{NsNode: nil}, nil
	}

	// 集群信息
	clusterName, err := models.SharedNSClusterDAO.FindEnabledNSClusterName(tx, int64(node.ClusterId))
	if err != nil {
		return nil, err
	}

	// 认证信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId)
	if err != nil {
		return nil, err
	}
	var respLogin *pb.NodeLogin = nil
	if login != nil {
		respLogin = &pb.NodeLogin{
			Id:     int64(login.Id),
			Name:   login.Name,
			Type:   login.Type,
			Params: login.Params,
		}
	}

	// 安装信息
	installStatus, err := node.DecodeInstallStatus()
	if err != nil {
		return nil, err
	}
	installStatusResult := &pb.NodeInstallStatus{}
	if installStatus != nil {
		installStatusResult = &pb.NodeInstallStatus{
			IsRunning:  installStatus.IsRunning,
			IsFinished: installStatus.IsFinished,
			IsOk:       installStatus.IsOk,
			Error:      installStatus.Error,
			ErrorCode:  installStatus.ErrorCode,
			UpdatedAt:  installStatus.UpdatedAt,
		}
	}

	return &pb.FindEnabledNSNodeResponse{NsNode: &pb.NSNode{
		Id:          int64(node.Id),
		Name:        node.Name,
		StatusJSON:  node.Status,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		IsInstalled: node.IsInstalled,
		InstallDir:  node.InstallDir,
		NsCluster: &pb.NSCluster{
			Id:   int64(node.ClusterId),
			Name: clusterName,
		},
		InstallStatus: installStatusResult,
		IsOn:          node.IsOn,
		IsActive:      node.IsActive,
		NodeLogin:     respLogin,
	}}, nil
}

// UpdateNSNode 修改节点
func (this *NSNodeService) UpdateNSNode(ctx context.Context, req *pb.UpdateNSNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNSNodeDAO.UpdateNode(tx, req.NsNodeId, req.Name, req.NsClusterId, req.IsOn)
	if err != nil {
		return nil, err
	}

	// 登录信息
	if req.NodeLogin == nil {
		err = models.SharedNodeLoginDAO.DisableNodeLogins(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId)
		if err != nil {
			return nil, err
		}
	} else {
		if req.NodeLogin.Id > 0 {
			err = models.SharedNodeLoginDAO.UpdateNodeLogin(tx, req.NodeLogin.Id, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = models.SharedNodeLoginDAO.CreateNodeLogin(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// InstallNSNode 安装节点
func (this *NSNodeService) InstallNSNode(ctx context.Context, req *pb.InstallNSNodeRequest) (*pb.InstallNSNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	goman.New(func() {
		err = installers.SharedNSNodeQueue().InstallNodeProcess(req.NsNodeId, false)
		if err != nil {
			logs.Println("[RPC]install dns node:" + err.Error())
		}
	})

	return &pb.InstallNSNodeResponse{}, nil
}

// FindNSNodeInstallStatus 读取节点安装状态
func (this *NSNodeService) FindNSNodeInstallStatus(ctx context.Context, req *pb.FindNSNodeInstallStatusRequest) (*pb.FindNSNodeInstallStatusResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	installStatus, err := models.SharedNSNodeDAO.FindNodeInstallStatus(tx, req.NsNodeId)
	if err != nil {
		return nil, err
	}
	if installStatus == nil {
		return &pb.FindNSNodeInstallStatusResponse{InstallStatus: nil}, nil
	}

	pbInstallStatus := &pb.NodeInstallStatus{
		IsRunning:  installStatus.IsRunning,
		IsFinished: installStatus.IsFinished,
		IsOk:       installStatus.IsOk,
		Error:      installStatus.Error,
		ErrorCode:  installStatus.ErrorCode,
		UpdatedAt:  installStatus.UpdatedAt,
	}
	return &pb.FindNSNodeInstallStatusResponse{InstallStatus: pbInstallStatus}, nil
}

// UpdateNSNodeIsInstalled 修改节点安装状态
func (this *NSNodeService) UpdateNSNodeIsInstalled(ctx context.Context, req *pb.UpdateNSNodeIsInstalledRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNSNodeDAO.UpdateNodeIsInstalled(tx, req.NsNodeId, req.IsInstalled)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNSNodeStatus 更新节点状态
func (this *NSNodeService) UpdateNSNodeStatus(ctx context.Context, req *pb.UpdateNSNodeStatusRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	if req.NodeId > 0 {
		nodeId = req.NodeId
	}

	if nodeId <= 0 {
		return nil, errors.New("'nodeId' should be greater than 0")
	}

	tx := this.NullTx()

	err = models.SharedNSNodeDAO.UpdateNodeStatus(tx, nodeId, req.StatusJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindCurrentNSNodeConfig 获取当前节点信息
func (this *NSNodeService) FindCurrentNSNodeConfig(ctx context.Context, req *pb.FindCurrentNSNodeConfigRequest) (*pb.FindCurrentNSNodeConfigResponse, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	config, err := models.SharedNSNodeDAO.ComposeNodeConfig(tx, nodeId)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return &pb.FindCurrentNSNodeConfigResponse{NsNodeJSON: nil}, nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindCurrentNSNodeConfigResponse{NsNodeJSON: configJSON}, nil
}

// CheckNSNodeLatestVersion 检查新版本
func (this *NSNodeService) CheckNSNodeLatestVersion(ctx context.Context, req *pb.CheckNSNodeLatestVersionRequest) (*pb.CheckNSNodeLatestVersionResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	deployFiles := installers.SharedDeployManager.LoadNSNodeFiles()
	for _, file := range deployFiles {
		if file.OS == req.Os && file.Arch == req.Arch && stringutil.VersionCompare(file.Version, req.CurrentVersion) > 0 {
			return &pb.CheckNSNodeLatestVersionResponse{
				HasNewVersion: true,
				NewVersion:    file.Version,
			}, nil
		}
	}
	return &pb.CheckNSNodeLatestVersionResponse{HasNewVersion: false}, nil
}

// DownloadNSNodeInstallationFile 下载最新DNS节点安装文件
func (this *NSNodeService) DownloadNSNodeInstallationFile(ctx context.Context, req *pb.DownloadNSNodeInstallationFileRequest) (*pb.DownloadNSNodeInstallationFileResponse, error) {
	_, err := this.ValidateNSNode(ctx)
	if err != nil {
		return nil, err
	}

	file := installers.SharedDeployManager.FindNSNodeFile(req.Os, req.Arch)
	if file == nil {
		return &pb.DownloadNSNodeInstallationFileResponse{}, nil
	}

	sum, err := file.Sum()
	if err != nil {
		return nil, err
	}

	data, offset, err := file.Read(req.ChunkOffset)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &pb.DownloadNSNodeInstallationFileResponse{
		Sum:       sum,
		Offset:    offset,
		ChunkData: data,
		Version:   file.Version,
		Filename:  filepath.Base(file.Path),
	}, nil
}

// UpdateNSNodeConnectedAPINodes 更改节点连接的API节点信息
func (this *NSNodeService) UpdateNSNodeConnectedAPINodes(ctx context.Context, req *pb.UpdateNSNodeConnectedAPINodesRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, _, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNSNodeDAO.UpdateNodeConnectedAPINodes(tx, nodeId, req.ApiNodeIds)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return this.Success()
}

// UpdateNSNodeLogin 修改节点登录信息
func (this *NSNodeService) UpdateNSNodeLogin(ctx context.Context, req *pb.UpdateNSNodeLoginRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if req.NodeLogin.Id <= 0 {
		_, err := models.SharedNodeLoginDAO.CreateNodeLogin(tx, nodeconfigs.NodeRoleDNS, req.NsNodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedNodeLoginDAO.UpdateNodeLogin(tx, req.NodeLogin.Id, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)

	return this.Success()
}

// StartNSNode 启动节点
func (this *NSNodeService) StartNSNode(ctx context.Context, req *pb.StartNSNodeRequest) (*pb.StartNSNodeResponse, error) {
	// 校验节点
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	err = installers.SharedNSNodeQueue().StartNode(req.NsNodeId)
	if err != nil {
		return &pb.StartNSNodeResponse{
			IsOk:  false,
			Error: err.Error(),
		}, nil
	}

	// 修改状态
	var tx = this.NullTx()
	err = models.SharedNSNodeDAO.UpdateNodeActive(tx, req.NsNodeId, true)
	if err != nil {
		return nil, err
	}

	return &pb.StartNSNodeResponse{IsOk: true}, nil
}

// StopNSNode 停止节点
func (this *NSNodeService) StopNSNode(ctx context.Context, req *pb.StopNSNodeRequest) (*pb.StopNSNodeResponse, error) {
	// 校验节点
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	err = installers.SharedNSNodeQueue().StopNode(req.NsNodeId)
	if err != nil {
		return &pb.StopNSNodeResponse{
			IsOk:  false,
			Error: err.Error(),
		}, nil
	}

	// 修改状态
	var tx = this.NullTx()
	err = models.SharedNSNodeDAO.UpdateNodeActive(tx, req.NsNodeId, false)
	if err != nil {
		return nil, err
	}

	return &pb.StopNSNodeResponse{IsOk: true}, nil
}
