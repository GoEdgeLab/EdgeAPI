package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/zero"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"time"
)

type SysLockerDAO dbs.DAO

// concurrent transactions control
// 考虑到存在多个API节点的可能性，容量不能太大，也不能使用mutex
var sysLockerConcurrentLimiter = make(chan zero.Zero, 8)

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

// Lock 开锁
func (this *SysLockerDAO) Lock(tx *dbs.Tx, key string, timeout int64) (ok bool, err error) {
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
			var op = NewSysLockerOperator()
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
		var op = NewSysLockerOperator()
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

// Unlock 解锁
func (this *SysLockerDAO) Unlock(tx *dbs.Tx, key string) error {
	_, err := this.Query(tx).
		Attr("key", key).
		Set("timeoutAt", time.Now().Unix()-86400*365).
		Update()
	return err
}

// Increase 增加版本号
func (this *SysLockerDAO) Increase(tx *dbs.Tx, key string, defaultValue int64) (int64, error) {
	if tx == nil {
		var result int64
		var err error

		sysLockerConcurrentLimiter <- zero.Zero{} // push
		defer func() {
			<-sysLockerConcurrentLimiter // pop
		}()

		err = this.Instance.RunTx(func(tx *dbs.Tx) error {
			result, err = this.Increase(tx, key, defaultValue)
			if err != nil {
				return err
			}
			return nil
		})
		return result, err
	}
	err := this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"key":     key,
			"version": defaultValue,
		}, maps.Map{
			"version": dbs.SQL("version+1"),
		})
	if err != nil {
		return 0, err
	}
	return this.Query(tx).
		Attr("key", key).
		Result("version").
		FindInt64Col(0)
}

// 读取当前版本号
func (this *SysLockerDAO) Read(tx *dbs.Tx, key string) (int64, error) {
	return this.Query(tx).
		Attr("key", key).
		Result("version").
		FindInt64Col(0)
}
