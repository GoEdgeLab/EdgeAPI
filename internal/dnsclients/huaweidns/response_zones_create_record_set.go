// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package huaweidns

type ZonesCreateRecordSetResponse struct {
	Id      string   `json:"id"`
	Line    string   `json:"line"`
	Records []string `json:"records"`
}
