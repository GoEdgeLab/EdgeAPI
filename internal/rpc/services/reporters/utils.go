// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package reporters

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/dbs"
	"google.golang.org/grpc/peer"
	"net"
)

// 校验客户端IP
func validateClient(tx *dbs.Tx, nodeId int64, ctx context.Context) error {
	allowIPs, err := models.SharedReportNodeDAO.FindNodeAllowIPs(tx, nodeId)
	if err != nil {
		return err
	}
	if len(allowIPs) == 0 {
		return nil
	}

	p, ok := peer.FromContext(ctx)
	if ok {
		host, _, _ := net.SplitHostPort(p.Addr.String())
		if len(host) > 0 {
			for _, ip := range allowIPs {
				r, err := shared.ParseIPRange(ip)
				if err == nil && r != nil {
					if r.Contains(host) {
						return nil
					}
				}
			}
		}
	}
	return errors.New("client was not allowed")
}

