package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

type NodeGrantService struct {
	BaseService
}

// CreateNodeGrant 创建认证
func (this *NodeGrantService) CreateNodeGrant(ctx context.Context, req *pb.CreateNodeGrantRequest) (*pb.CreateNodeGrantResponse, error) {
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	grantId, err := models.SharedNodeGrantDAO.CreateGrant(tx, adminId, req.Name, req.Method, req.Username, req.Password, req.PrivateKey, req.Passphrase, req.Description, req.NodeId, req.Su)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeGrantResponse{
		NodeGrantId: grantId,
	}, err
}

// UpdateNodeGrant 修改认证
func (this *NodeGrantService) UpdateNodeGrant(ctx context.Context, req *pb.UpdateNodeGrantRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.NodeGrantId <= 0 {
		return nil, errors.New("wrong grantId")
	}

	var tx = this.NullTx()

	err = models.SharedNodeGrantDAO.UpdateGrant(tx, req.NodeGrantId, req.Name, req.Method, req.Username, req.Password, req.PrivateKey, req.Passphrase, req.Description, req.NodeId, req.Su)
	return this.Success()
}

// DisableNodeGrant 禁用认证
func (this *NodeGrantService) DisableNodeGrant(ctx context.Context, req *pb.DisableNodeGrantRequest) (*pb.DisableNodeGrantResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedNodeGrantDAO.DisableNodeGrant(tx, req.NodeGrantId)
	return &pb.DisableNodeGrantResponse{}, err
}

