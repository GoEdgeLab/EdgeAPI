package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/encrypt"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/maps"
	"google.golang.org/grpc/metadata"
	"time"
)

type AdminService struct {
	debug bool
}

func (this *AdminService) Login(ctx context.Context, req *pb.AdminLoginRequest) (*pb.AdminLoginResponse, error) {
	_, err := this.validateRequest(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.Username) == 0 || len(req.Password) == 0 {
		return &pb.AdminLoginResponse{
			AdminId: 0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	adminId, err := models.SharedAdminDAO.CheckAdminPassword(req.Username, req.Password)
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	if adminId <= 0 {
		return &pb.AdminLoginResponse{
			AdminId: 0,
			IsOk:    false,
			Message: "请输入正确的用户名密码",
		}, nil
	}

	return &pb.AdminLoginResponse{
		AdminId: int64(adminId),
		IsOk:    true,
	}, nil
}

func (this *AdminService) CreateLog(ctx context.Context, req *pb.AdminCreateLogRequest) (*pb.AdminCreateLogResponse, error) {
	adminId, err := this.validateAdminRequest(ctx)
	if err != nil {
		return nil, err
	}
	err = models.SharedLogDAO.CreateAdminLog(adminId, req.Level, req.Description, req.Action, req.Ip)
	return &pb.AdminCreateLogResponse{
		IsOk: err != nil,
	}, err
}

func (this *AdminService) CheckAdminExists(ctx context.Context, req *pb.AdminCheckAdminExistsRequest) (*pb.AdminCheckAdminExistsResponse, error) {
	_, err := this.validateRequest(ctx)
	if err != nil {
		return nil, err
	}

	if req.AdminId <= 0 {
		return &pb.AdminCheckAdminExistsResponse{
			IsOk: false,
		}, nil
	}

	ok, err := models.SharedAdminDAO.ExistEnabledAdmin(int(req.AdminId))
	if err != nil {
		return nil, err
	}

	return &pb.AdminCheckAdminExistsResponse{
		IsOk: ok,
	}, nil
}

func (this *AdminService) FindAdminFullname(ctx context.Context, req *pb.AdminFindAdminNameRequest) (*pb.AdminFindAdminNameResponse, error) {
	_, err := this.validateRequest(ctx)
	if err != nil {
		return nil, err
	}

	fullname, err := models.SharedAdminDAO.FindAdminFullname(int(req.AdminId))
	if err != nil {
		utils.PrintError(err)
		return nil, err
	}

	return &pb.AdminFindAdminNameResponse{
		Fullname: fullname,
	}, nil
}

func (this *AdminService) FindAllEnabledClusters(ctx context.Context, req *pb.AdminFindAllEnabledClustersRequest) (*pb.AdminFindAllEnabledClustersResponse, error) {
	_ = req

	_, err := this.validateAdminRequest(ctx)
	if err != nil {
		return nil, err
	}

	clusters, err := models.SharedNodeClusterDAO.FindAllEnableClusters()
	if err != nil {
		return nil, err
	}

	result := []*pb.Cluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.Cluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
		})
	}

	return &pb.AdminFindAllEnabledClustersResponse{
		Clusters: result,
	}, nil
}

func (this *AdminService) CreateNode(ctx context.Context, req *pb.AdminCreateNodeRequest) (*pb.AdminCreateNodeResponse, error) {
	_, err := this.validateAdminRequest(ctx)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedNodeDAO.CreateNode(req.Name, int(req.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.AdminCreateNodeResponse{
		NodeId: int64(nodeId),
	}, nil
}

func (this *AdminService) CountAllEnabledNodes(ctx context.Context, req *pb.AdminCountAllEnabledNodesRequest) (*pb.AdminCountAllEnabledNodesResponse, error) {
	_, err := this.validateAdminRequest(ctx)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedNodeDAO.CountAllEnabledNodes()
	if err != nil {
		return nil, err
	}

	return &pb.AdminCountAllEnabledNodesResponse{Count: count}, nil
}

func (this *AdminService) ListEnabledNodes(ctx context.Context, req *pb.AdminListEnabledNodesRequest) (*pb.AdminListEnabledNodesResponse, error) {
	_, err := this.validateAdminRequest(ctx)
	if err != nil {
		return nil, err
	}
	nodes, err := models.SharedNodeDAO.ListEnabledNodes(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.Node{}
	for _, node := range nodes {
		// 集群信息
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(node.ClusterId))
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.Node{
			Id:   int64(node.Id),
			Name: node.Name,
			Cluster: &pb.Cluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
		})
	}

	return &pb.AdminListEnabledNodesResponse{
		Nodes: result,
	}, nil
}

func (this *AdminService) validateRequest(ctx context.Context) (adminId int, err error) {
	var md metadata.MD
	var ok bool
	if this.debug {
		md, ok = metadata.FromOutgoingContext(ctx)
	} else {
		md, ok = metadata.FromIncomingContext(ctx)
	}
	if !ok {
		return 0, errors.New("context: need 'nodeId'")
	}
	nodeIds := md.Get("nodeid")
	if len(nodeIds) == 0 || len(nodeIds[0]) == 0 {
		return 0, errors.New("context: need 'nodeId'")
	}
	nodeId := nodeIds[0]

	// 获取Node信息
	apiToken, err := models.SharedApiTokenDAO.FindEnabledTokenWithNode(nodeId)
	if err != nil {
		utils.PrintError(err)
		return 0, err
	}
	if apiToken == nil {
		return 0, errors.New("can not find token from node id: " + err.Error())
	}

	tokens := md.Get("token")
	if len(tokens) == 0 || len(tokens[0]) == 0 {
		return 0, errors.New("context: need 'token'")
	}
	token := tokens[0]

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}

	method, err := encrypt.NewMethodInstance(teaconst.EncryptMethod, apiToken.Secret, nodeId)
	if err != nil {
		utils.PrintError(err)
		return 0, err
	}
	data, err = method.Decrypt(data)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, errors.New("invalid token")
	}

	m := maps.Map{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return 0, errors.New("decode token error: " + err.Error())
	}

	timestamp := m.GetInt64("timestamp")
	if time.Now().Unix()-timestamp > 600 {
		// 请求超过10分钟认为超时
		return 0, errors.New("authenticate timeout")
	}

	adminId = m.GetInt("adminId")
	return
}

func (this *AdminService) validateAdminRequest(ctx context.Context) (adminId int, err error) {
	adminId, err = this.validateRequest(ctx)
	if err != nil {
		return 0, err
	}
	if adminId <= 0 {
		return 0, errors.New("invalid admin id")
	}
	return
}
