package clients

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type ClientAgentDAO dbs.DAO

func NewClientAgentDAO() *ClientAgentDAO {
	return dbs.NewDAO(&ClientAgentDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeClientAgents",
			Model:  new(ClientAgent),
			PkName: "id",
		},
	}).(*ClientAgentDAO)
}

var SharedClientAgentDAO *ClientAgentDAO

func init() {
	dbs.OnReady(func() {
		SharedClientAgentDAO = NewClientAgentDAO()
	})
}

// FindClientAgentName 根据主键查找名称
func (this *ClientAgentDAO) FindClientAgentName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAgent 查找Agent
func (this *ClientAgentDAO) FindAgent(tx *dbs.Tx, agentId int64) (*ClientAgent, error) {
	if agentId <= 0 {
		return nil, nil
	}

	one, err := this.Query(tx).
		Pk(agentId).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*ClientAgent), nil
}

// FindAgentIdWithCode 根据代号查找ID
func (this *ClientAgentDAO) FindAgentIdWithCode(tx *dbs.Tx, code string) (int64, error) {
	return this.Query(tx).
		ResultPk().
		Attr("code", code).
		FindInt64Col(0)
}

// FindAgentNameWithCode 根据代号查找Agent名称
func (this *ClientAgentDAO) FindAgentNameWithCode(tx *dbs.Tx, code string) (string, error) {
	return this.Query(tx).
		Result("name").
		Attr("code", code).
		FindStringCol("")
}

// UpdateAgentCountIPs 修改Agent拥有的IP数量
func (this *ClientAgentDAO) UpdateAgentCountIPs(tx *dbs.Tx, agentId int64, countIPs int64) error {
	return this.Query(tx).
		Pk(agentId).
		Set("countIPs", countIPs).
		UpdateQuickly()
}

// FindAllAgents 查找所有Agents
func (this *ClientAgentDAO) FindAllAgents(tx *dbs.Tx) (result []*ClientAgent, err error) {
	_, err = this.Query(tx).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllNSAgents 查找所有DNS可以使用的Agents
func (this *ClientAgentDAO) FindAllNSAgents(tx *dbs.Tx) (result []*ClientAgent, err error) {
	// 注意：允许NS使用所有的Agent，不管有没有IP数据
	_, err = this.Query(tx).
		Result("id", "name", "code").
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}
