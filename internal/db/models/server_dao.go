package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
)

const (
	ServerStateEnabled  = 1 // 已启用
	ServerStateDisabled = 0 // 已禁用
)

type ServerDAO dbs.DAO

func NewServerDAO() *ServerDAO {
	return dbs.NewDAO(&ServerDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServers",
			Model:  new(Server),
			PkName: "id",
		},
	}).(*ServerDAO)
}

var SharedServerDAO = NewServerDAO()

// 启用条目
func (this *ServerDAO) EnableServer(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ServerStateEnabled).
		Update()
}

// 禁用条目
func (this *ServerDAO) DisableServer(id int64) (err error) {
	_, err = this.Query().
		Pk(id).
		Set("state", ServerStateDisabled).
		Update()
	return
}

// 查找启用中的条目
func (this *ServerDAO) FindEnabledServer(id int64) (*Server, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ServerStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Server), err
}

// 创建服务
func (this *ServerDAO) CreateServer(adminId int64, userId int64, clusterId int64, configJSON string, includeNodesJSON string, excludeNodesJSON string) (serverId int64, err error) {
	uniqueId, err := this.genUniqueId()
	if err != nil {
		return 0, err
	}

	op := NewServerOperator()
	op.UniqueId = uniqueId
	op.UserId = userId
	op.AdminId = adminId
	op.ClusterId = clusterId
	if len(configJSON) > 0 {
		op.Config = configJSON
	}
	if len(includeNodesJSON) > 0 {
		op.IncludeNodes = includeNodesJSON
	}
	if len(excludeNodesJSON) > 0 {
		op.ExcludeNodes = excludeNodesJSON
	}
	op.GroupIds = "[]"
	op.Version = 1
	op.IsOn = 1
	op.State = ServerStateEnabled
	_, err = this.Save(op)
	return types.Int64(op.Id), err
}

// 修改服务
func (this *ServerDAO) UpdateServer(serverId int64, clusterId int64, configJSON string, includeNodesJSON string, excludeNodesJSON string) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	op := NewServerOperator()
	op.Id = serverId
	op.ClusterId = clusterId
	if len(configJSON) > 0 {
		op.Config = configJSON
	}
	if len(includeNodesJSON) > 0 {
		op.IncludeNodes = includeNodesJSON
	}
	if len(excludeNodesJSON) > 0 {
		op.ExcludeNodes = excludeNodesJSON
	}
	op.Version = dbs.SQL("version=version+1")
	_, err := this.Save(op)
	return err
}

// 计算所有可用服务数量
func (this *ServerDAO) CountAllEnabledServers() (int64, error) {
	return this.Query().
		State(ServerStateEnabled).
		Count()
}

// 列出单页的服务
func (this *ServerDAO) ListEnabledServers(offset int64, size int64) (result []*Server, err error) {
	_, err = this.Query().
		State(ServerStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 获取节点中的所有服务
func (this *ServerDAO) FindAllEnabledServersWithNode(nodeId int64) (result []*Server, err error) {
	// 节点所在集群
	clusterId, err := SharedNodeDAO.FindNodeClusterId(nodeId)
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, nil
	}

	_, err = this.Query().
		Attr("clusterId", clusterId).
		State(ServerStateEnabled).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 生成唯一ID
func (this *ServerDAO) genUniqueId() (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query().
			Attr("uniqueId", uniqueId).
			Exist()
		if err != nil {
			return "", err
		}
		if ok {
			continue
		}
		return uniqueId, nil
	}
}
