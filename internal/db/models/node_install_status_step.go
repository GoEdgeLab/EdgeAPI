package models

type NodeInstallStatusStep struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Percent     int    `json:"percent"`
}
