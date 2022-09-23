// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import "github.com/iwind/TeaGo/dbs"

var nsAccessLogDAOMapping = map[int64]any{} // dbNodeId => DAO

func initAccessLogDAO(db *dbs.DB, node *DBNode) {
}
