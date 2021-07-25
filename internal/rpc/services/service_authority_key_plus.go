// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
// +build plus

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/authority"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	plusutils "github.com/TeaOSLab/EdgePlus/pkg/utils"
)

// AuthorityKeyService 版本认证
type AuthorityKeyService struct {
	BaseService
}

// UpdateAuthorityKey 设置Key
func (this *AuthorityKeyService) UpdateAuthorityKey(ctx context.Context, req *pb.UpdateAuthorityKeyRequest) (*pb.RPCSuccess, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAuthority)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = authority.SharedAuthorityKeyDAO.UpdateKey(tx, req.Value, req.DayFrom, req.DayTo, req.Hostname, req.MacAddresses, req.Company)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ReadAuthorityKey 读取Key
func (this *AuthorityKeyService) ReadAuthorityKey(ctx context.Context, req *pb.ReadAuthorityKeyRequest) (*pb.ReadAuthorityKeyResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeMonitor, rpcutils.UserTypeProvider, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	key, err := authority.SharedAuthorityKeyDAO.ReadKey(tx)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &pb.ReadAuthorityKeyResponse{AuthorityKey: nil}, nil
	}

	if len(key.Value) == 0 {
		return &pb.ReadAuthorityKeyResponse{AuthorityKey: nil}, nil
	}

	m, err := plusutils.Decode([]byte(key.Value))
	if err != nil {
		return nil, err
	}

	macAddresses := []string{}
	if len(key.MacAddresses) > 0 {
		err = json.Unmarshal([]byte(key.MacAddresses), &macAddresses)
		if err != nil {
			return nil, err
		}
	}

	return &pb.ReadAuthorityKeyResponse{AuthorityKey: &pb.AuthorityKey{
		Value:        key.Value,
		DayFrom:      m.GetString("dayFrom"),
		DayTo:        m.GetString("dayTo"),
		Hostname:     key.Hostname,
		MacAddresses: macAddresses,
		Company:      key.Company,
		UpdatedAt:    int64(key.UpdatedAt),
	}}, nil
}

// ResetAuthorityKey 重置Key
func (this *AuthorityKeyService) ResetAuthorityKey(ctx context.Context, req *pb.ResetAuthorityKeyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	err = authority.SharedAuthorityKeyDAO.ResetKey(nil)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
