// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package accesslogs

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"sync"
	"time"
)

var SharedStorageManager = NewStorageManager()

type StorageManager struct {
	storageMap map[int64]StorageInterface // policyId => Storage

	locker sync.Mutex
}

func NewStorageManager() *StorageManager {
	return &StorageManager{
		storageMap: map[int64]StorageInterface{},
	}
}

func (this *StorageManager) Start() {
	var ticker = time.NewTicker(1 * time.Minute)
	if Tea.IsTesting() {
		ticker = time.NewTicker(5 * time.Second)
	}

	// 启动时执行一次
	var err = this.Loop()
	if err != nil {
		remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "update error: "+err.Error())
	}

	// 循环执行
	for range ticker.C {
		err := this.Loop()
		if err != nil {
			remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "update error: "+err.Error())
		}
	}
}

// Loop 更新
func (this *StorageManager) Loop() error {
	policies, err := models.SharedHTTPAccessLogPolicyDAO.FindAllEnabledAndOnPolicies(nil)
	if err != nil {
		return err
	}
	var policyIds = []int64{}
	for _, policy := range policies {
		if policy.IsOn == 1 {
			policyIds = append(policyIds, int64(policy.Id))
		}
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	// 关闭不用的
	for policyId, storage := range this.storageMap {
		if !lists.ContainsInt64(policyIds, policyId) {
			err := storage.Close()
			if err != nil {
				remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "close '"+types.String(policyId)+"' failed: "+err.Error())
			}
			delete(this.storageMap, policyId)
			remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "remove '"+types.String(policyId)+"'")
		}
	}

	for _, policy := range policies {
		var policyId = int64(policy.Id)
		storage, ok := this.storageMap[policyId]
		if ok {
			// 检查配置是否有变更
			if types.Int(policy.Version) != storage.Version() {
				err = storage.Close()
				if err != nil {
					remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "close policy '"+types.String(policyId)+"' failed: "+err.Error())

					// 继续往下执行
				}

				if len(policy.Options) > 0 {
					err = json.Unmarshal(policy.Options, storage.Config())
					if err != nil {
						remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "unmarshal policy '"+types.String(policyId)+"' config failed: "+err.Error())
						storage.SetOk(false)
						continue
					}
				}

				storage.SetVersion(types.Int(policy.Version))
				err := storage.Start()
				if err != nil {
					remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "start policy '"+types.String(policyId)+"' failed: "+err.Error())
					continue
				}
				storage.SetOk(true)
				remotelogs.Println("ACCESS_LOG_STORAGE_MANAGER", "restart policy '"+types.String(policyId)+"'")
			}
		} else {
			storage, err := this.createStorage(policy.Type, policy.Options)
			if err != nil {
				remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "create policy '"+types.String(policyId)+"' failed: "+err.Error())
				continue
			}
			storage.SetVersion(types.Int(policy.Version))
			this.storageMap[policyId] = storage
			err = storage.Start()
			if err != nil {
				remotelogs.Error("ACCESS_LOG_STORAGE_MANAGER", "start policy '"+types.String(policyId)+"' failed: "+err.Error())
				continue
			}
			storage.SetOk(true)
			remotelogs.Println("ACCESS_LOG_STORAGE_MANAGER", "start policy '"+types.String(policyId)+"'")
		}
	}

	return nil
}

func (this *StorageManager) createStorage(storageType string, optionsJSON []byte) (StorageInterface, error) {
	switch storageType {
	case serverconfigs.AccessLogStorageTypeFile:
		var config = &serverconfigs.AccessLogFileStorageConfig{}
		if len(optionsJSON) > 0 {
			err := json.Unmarshal(optionsJSON, config)
			if err != nil {
				return nil, err
			}
		}
		return NewFileStorage(config), nil
	case serverconfigs.AccessLogStorageTypeES:
		var config = &serverconfigs.AccessLogESStorageConfig{}
		if len(optionsJSON) > 0 {
			err := json.Unmarshal(optionsJSON, config)
			if err != nil {
				return nil, err
			}
		}
		return NewESStorage(config), nil
	case serverconfigs.AccessLogStorageTypeTCP:
		var config = &serverconfigs.AccessLogTCPStorageConfig{}
		if len(optionsJSON) > 0 {
			err := json.Unmarshal(optionsJSON, config)
			if err != nil {
				return nil, err
			}
		}
		return NewTCPStorage(config), nil
	case serverconfigs.AccessLogStorageTypeSyslog:
		var config = &serverconfigs.AccessLogSyslogStorageConfig{}
		if len(optionsJSON) > 0 {
			err := json.Unmarshal(optionsJSON, config)
			if err != nil {
				return nil, err
			}
		}
		return NewSyslogStorage(config), nil
	case serverconfigs.AccessLogStorageTypeCommand:
		var config = &serverconfigs.AccessLogCommandStorageConfig{}
		if len(optionsJSON) > 0 {
			err := json.Unmarshal(optionsJSON, config)
			if err != nil {
				return nil, err
			}
		}
		return NewCommandStorage(config), nil
	}

	return nil, errors.New("invalid policy type '" + storageType + "'")
}
