package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"time"
)

type APIAccessTokenDAO dbs.DAO

func NewAPIAccessTokenDAO() *APIAccessTokenDAO {
	return dbs.NewDAO(&APIAccessTokenDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeAPIAccessTokens",
			Model:  new(APIAccessToken),
			PkName: "id",
		},
	}).(*APIAccessTokenDAO)
}

var SharedAPIAccessTokenDAO *APIAccessTokenDAO

func init() {
	dbs.OnReady(func() {
		SharedAPIAccessTokenDAO = NewAPIAccessTokenDAO()
	})
}

// GenerateAccessToken 生成AccessToken
func (this *APIAccessTokenDAO) GenerateAccessToken(tx *dbs.Tx, adminId int64, userId int64) (token string, expiresAt int64, err error) {
	if adminId <= 0 && userId <= 0 {
		err = errors.New("either 'adminId' or 'userId' should not be zero")
		return
	}

	if adminId > 0 {
		userId = 0
	}
	if userId > 0 {
		adminId = 0
	}

	// 查询以前的
	accessToken, err := this.Query(tx).
		Attr("adminId", adminId).
		Attr("userId", userId).
		Find()
	if err != nil {
		return "", 0, err
	}

	token = rands.String(128) // TODO 增强安全性，将来使用 base64_encode(encrypt(salt+random)) 算法来代替
	expiresAt = time.Now().Unix() + 7200

	op := NewAPIAccessTokenOperator()

	if accessToken != nil {
		op.Id = accessToken.(*APIAccessToken).Id
	}

	op.AdminId = adminId
	op.UserId = userId
	op.Token = token
	op.CreatedAt = time.Now().Unix()
	op.ExpiredAt = expiresAt
	err = this.Save(tx, op)
	return
}

// FindAccessToken 查找AccessToken
func (this *APIAccessTokenDAO) FindAccessToken(tx *dbs.Tx, token string) (*APIAccessToken, error) {
	one, err := this.Query(tx).
		Attr("token", token).
		Find()
	if one == nil || err != nil {
		return nil, err
	}
	return one.(*APIAccessToken), nil
}
