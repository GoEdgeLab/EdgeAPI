package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"time"
)

type SysLockerDAO dbs.DAO

func NewSysLockerDAO() *SysLockerDAO {
	return dbs.NewDAO(&SysLockerDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeSysLockers",
			Model:  new(SysLocker),
			PkName: "id",
		},
	}).(*SysLockerDAO)
}

var SharedSysLockerDAO *SysLockerDAO

func init() {
	dbs.OnReady(func() {
		SharedSysLockerDAO = NewSysLockerDAO()
	})
}

// 开锁
func (this *SysLockerDAO) Lock(tx *dbs.Tx, key string, timeout int64) (bool, error) {
	maxErrors := 5
	for {
		one, err := this.Query(tx).
			Attr("key", key).
			Find()
		if err != nil {
			maxErrors--
			if maxErrors < 0 {
				return false, err
			}
			continue
		}

		// 如果没有锁，则创建
		if one == nil {
			op := NewSysLockerOperator()
			op.Key = key
			op.TimeoutAt = time.Now().Unix() + timeout
			op.Version = 1
			err := this.Save(tx, op)
			if err != nil {
				maxErrors--
				if maxErrors < 0 {
					return false, err
				}
				continue
			}

			return true, nil
		}

		// 如果已经有锁
		locker := one.(*SysLocker)
		if time.Now().Unix() <= int64(locker.TimeoutAt) {
			return false, nil
		}

		// 修改
		op := NewSysLockerOperator()
		op.Id = locker.Id
		op.Version = locker.Version + 1
		op.TimeoutAt = time.Now().Unix() + timeout
		err = this.Save(tx, op)
		if err != nil {
			maxErrors--
			if maxErrors < 0 {
				return false, err
			}
			continue
		}

		// 再次查询版本
		version, err := this.Query(tx).
			Attr("key", key).
			Result("version").
			FindCol("0")
		if err != nil {
			maxErrors--
			if maxErrors < 0 {
				return false, err
			}
			continue
		}
		if types.Int64(version) != int64(locker.Version)+1 {
			return false, nil
		}

		return true, nil
	}
}

// 解锁
func (this *SysLockerDAO) Unlock(tx *dbs.Tx, key string) error {
	_, err := this.Query(tx).
		Attr("key", key).
		Set("timeoutAt", time.Now().Unix()-86400*365).
		Update()
	return err
}
