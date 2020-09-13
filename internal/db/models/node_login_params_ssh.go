package models

type NodeLoginSSHParams struct {
	GrantId int64  `json:"grantId"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}
