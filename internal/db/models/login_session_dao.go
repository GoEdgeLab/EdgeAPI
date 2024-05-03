package models

import (
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"time"
)

// TODO 定时清理过期的SESSION

type LoginSessionDAO dbs.DAO

func NewLoginSessionDAO() *LoginSessionDAO {
	return dbs.NewDAO(&LoginSessionDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeLoginSessions",
			Model:  new(LoginSession),
			PkName: "id",
		},
	}).(*LoginSessionDAO)
}

var SharedLoginSessionDAO *LoginSessionDAO

func init() {
	dbs.OnReady(func() {
		SharedLoginSessionDAO = NewLoginSessionDAO()
	})
}

// CreateSession 创建SESSION
func (this *LoginSessionDAO) CreateSession(tx *dbs.Tx, sid string, ip string, expiresAt int64) (int64, error) {
	if len(sid) == 0 || len(sid) > 64 {
		return 0, errors.New("invalid 'sid'")
	}

	// 是否已存在
	oldSessionId, err := this.Query(tx).
		Attr("sid", sid).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	var op = NewLoginSessionOperator()
	if oldSessionId > 0 {
		op.Id = oldSessionId
	}

	op.Sid = sid
	op.Ip = ip
	op.Values = "{}"
	op.ExpiresAt = expiresAt
	op.CreatedAt = time.Now().Unix()

	if oldSessionId > 0 {
		err := this.Save(tx, op)
		if err != nil {
			return 0, err
		}
		return oldSessionId, nil
	}
	return this.SaveInt64(tx, op)
}

// WriteSessionValue 向SESSION中写入数据
func (this *LoginSessionDAO) WriteSessionValue(tx *dbs.Tx, sid string, key string, value any) error {
	if len(sid) == 0 || len(sid) > 64 {
		return errors.New("invalid 'sid'")
	}

	// 是否存在
	sessionOne, err := this.Query(tx).
		Attr("sid", sid).
		Find()
	if err != nil {
		return err
	}
	var sessionId int64
	var valueMap = maps.Map{}
	if sessionOne != nil {
		var session = sessionOne.(*LoginSession)
		if session.IsAvailable() {
			sessionId = int64(session.Id)

			if !IsNull(session.Values) {
				err = json.Unmarshal(session.Values, &valueMap)
				if err != nil {
					return err
				}
			}
		} else {
			// 不可用则删除之
			err = this.Query(tx).
				Pk(session.Id).
				DeleteQuickly()
			if err != nil {
				return err
			}
		}
	}
	if sessionId == 0 {
		// 不存在，则创建之
		sessionId, err = this.CreateSession(tx, sid, "", time.Now().Unix()+30*86400 /** 默认30天**/)
		if err != nil {
			return err
		}
	}

	var sessionOp = NewLoginSessionOperator()
	sessionOp.Id = sessionId

	// 获取用户ID
	var adminId int64
	var userId int64

	switch key {
	case "adminId":
		adminId = types.Int64(value)
	case "userId":
		userId = types.Int64(value)
	}

	if adminId > 0 || userId > 0 {
		sessionOp.AdminId = adminId
		sessionOp.UserId = userId
	}

	// 写入数据
	valueMap[key] = value
	sessionOp.Values = valueMap.AsJSON()

	// IP
	if key == "@ip" {
		sessionOp.Ip = value
	}

	return this.Save(tx, sessionOp)
}

// DeleteSession 删除SESSION
func (this *LoginSessionDAO) DeleteSession(tx *dbs.Tx, sid string) error {
	return this.Query(tx).
		Attr("sid", sid).
		DeleteQuickly()
}

// FindSession 查询SESSION
func (this *LoginSessionDAO) FindSession(tx *dbs.Tx, sid string) (*LoginSession, error) {
	one, err := this.Query(tx).
		Attr("sid", sid).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	var session = one.(*LoginSession)

	// 不可用则删除
	if !session.IsAvailable() {
		err = this.Query(tx).
			Pk(session.Id).
			DeleteQuickly()
		if err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (this *LoginSessionDAO) ClearOldSessions(tx *dbs.Tx, adminId int64, userId int64, sid string, ip string) error {
	// 删除此用户之前创建的SESSION
	err := this.Query(tx).
		Attr("adminId", adminId).
		Attr("userId", userId).
		Neq("sid", sid).
		Neq("ip", ip). // 同一个IP允许多个SID，因为有人可能会同时使用手机端和PC端
		DeleteQuickly()
	if err != nil {
		return err
	}

	// 删除过多的SESSION
	oldOnes, queryErr := this.Query(tx).
		ResultPk().
		Attr("adminId", adminId).
		Attr("userId", userId).
		Neq("sid", sid).
		AscPk().
		FindAll()
	if queryErr != nil {
		return queryErr
	}
	var oldCount = len(oldOnes)
	if oldCount > 3 {
		for _, oldOne := range oldOnes[:oldCount-3] {
			var oldId = oldOne.(*LoginSession).Id
			if oldOne.(*LoginSession).Sid == sid {
				continue
			}
			err = this.Query(tx).
				Pk(oldId).
				DeleteQuickly()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
