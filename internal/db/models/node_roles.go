package models

type NodeRole = string

const (
	NodeRoleAdmin    NodeRole = "admin"
	NodeRoleUser     NodeRole = "user"
	NodeRoleProvider NodeRole = "provider"
	NodeRoleAPI      NodeRole = "api"
	NodeRoleDatabase NodeRole = "database"
	NodeRoleLog      NodeRole = "log"
	NodeRoleDNS      NodeRole = "dns"
	NodeRoleMonitor  NodeRole = "monitor"
	NodeRoleNode     NodeRole = "node"
	NodeRoleCluster  NodeRole = "cluster"
)
