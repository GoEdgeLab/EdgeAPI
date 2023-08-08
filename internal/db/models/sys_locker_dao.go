package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strings"
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
		var locker = one.(*SysLocker)
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
		if types.Int64(version) > int64(locker.Version)+1 {
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

const sysLockerStep = 8

var increment = NewSysLockerIncrement(sysLockerStep)

// Increase 增加版本号
func (this *SysLockerDAO) Increase(tx *dbs.Tx, key string, defaultValue int64) (int64, error) {
	// validate key
	if strings.Contains(key, "'") {
		return 0, errors.New("invalid key '" + key + "'")
	}

	if tx == nil {
		var result int64
		var err error

		{
			colValue, err := this.Query(tx).
				Result("version").
				Attr("key", key).
				FindInt64Col(0)
			if err != nil {
				return 0, err
			}
			var lastVersion = types.Int64(colValue)
			if lastVersion <= increment.MaxValue(key) {
				value, ok := increment.Pop(key)
				if ok {
					return value, nil
				}
			}
		}

		err = this.Instance.RunTx(func(tx *dbs.Tx) error {
			result, err = this.Increase(tx, key, defaultValue)
			if err != nil {
				return err
			}
			return nil
		})
		return result, err
	}

	// combine statements to make increasing faster
	colValue, err := tx.FindCol(0, "INSERT INTO `"+this.Table+"` (`key`, `version`) VALUES ('"+key+"', "+types.String(defaultValue+sysLockerStep)+") ON DUPLICATE KEY UPDATE `version`=`version`+"+types.String(sysLockerStep)+"; SELECT `version` FROM `"+this.Table+"` WHERE `key`='"+key+"'")
	if err != nil {
		if CheckSQLErrCode(err, 1064 /** syntax error **/) {
			// continue to use separated query
			err = nil
		} else {
			return 0, err
		}
	} else {
		var maxVersion = types.Int64(colValue)
		var minVersion = maxVersion - sysLockerStep + 1
		increment.Push(key, minVersion+1, maxVersion)

		return minVersion, nil
	}

	err = this.Query(tx).
		Reuse(false). // no need to prepare statement in every transaction
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
		Reuse(false). // no need to prepare statement in every transaction
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
