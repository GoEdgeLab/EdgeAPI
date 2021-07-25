// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NSKeyService NS密钥相关服务
type NSKeyService struct {
	services.BaseService
}

// CreateNSKey 创建密钥
func (this *NSKeyService) CreateNSKey(ctx context.Context, req *pb.CreateNSKeyRequest) (*pb.CreateNSKeyResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	keyId, err := nameservers.SharedNSKeyDAO.CreateKey(tx, req.NsDomainId, req.NsZoneId, req.Name, req.Algo, req.Secret, req.SecretType)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSKeyResponse{NsKeyId: keyId}, nil
}

// UpdateNSKey 修改密钥
func (this *NSKeyService) UpdateNSKey(ctx context.Context, req *pb.UpdateNSKeyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = nameservers.SharedNSKeyDAO.UpdateKey(tx, req.NsKeyId, req.Name, req.Algo, req.Secret, req.SecretType, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSKey 删除密钥
func (this *NSKeyService) DeleteNSKey(ctx context.Context, req *pb.DeleteNSKeyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSKeyDAO.DisableNSKey(tx, req.NsKeyId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledNSKey 查找单个密钥
func (this *NSKeyService) FindEnabledNSKey(ctx context.Context, req *pb.FindEnabledNSKeyRequest) (*pb.FindEnabledNSKeyResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	key, err := nameservers.SharedNSKeyDAO.FindEnabledNSKey(tx, req.NsKeyId)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &pb.FindEnabledNSKeyResponse{NsKey: nil}, nil
	}
	return &pb.FindEnabledNSKeyResponse{
		NsKey: &pb.NSKey{
			Id:         int64(key.Id),
			IsOn:       key.IsOn == 1,
			Name:       key.Name,
			Algo:       key.Algo,
			Secret:     key.Secret,
			SecretType: key.SecretType,
		},
	}, nil
}

// CountAllEnabledNSKeys 计算密钥数量
func (this *NSKeyService) CountAllEnabledNSKeys(ctx context.Context, req *pb.CountAllEnabledNSKeysRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := nameservers.SharedNSKeyDAO.CountEnabledKeys(tx, req.NsDomainId, req.NsZoneId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledNSKeys 列出单页密钥
func (this *NSKeyService) ListEnabledNSKeys(ctx context.Context, req *pb.ListEnabledNSKeysRequest) (*pb.ListEnabledNSKeysResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	keys, err := nameservers.SharedNSKeyDAO.ListEnabledKeys(tx, req.NsDomainId, req.NsZoneId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbKeys = []*pb.NSKey{}
	for _, key := range keys {
		pbKeys = append(pbKeys, &pb.NSKey{
			Id:         int64(key.Id),
			IsOn:       key.IsOn == 1,
			Name:       key.Name,
			Algo:       key.Algo,
			Secret:     key.Secret,
			SecretType: key.SecretType,
		})
	}
	return &pb.ListEnabledNSKeysResponse{NsKeys: pbKeys}, nil
}
