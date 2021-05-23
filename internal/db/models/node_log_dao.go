package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	stringutil "github.com/iwind/TeaGo/utils/string"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strconv"
	"strings"
	"time"
)

type NodeLogDAO dbs.DAO

func NewNodeLogDAO() *NodeLogDAO {
	return dbs.NewDAO(&NodeLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeLogs",
			Model:  new(NodeLog),
			PkName: "id",
		},
	}).(*NodeLogDAO)
}

var SharedNodeLogDAO *NodeLogDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeLogDAO = NewNodeLogDAO()
	})
}

// CreateLog 创建日志
func (this *NodeLogDAO) CreateLog(tx *dbs.Tx, nodeRole NodeRole, nodeId int64, serverId int64, level string, tag string, description string, createdAt int64) error {
	hash := stringutil.Md5(nodeRole + "@" + strconv.FormatInt(nodeId, 10) + "@" + strconv.FormatInt(serverId, 10) + "@" + level + "@" + tag + "@" + description)

	// 检查是否在重复最后一条，避免重复创建
	lastLog, err := this.Query(tx).
		Result("id", "hash", "createdAt").
		DescPk().
		Find()
	if err != nil {
		return err
	}
	if lastLog != nil {
		nodeLog := lastLog.(*NodeLog)
		if nodeLog.Hash == hash && time.Now().Unix()-int64(nodeLog.CreatedAt) < 1800 {
			err = this.Query(tx).
				Pk(nodeLog.Id).
				Set("count", dbs.SQL("count+1")).
				UpdateQuickly()
			return err
		}
	}

	op := NewNodeLogOperator()
	op.Role = nodeRole
	op.NodeId = nodeId
	op.ServerId = serverId
	op.Level = level
	op.Tag = tag
	op.Description = description
	op.CreatedAt = createdAt
	op.Day = timeutil.FormatTime("Ymd", createdAt)
	op.Hash = hash
	err = this.Save(tx, op)
	return err
}

// DeleteExpiredLogs 清除超出一定日期的日志
func (this *NodeLogDAO) DeleteExpiredLogs(tx *dbs.Tx, days int) error {
	if days <= 0 {
		return errors.New("invalid days '" + strconv.Itoa(days) + "'")
	}
	date := time.Now().AddDate(0, 0, -days)
	expireDay := timeutil.Format("Ymd", date)
	_, err := this.Query(tx).
		Where("day<=:day").
		Param("day", expireDay).
		Delete()
	return err
}

// CountNodeLogs 计算节点日志数量
func (this *NodeLogDAO) CountNodeLogs(tx *dbs.Tx, role string, nodeId int64, serverId int64, dayFrom string, dayTo string, keyword string, level string) (int64, error) {
	query := this.Query(tx).
		Attr("role", role)
	if nodeId > 0 {
		query.Attr("nodeId", nodeId)
	} else {
		switch role {
		case NodeRoleNode:
			query.Where("nodeId IN (SELECT id FROM " + SharedNodeDAO.Table + " WHERE state=1)")
		}
	}
	if serverId > 0 {
		query.Attr("serverId", serverId)
	}
	if len(dayFrom) > 0 {
		dayFrom = strings.ReplaceAll(dayFrom, "-", "")
		query.Gte("day", dayFrom)
	}
	if len(dayTo) > 0 {
		dayTo = strings.ReplaceAll(dayTo, "-", "")
		query.Lte("day", dayTo)
	}
	if len(keyword) > 0 {
		query.Where("(tag LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	if len(level) > 0 {
		query.Attr("level", level)
	}

	return query.Count()
}

// ListNodeLogs 列出单页日志
func (this *NodeLogDAO) ListNodeLogs(tx *dbs.Tx,
	role string,
	nodeId int64,
	serverId int64,
	allServers bool,
	dayFrom string,
	dayTo string,
	keyword string,
	level string,
	fixedState configutils.BoolState,
	offset int64,
	size int64) (result []*NodeLog, err error) {
	query := this.Query(tx).
		Attr("role", role)
	if nodeId > 0 {
		query.Attr("nodeId", nodeId)
	} else {
		switch role {
		case NodeRoleNode:
			query.Where("nodeId IN (SELECT id FROM " + SharedNodeDAO.Table + " WHERE state=1)")
		}
	}
	if serverId > 0 {
		query.Attr("serverId", serverId)
	} else if allServers {
		query.Where("serverId>0")
	}
	if fixedState == configutils.BoolStateYes {
		query.Attr("isFixed", 1)
	} else if fixedState == configutils.BoolStateNo {
		query.Attr("isFixed", 0)
	}
	if len(dayFrom) > 0 {
		dayFrom = strings.ReplaceAll(dayFrom, "-", "")
		query.Gte("day", dayFrom)
	}
	if len(dayTo) > 0 {
		dayTo = strings.ReplaceAll(dayTo, "-", "")
		query.Lte("day", dayTo)
	}
	if len(keyword) > 0 {
		query.Where("(tag LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	if len(level) > 0 {
		query.Attr("level", level)
	}
	_, err = query.
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// UpdateNodeLogFixed 设置节点日志为已修复
func (this *NodeLogDAO) UpdateNodeLogFixed(tx *dbs.Tx, logId int64) error {
	if logId <= 0 {
		return errors.New("invalid logId")
	}

	// 我们把相同内容的日志都置为已修复
	hash, err := this.Query(tx).
		Pk(logId).
		Result("hash").
		FindStringCol("")
	if err != nil {
		return err
	}
	if len(hash) == 0 {
		return nil
	}

	err = this.Query(tx).
		Attr("hash", hash).
		Attr("isFixed", false).
		Set("isFixed", true).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return nil
}
