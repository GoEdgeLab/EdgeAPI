// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
	"net"
	"sync"
	"time"
)

// NodeLoginService 节点登录相关
type NodeLoginService struct {
	BaseService
}

// FindNodeLoginSuggestPorts 读取建议的端口
func (this *NodeLoginService) FindNodeLoginSuggestPorts(ctx context.Context, req *pb.FindNodeLoginSuggestPortsRequest) (*pb.FindNodeLoginSuggestPortsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	ports, err := models.SharedNodeLoginDAO.FindFrequentPorts(tx)
	if err != nil {
		return nil, err
	}

	var availablePorts = []int32{}

	// 测试端口连通性
	if len(ports) > 0 && len(req.Host) > 0 {
		var host = configutils.QuoteIP(req.Host)

		wg := sync.WaitGroup{}
		wg.Add(len(ports))

		var locker sync.Mutex

		for _, port := range ports {
			go func(port int32) {
				defer wg.Done()

				conn, err := net.DialTimeout("tcp", host+":"+types.String(port), 2*time.Second)
				if err != nil {
					return
				}
				_ = conn.Close()

				locker.Lock()
				availablePorts = append(availablePorts, port)
				locker.Unlock()
			}(port)
		}
		wg.Wait()

	}

	return &pb.FindNodeLoginSuggestPortsResponse{
		Ports:          ports,
		AvailablePorts: availablePorts,
	}, nil
}
