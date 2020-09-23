package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type TCPFirewallPolicyDAO dbs.DAO

func NewTCPFirewallPolicyDAO() *TCPFirewallPolicyDAO {
	return dbs.NewDAO(&TCPFirewallPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeTCPFirewallPolicies",
			Model:  new(TCPFirewallPolicy),
			PkName: "id",
		},
	}).(*TCPFirewallPolicyDAO)
}

var SharedTCPFirewallPolicyDAO = NewTCPFirewallPolicyDAO()