// CountAllEnabledNodeGrants 计算认证的数量
func (this *NodeGrantService) CountAllEnabledNodeGrants(ctx context.Context, req *pb.CountAllEnabledNodeGrantsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedNodeGrantDAO.CountAllEnabledGrants(tx, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledNodeGrants 列出单页认证
func (this *NodeGrantService) ListEnabledNodeGrants(ctx context.Context, req *pb.ListEnabledNodeGrantsRequest) (*pb.ListEnabledNodeGrantsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	grants, err := models.SharedNodeGrantDAO.ListEnabledGrants(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeGrant{}
	for _, grant := range grants {
		result = append(result, &pb.NodeGrant{
			Id:          int64(grant.Id),
			Name:        grant.Name,
			Method:      grant.Method,
			Username:    grant.Username,
			Password:    grant.Password,
			Su:          grant.Su == 1,
			PrivateKey:  grant.PrivateKey,
			Description: grant.Description,
			NodeId:      int64(grant.NodeId),
		})
	}

	return &pb.ListEnabledNodeGrantsResponse{NodeGrants: result}, nil
}

// FindAllEnabledNodeGrants 列出所有认证信息
func (this *NodeGrantService) FindAllEnabledNodeGrants(ctx context.Context, req *pb.FindAllEnabledNodeGrantsRequest) (*pb.FindAllEnabledNodeGrantsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	grants, err := models.SharedNodeGrantDAO.FindAllEnabledGrants(this.NullTx())
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeGrant{}
	for _, grant := range grants {
		result = append(result, &pb.NodeGrant{
			Id:          int64(grant.Id),
			Name:        grant.Name,
			Method:      grant.Method,
			Username:    grant.Username,
			Password:    grant.Password,
			Su:          grant.Su == 1,
			PrivateKey:  grant.PrivateKey,
			Description: grant.Description,
			NodeId:      int64(grant.NodeId),
		})
	}

	return &pb.FindAllEnabledNodeGrantsResponse{NodeGrants: result}, nil
}

// FindEnabledNodeGrant 获取单个认证信息
func (this *NodeGrantService) FindEnabledNodeGrant(ctx context.Context, req *pb.FindEnabledNodeGrantRequest) (*pb.FindEnabledNodeGrantResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(this.NullTx(), req.NodeGrantId)
	if err != nil {
		return nil, err
	}
	if grant == nil {
		return &pb.FindEnabledNodeGrantResponse{}, nil
	}
	return &pb.FindEnabledNodeGrantResponse{NodeGrant: &pb.NodeGrant{
		Id:          int64(grant.Id),
		Name:        grant.Name,
		Method:      grant.Method,
		Username:    grant.Username,
		Password:    grant.Password,
		Su:          grant.Su == 1,
		PrivateKey:  grant.PrivateKey,
		Passphrase:  grant.Passphrase,
		Description: grant.Description,
		NodeId:      int64(grant.NodeId),
	}}, nil
}

// TestNodeGrant 测试连接
func (this *NodeGrantService) TestNodeGrant(ctx context.Context, req *pb.TestNodeGrantRequest) (*pb.TestNodeGrantResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var hostKeyCallback ssh.HostKeyCallback = nil

	resp := &pb.TestNodeGrantResponse{
		IsOk:  false,
		Error: "",
	}

	var tx = this.NullTx()
	grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(tx, req.NodeGrantId)
	if err != nil {
		return nil, err
	}
	if grant == nil {
		resp.Error = "can not find grant with id '" + numberutils.FormatInt64(req.NodeGrantId) + "'"
		return resp, nil
	}

	// 检查参数
	if len(req.Host) == 0 {
		resp.Error = "'host' should not be empty"
		return resp, nil
	}
	if req.Port <= 0 {
		resp.Error = "'port' should be greater than 0"
		return resp, nil
	}

	if len(grant.Password) == 0 && len(grant.PrivateKey) == 0 {
		resp.Error = "require user 'password' or 'privateKey'"
		return resp, nil
	}

	// 不使用known_hosts
	if hostKeyCallback == nil {
		hostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}
	}

	// 认证
	methods := []ssh.AuthMethod{}
	if grant.Method == "user" {
		{
			authMethod := ssh.Password(grant.Password)
			methods = append(methods, authMethod)
		}

		{
			authMethod := ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				if len(questions) == 0 {
					return []string{}, nil
				}
				return []string{grant.Password}, nil
			})
			methods = append(methods, authMethod)
		}
	} else if grant.Method == "privateKey" {
		var signer ssh.Signer
		if len(grant.Passphrase) != 0 {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(grant.PrivateKey), []byte(grant.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(grant.PrivateKey))
		}
		if err != nil {
			resp.Error = "parse private key: " + err.Error()
			return resp, nil
		}
		authMethod := ssh.PublicKeys(signer)
		methods = append(methods, authMethod)
	} else {
		return nil, errors.New("invalid method '" + grant.Method + "'")
	}

	// SSH客户端
	if len(grant.Username) == 0 {
		grant.Username = "root"
	}
	config := &ssh.ClientConfig{
		User:            grant.Username,
		Auth:            methods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         5 * time.Second, // TODO 后期可以设置这个超时时间
	}

	sshClient, err := ssh.Dial("tcp", configutils.QuoteIP(req.Host)+":"+fmt.Sprintf("%d", req.Port), config)
	if err != nil {
		resp.Error = "connect failed: " + err.Error()
		return resp, nil
	}
	defer func() {
		_ = sshClient.Close()
	}()

	resp.IsOk = true
	return resp, nil
}

// FindSuggestNodeGrants 查找集群推荐的认证
func (this *NodeGrantService) FindSuggestNodeGrants(ctx context.Context, req *pb.FindSuggestNodeGrantsRequest) (*pb.FindSuggestNodeGrantsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var pbGrants = []*pb.NodeGrant{}
	var tx = this.NullTx()
	grantIds, err := models.SharedNodeLoginDAO.FindFrequentGrantIds(tx, req.NodeClusterId, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	for _, grantId := range grantIds {
		grant, err := models.SharedNodeGrantDAO.FindEnabledNodeGrant(tx, grantId)
		if err != nil {
			return nil, err
		}
		if grant != nil {
			pbGrants = append(pbGrants, &pb.NodeGrant{
				Id:          int64(grant.Id),
				Name:        grant.Name,
				Method:      grant.Method,
				Username:    grant.Username,
				Su:          grant.Su == 1,
				Description: grant.Description,
			})
		}
	}
	return &pb.FindSuggestNodeGrantsResponse{NodeGrants: pbGrants}, nil
}
