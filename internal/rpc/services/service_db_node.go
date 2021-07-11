package services

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"strings"
)

// DBNodeService 数据库节点相关服务
type DBNodeService struct {
	BaseService
}

// CreateDBNode 创建数据库节点
func (this *DBNodeService) CreateDBNode(ctx context.Context, req *pb.CreateDBNodeRequest) (*pb.CreateDBNodeResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedDBNodeDAO.CreateDBNode(tx, req.IsOn, req.Name, req.Description, req.Host, req.Port, req.Database, req.Username, req.Password, req.Charset)
	if err != nil {
		return nil, err
	}
	return &pb.CreateDBNodeResponse{DbNodeId: nodeId}, nil
}

// UpdateDBNode 修改数据库节点
func (this *DBNodeService) UpdateDBNode(ctx context.Context, req *pb.UpdateDBNodeRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedDBNodeDAO.UpdateNode(tx, req.DbNodeId, req.IsOn, req.Name, req.Description, req.Host, req.Port, req.Database, req.Username, req.Password, req.Charset)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteDBNode 删除节点
func (this *DBNodeService) DeleteDBNode(ctx context.Context, req *pb.DeleteDBNodeRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedDBNodeDAO.DisableDBNode(tx, req.DbNodeId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledDBNodes 计算可用的数据库节点数量
func (this *DBNodeService) CountAllEnabledDBNodes(ctx context.Context, req *pb.CountAllEnabledDBNodesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedDBNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledDBNodes 列出单页的数据库节点
func (this *DBNodeService) ListEnabledDBNodes(ctx context.Context, req *pb.ListEnabledDBNodesRequest) (*pb.ListEnabledDBNodesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedDBNodeDAO.ListEnabledNodes(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.DBNode{}
	for _, node := range nodes {
		status := &pb.DBNodeStatus{}

		// 是否能够连接
		if node.IsOn == 1 {
			db, err := dbs.NewInstanceFromConfig(node.DBConfig())
			if err != nil {
				status.Error = err.Error()
			} else {
				one, err := db.FindOne("SELECT SUM(DATA_LENGTH+INDEX_LENGTH) AS size FROM information_schema.`TABLES` WHERE TABLE_SCHEMA=?", db.Name())
				if err != nil {
					status.Error = err.Error()
					_ = db.Close()
				} else if one == nil {
					status.Error = "unable to read size from database server"
					_ = db.Close()
				} else {
					status.IsOk = true
					status.Size = one.GetInt64("size")
					_ = db.Close()
				}
			}
		}

		result = append(result, &pb.DBNode{
			Id:          int64(node.Id),
			Name:        node.Name,
			Description: node.Description,
			IsOn:        node.IsOn == 1,
			Host:        node.Host,
			Port:        types.Int32(node.Port),
			Database:    node.Database,
			Username:    node.Username,
			Password:    node.Password,
			Charset:     node.Charset,
			Status:      status,
		})
	}
	return &pb.ListEnabledDBNodesResponse{DbNodes: result}, nil
}

// FindEnabledDBNode 根据ID查找可用的数据库节点
func (this *DBNodeService) FindEnabledDBNode(ctx context.Context, req *pb.FindEnabledDBNodeRequest) (*pb.FindEnabledDBNodeResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedDBNodeDAO.FindEnabledDBNode(tx, req.DbNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindEnabledDBNodeResponse{DbNode: nil}, nil
	}
	return &pb.FindEnabledDBNodeResponse{DbNode: &pb.DBNode{
		Id:          int64(node.Id),
		Name:        node.Name,
		Description: node.Description,
		IsOn:        node.IsOn == 1,
		Host:        node.Host,
		Port:        types.Int32(node.Port),
		Database:    node.Database,
		Username:    node.Username,
		Password:    node.Password,
		Charset:     node.Charset,
	}}, nil
}

// FindAllDBNodeTables 获取所有表信息
func (this *DBNodeService) FindAllDBNodeTables(ctx context.Context, req *pb.FindAllDBNodeTablesRequest) (*pb.FindAllDBNodeTablesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	node, err := models.SharedDBNodeDAO.FindEnabledDBNode(tx, req.DbNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, dbs.ErrNotFound
	}
	db, err := dbs.NewInstanceFromConfig(node.DBConfig())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	ones, _, err := db.FindOnes("SELECT * FROM information_schema.`TABLES` WHERE TABLE_SCHEMA=?", db.Name())
	if err != nil {
		return nil, err
	}
	pbTables := []*pb.DBTable{}
	for _, one := range ones {
		lowerTableName := strings.ToLower(one.GetString("TABLE_NAME"))
		canDelete := false
		canClean := false
		if strings.HasPrefix(lowerTableName, "edgehttpaccesslogs_") {
			canDelete = true
			canClean = true
		} else if lists.ContainsString([]string{"edgemessages", "edgelogs", "edgenodelogs"}, lowerTableName) {
			canClean = true
		}

		pbTables = append(pbTables, &pb.DBTable{
			Name:        one.GetString("TABLE_NAME"),
			Schema:      one.GetString("TABLE_SCHEMA"),
			Type:        one.GetString("TABLE_TYPE"),
			Engine:      one.GetString("ENGINE"),
			Rows:        one.GetInt64("TABLE_ROWS"),
			DataLength:  one.GetInt64("DATA_LENGTH"),
			IndexLength: one.GetInt64("INDEX_LENGTH"),
			Comment:     one.GetString("TABLE_COMMENT"),
			Collation:   one.GetString("TABLE_COLLATION"),
			IsBaseTable: one.GetString("TABLE_TYPE") == "BASE TABLE",
			CanClean:    canClean,
			CanDelete:   canDelete,
		})
	}

	return &pb.FindAllDBNodeTablesResponse{DbNodeTables: pbTables}, nil
}

// DeleteDBNodeTable 删除表
func (this *DBNodeService) DeleteDBNodeTable(ctx context.Context, req *pb.DeleteDBNodeTableRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	node, err := models.SharedDBNodeDAO.FindEnabledDBNode(tx, req.DbNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, dbs.ErrNotFound
	}
	db, err := dbs.NewInstanceFromConfig(node.DBConfig())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	// 检查是否能够删除
	if !strings.HasPrefix(strings.ToLower(req.DbNodeTable), "edgehttpaccesslogs_") {
		return nil, errors.New("forbidden to delete the table")
	}

	_, err = db.Exec("DROP TABLE `" + req.DbNodeTable + "`")
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// TruncateDBNodeTable 清空表
func (this *DBNodeService) TruncateDBNodeTable(ctx context.Context, req *pb.TruncateDBNodeTableRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	node, err := models.SharedDBNodeDAO.FindEnabledDBNode(tx, req.DbNodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, dbs.ErrNotFound
	}
	db, err := dbs.NewInstanceFromConfig(node.DBConfig())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	_, err = db.Exec("TRUNCATE TABLE `" + req.DbNodeTable + "`")
	if err != nil {
		return nil, err
	}
	return this.Success()
}
