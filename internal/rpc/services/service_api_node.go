package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type APINodeService struct {
	BaseService
}

// CreateAPINode 创建API节点
func (this *APINodeService) CreateAPINode(ctx context.Context, req *pb.CreateAPINodeRequest) (*pb.CreateAPINodeResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodeId, err := models.SharedAPINodeDAO.CreateAPINode(tx, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.RestIsOn, req.RestHTTPJSON, req.RestHTTPSJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateAPINodeResponse{ApiNodeId: nodeId}, nil
}

// UpdateAPINode 修改API节点
func (this *APINodeService) UpdateAPINode(ctx context.Context, req *pb.UpdateAPINodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedAPINodeDAO.UpdateAPINode(tx, req.ApiNodeId, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.RestIsOn, req.RestHTTPJSON, req.RestHTTPSJSON, req.AccessAddrsJSON, req.IsOn, req.IsPrimary)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteAPINode 删除API节点
func (this *APINodeService) DeleteAPINode(ctx context.Context, req *pb.DeleteAPINodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedAPINodeDAO.DisableAPINode(tx, req.ApiNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledAPINodes 列出所有可用API节点
func (this *APINodeService) FindAllEnabledAPINodes(ctx context.Context, req *pb.FindAllEnabledAPINodesRequest) (*pb.FindAllEnabledAPINodesResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser, rpcutils.UserTypeNode, rpcutils.UserTypeMonitor, rpcutils.UserTypeDNS, rpcutils.UserTypeAuthority)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodes, err := models.SharedAPINodeDAO.FindAllEnabledAPINodes(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.APINode{}
	for _, node := range nodes {
		accessAddrs, err := node.DecodeAccessAddrStrings()
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.APINode{
			Id:              int64(node.Id),
			IsOn:            node.IsOn,
			NodeClusterId:   int64(node.ClusterId),
			UniqueId:        node.UniqueId,
			Secret:          node.Secret,
			Name:            node.Name,
			Description:     node.Description,
			HttpJSON:        node.Http,
			HttpsJSON:       node.Https,
			AccessAddrsJSON: node.AccessAddrs,
			AccessAddrs:     accessAddrs,
			IsPrimary:       node.IsPrimary,
		})
	}

	return &pb.FindAllEnabledAPINodesResponse{ApiNodes: result}, nil
}

// CountAllEnabledAPINodes 计算API节点数量
func (this *APINodeService) CountAllEnabledAPINodes(ctx context.Context, req *pb.CountAllEnabledAPINodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedAPINodeDAO.CountAllEnabledAPINodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// CountAllEnabledAndOnAPINodes 计算API节点数量
func (this *APINodeService) CountAllEnabledAndOnAPINodes(ctx context.Context, req *pb.CountAllEnabledAndOnAPINodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedAPINodeDAO.CountAllEnabledAndOnAPINodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledAPINodes 列出单页的API节点
func (this *APINodeService) ListEnabledAPINodes(ctx context.Context, req *pb.ListEnabledAPINodesRequest) (*pb.ListEnabledAPINodesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodes, err := models.SharedAPINodeDAO.ListEnabledAPINodes(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.APINode{}
	for _, node := range nodes {
		accessAddrs, err := node.DecodeAccessAddrStrings()
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.APINode{
			Id:              int64(node.Id),
			IsOn:            node.IsOn,
			NodeClusterId:   int64(node.ClusterId),
			UniqueId:        node.UniqueId,
			Secret:          node.Secret,
			Name:            node.Name,
			Description:     node.Description,
			HttpJSON:        node.Http,
			HttpsJSON:       node.Https,
			RestIsOn:        node.RestIsOn == 1,
			RestHTTPJSON:    node.RestHTTP,
			RestHTTPSJSON:   node.RestHTTPS,
			AccessAddrsJSON: node.AccessAddrs,
			AccessAddrs:     accessAddrs,
			StatusJSON:      node.Status,
			IsPrimary:       node.IsPrimary,
		})
	}

	return &pb.ListEnabledAPINodesResponse{ApiNodes: result}, nil
}

// FindEnabledAPINode 根据ID查找节点
func (this *APINodeService) FindEnabledAPINode(ctx context.Context, req *pb.FindEnabledAPINodeRequest) (*pb.FindEnabledAPINodeResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	node, err := models.SharedAPINodeDAO.FindEnabledAPINode(tx, req.ApiNodeId, nil)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledAPINodeResponse{ApiNode: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	result := &pb.APINode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn,
		NodeClusterId:   int64(node.ClusterId),
		UniqueId:        node.UniqueId,
		Secret:          node.Secret,
		Name:            node.Name,
		Description:     node.Description,
		HttpJSON:        node.Http,
		HttpsJSON:       node.Https,
		RestIsOn:        node.RestIsOn == 1,
		RestHTTPJSON:    node.RestHTTP,
		RestHTTPSJSON:   node.RestHTTPS,
		AccessAddrsJSON: node.AccessAddrs,
		AccessAddrs:     accessAddrs,
		IsPrimary:       node.IsPrimary,
		StatusJSON:      node.Status,
	}
	return &pb.FindEnabledAPINodeResponse{ApiNode: result}, nil
}

// FindCurrentAPINodeVersion 获取当前API节点的版本
func (this *APINodeService) FindCurrentAPINodeVersion(ctx context.Context, req *pb.FindCurrentAPINodeVersionRequest) (*pb.FindCurrentAPINodeVersionResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.FindCurrentAPINodeVersionResponse{
		Version: teaconst.Version,
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}, nil
}

// FindCurrentAPINode 获取当前API节点的信息
func (this *APINodeService) FindCurrentAPINode(ctx context.Context, req *pb.FindCurrentAPINodeRequest) (*pb.FindCurrentAPINodeResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var nodeId = teaconst.NodeId
	var tx *dbs.Tx
	node, err := models.SharedAPINodeDAO.FindEnabledAPINode(tx, nodeId, nil)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindCurrentAPINodeResponse{ApiNode: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	return &pb.FindCurrentAPINodeResponse{ApiNode: &pb.APINode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn,
		NodeClusterId:   0,
		UniqueId:        "",
		Secret:          "",
		Name:            "",
		Description:     "",
		HttpJSON:        nil,
		HttpsJSON:       nil,
		RestIsOn:        false,
		RestHTTPJSON:    nil,
		RestHTTPSJSON:   nil,
		AccessAddrsJSON: node.AccessAddrs,
		AccessAddrs:     accessAddrs,
		StatusJSON:      node.Status,
		IsPrimary:       node.IsPrimary,
		InstanceCode:    teaconst.InstanceCode,
	}}, nil
}

// CountAllEnabledAPINodesWithSSLCertId 计算使用某个SSL证书的API节点数量
func (this *APINodeService) CountAllEnabledAPINodesWithSSLCertId(ctx context.Context, req *pb.CountAllEnabledAPINodesWithSSLCertIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}
	if len(policyIds) == 0 {
		return this.SuccessCount(0)
	}

	count, err := models.SharedAPINodeDAO.CountAllEnabledAPINodesWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// DebugAPINode 修改调试模式状态
func (this *APINodeService) DebugAPINode(ctx context.Context, req *pb.DebugAPINodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	teaconst.Debug = req.Debug
	return this.Success()
}

// UploadAPINodeFile 上传新版API节点文件
func (this *APINodeService) UploadAPINodeFile(ctx context.Context, req *pb.UploadAPINodeFileRequest) (*pb.UploadAPINodeFileResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, errors.New("can not find executable file: " + err.Error())
	}

	var targetDir = filepath.Dir(exe)
	var targetFilename = teaconst.ProcessName // 这里不使用 filepath.Base() 是因为文件名可能变成修改后的临时文件名
	var targetCompressedFile = targetDir + "/." + targetFilename + ".gz"
	var targetFile = targetDir + "/." + targetFilename

	if req.IsFirstChunk {
		_ = os.Remove(targetCompressedFile)
		_ = os.Remove(targetFile)
	}

	if len(req.ChunkData) > 0 {
		err = func() error {
			var flags = os.O_CREATE | os.O_WRONLY
			if req.IsFirstChunk {
				flags |= os.O_TRUNC
			} else {
				flags |= os.O_APPEND
			}
			fp, err := os.OpenFile(targetCompressedFile, flags, 0666)
			if err != nil {
				return err
			}
			defer func() {
				_ = fp.Close()
			}()

			_, err = fp.Write(req.ChunkData)
			return err
		}()
		if err != nil {
			return nil, errors.New("write file failed: " + err.Error())
		}
	}

	if req.IsLastChunk {
		err = func() error {
			// 删除压缩文件
			defer func() {
				_ = os.Remove(targetCompressedFile)
			}()

			// 检查SUM
			fp, err := os.Open(targetCompressedFile)
			if err != nil {
				return err
			}
			defer func() {
				_ = fp.Close()
			}()

			var hash = md5.New()
			_, err = io.Copy(hash, fp)
			if err != nil {
				return err
			}

			var sum = fmt.Sprintf("%x", hash.Sum(nil))
			if sum != req.Sum {
				return errors.New("check sum failed: '" + sum + "' expected: '" + req.Sum + "'")
			}

			// 解压
			fp2, err := os.Open(targetCompressedFile)
			if err != nil {
				return err
			}

			defer func() {
				_ = fp2.Close()
			}()

			gzipReader, err := gzip.NewReader(fp2)
			if err != nil {
				return err
			}
			defer func() {
				_ = gzipReader.Close()
			}()
			targetWriter, err := os.OpenFile(targetFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
			if err != nil {
				return err
			}
			defer func() {
				_ = targetWriter.Close()
			}()
			_, err = io.Copy(targetWriter, gzipReader)
			if err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			return nil, errors.New("extract file failed: " + err.Error())
		}

		// 检查文件是否可执行
		var versionCmd = exec.Command(targetFile, "-V")
		var versionBuf = &bytes.Buffer{}
		versionCmd.Stdout = versionBuf
		err = versionCmd.Run()
		if err != nil {
			return nil, errors.New("test file failed: " + err.Error())
		}

		// 检查版本
		if stringutil.VersionCompare(versionCmd.String(), teaconst.Version) >= 0 {
			return &pb.UploadAPINodeFileResponse{}, nil
		}

		// 替换文件
		err = os.Remove(exe)
		if err != nil {
			return nil, errors.New("remove old file failed: " + err.Error())
		}
		err = os.Rename(targetFile, exe)
		if err != nil {
			return nil, errors.New("rename file failed: " + err.Error())
		}

		// 重启
		var cmd = exec.Command(exe, "restart")
		err = cmd.Start()
		if err != nil {
			return nil, errors.New("start new process failed: " + err.Error())
		}
	}

	return &pb.UploadAPINodeFileResponse{}, nil
}
