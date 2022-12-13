package clients

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

// TODO 需要定时对所有IP的PTR进行检查，剔除已经变更的IP

type ClientAgentIPDAO dbs.DAO

func NewClientAgentIPDAO() *ClientAgentIPDAO {
	return dbs.NewDAO(&ClientAgentIPDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeClientAgentIPs",
			Model:  new(ClientAgentIP),
			PkName: "id",
		},
	}).(*ClientAgentIPDAO)
}

var SharedClientAgentIPDAO *ClientAgentIPDAO

func init() {
	dbs.OnReady(func() {
		SharedClientAgentIPDAO = NewClientAgentIPDAO()
	})
}

// CreateIP 写入IP
func (this *ClientAgentIPDAO) CreateIP(tx *dbs.Tx, agentId int64, ip string, ptr string) error {
	// 检查数据有效性
	if agentId <= 0 || len(ip) == 0 {
		return nil
	}

	// 限制ptr长度
	if len(ptr) > 100 {
		ptr = ptr[:100]
	}

	// 检查是否存在
	exists, err := this.Query(tx).
		Attr("agentId", agentId).
		Attr("ip", ip).
		Exist()

	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	var op = NewClientAgentIPOperator()
	op.AgentId = agentId
	op.IP = ip
	op.Ptr = ptr
	err = this.Save(tx, op)
	if err != nil {
		// 忽略duplicate错误
		if models.CheckSQLDuplicateErr(err) {
			return nil
		}
		return err
	}

	// 更新Agent IP数量
	countIPs, err := this.CountAgentIPs(tx, agentId)
	if err != nil {
		return err
	}
	err = SharedClientAgentDAO.UpdateAgentCountIPs(tx, agentId, countIPs)
	if err != nil {
		return err
	}

	return nil
}

// ListIPsAfterId 列出某个ID之后的IP
func (this *ClientAgentIPDAO) ListIPsAfterId(tx *dbs.Tx, id int64, size int64) (result []*ClientAgentIP, err error) {
	if id < 0 {
		id = 0
	}

	_, err = this.Query(tx).
		Result("id", "ip", "agentId").
		Gt("id", id).
		AscPk().
		Limit(size). // 限制单次读取个数
		Slice(&result).
		FindAll()
	return
}

// CountAgentIPs 计算Agent IP数量
func (this *ClientAgentIPDAO) CountAgentIPs(tx *dbs.Tx, agentId int64) (int64, error) {
	return this.Query(tx).
		Attr("agentId", agentId).
		Count()
}
