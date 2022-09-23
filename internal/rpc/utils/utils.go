package rpcutils

import (
	"errors"
)

type UserType = string

const (
	UserTypeNone      = "none"
	UserTypeAdmin     = "admin"
	UserTypeUser      = "user"
	UserTypeProvider  = "provider"
	UserTypeNode      = "node"
	UserTypeCluster   = "cluster"
	UserTypeMonitor   = "monitor"
	UserTypeStat      = "stat"
	UserTypeDNS       = "dns"
	UserTypeLog       = "log"
	UserTypeAPI       = "api"
	UserTypeAuthority = "authority"
	UserTypeReport    = "report"
)

// Wrap 包装错误
func Wrap(description string, err error) error {
	if err == nil {
		return errors.New(description)
	}
	return errors.New(description + ": " + err.Error())
}
