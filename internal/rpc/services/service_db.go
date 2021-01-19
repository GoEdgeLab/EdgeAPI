package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"strings"
)

// 数据库相关服务
type DBService struct {
	BaseService
}

// 获取所有表信息
func (this *DBService) FindAllDBTables(ctx context.Context, req *pb.FindAllDBTablesRequest) (*pb.FindAllDBTablesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
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

	return &pb.FindAllDBTablesResponse{DbTables: pbTables}, nil
}

// 删除表
func (this *DBService) DeleteDBTable(ctx context.Context, req *pb.DeleteDBTableRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}

	// 检查是否能够删除
	if !strings.HasPrefix(strings.ToLower(req.DbTable), "edgehttpaccesslogs_") {
		return nil, errors.New("forbidden to delete the table")
	}

	_, err = db.Exec("DROP TABLE `" + req.DbTable + "`")
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 清空表
func (this *DBService) TruncateDBTable(ctx context.Context, req *pb.TruncateDBTableRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("TRUNCATE TABLE `" + req.DbTable + "`")
	if err != nil {
		return nil, err
	}
	return this.Success()
}
