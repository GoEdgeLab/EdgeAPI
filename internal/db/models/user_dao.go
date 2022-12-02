package models

import (
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

const (
	UserStateEnabled  = 1 // 已启用
	UserStateDisabled = 0 // 已禁用
)

type UserDAO dbs.DAO

func NewUserDAO() *UserDAO {
	return dbs.NewDAO(&UserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUsers",
			Model:  new(User),
			PkName: "id",
		},
	}).(*UserDAO)
}

var SharedUserDAO *UserDAO

func init() {
	dbs.OnReady(func() {
		SharedUserDAO = NewUserDAO()
	})
}

// EnableUser 启用条目
func (this *UserDAO) EnableUser(tx *dbs.Tx, userId int64) error {
	if userId <= 0 {
		return errors.New("invalid 'userId'")
	}

	_, err := this.Query(tx).
		Pk(userId).
		Set("state", UserStateEnabled).
		Update()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, userId)
}

// DisableUser 禁用条目
func (this *UserDAO) DisableUser(tx *dbs.Tx, userId int64) error {
	if userId <= 0 {
		return errors.New("invalid 'userId'")
	}

	_, err := this.Query(tx).
		Pk(userId).
		Set("state", UserStateDisabled).
		Update()
	if err != nil {
		return err
	}

	err = SharedAPIAccessTokenDAO.DeleteAccessTokens(tx, 0, userId)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, userId)
}

// FindEnabledUser 查找启用的用户
func (this *UserDAO) FindEnabledUser(tx *dbs.Tx, userId int64, cacheMap *utils.CacheMap) (*User, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":FindEnabledUser:" + types.String(userId)
	cache, ok := cacheMap.Get(cacheKey)
	if ok {
		return cache.(*User), nil
	}

	result, err := this.Query(tx).
		Pk(userId).
		Attr("state", UserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, result)
	}

	return result.(*User), err
}

// FindEnabledBasicUser 查找用户基本信息
func (this *UserDAO) FindEnabledBasicUser(tx *dbs.Tx, id int64) (*User, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserStateEnabled).
		Result("id", "fullname", "username").
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*User), err
}

// FindBasicUserWithoutState 查找用户基本信息，并忽略状态
func (this *UserDAO) FindBasicUserWithoutState(tx *dbs.Tx, id int64) (*User, error) {
	result, err := this.Query(tx).
		Pk(id).
		Result("id", "fullname", "username").
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*User), err
}

// FindEnabledUserIdWithUsername 根据用户名查找用户ID
func (this *UserDAO) FindEnabledUserIdWithUsername(tx *dbs.Tx, username string) (int64, error) {
	return this.Query(tx).
		ResultPk().
		State(UserStateEnabled).
		Attr("username", username).
		FindInt64Col(0)
}

// FindUserFullname 获取管理员名称
func (this *UserDAO) FindUserFullname(tx *dbs.Tx, userId int64) (string, error) {
	return this.Query(tx).
		Pk(userId).
		Result("fullname").
		FindStringCol("")
}

// CreateUser 创建用户
func (this *UserDAO) CreateUser(tx *dbs.Tx, username string,
	password string,
	fullname string,
	mobile string,
	tel string,
	email string,
	remark string,
	source string,
	clusterId int64,
	features []string,
	registeredIP string,
	isVerified bool) (int64, error) {
	var op = NewUserOperator()
	op.Username = username
	op.Password = stringutil.Md5(password)
	op.Fullname = fullname
	op.Mobile = mobile
	op.Tel = tel
	op.Email = email
	op.EmailIsVerified = false
	op.Remark = remark
	op.Source = source
	op.ClusterId = clusterId
	op.Day = timeutil.Format("Ymd")
	op.IsVerified = isVerified
	op.RegisteredIP = registeredIP

	// features
	if len(features) == 0 {
		op.Features = "[]"
	} else {
		featuresJSON, err := json.Marshal(features)
		if err != nil {
			return 0, err
		}
		op.Features = featuresJSON
	}

	op.IsOn = true
	op.State = UserStateEnabled
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateUser 修改用户
func (this *UserDAO) UpdateUser(tx *dbs.Tx, userId int64, username string, password string, fullname string, mobile string, tel string, email string, remark string, isOn bool, nodeClusterId int64) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}

	var op = NewUserOperator()
	op.Id = userId
	op.Username = username
	if len(password) > 0 {
		op.Password = stringutil.Md5(password)
	}
	op.Fullname = fullname
	op.Mobile = mobile
	op.Tel = tel
	op.Email = email
	op.Remark = remark
	op.ClusterId = nodeClusterId
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	// 删除AccessTokens
	if !isOn {
		err = SharedAPIAccessTokenDAO.DeleteAccessTokens(tx, 0, userId)
		if err != nil {
			return err
		}
	}

	return this.NotifyUpdate(tx, userId)
}

// UpdateUserInfo 修改用户基本信息
func (this *UserDAO) UpdateUserInfo(tx *dbs.Tx, userId int64, fullname string, mobile string, email string) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	var op = NewUserOperator()
	op.Id = userId
	op.Fullname = fullname
	op.Mobile = mobile
	op.Email = email
	return this.Save(tx, op)
}

// UpdateUserLogin 修改用户登录信息
func (this *UserDAO) UpdateUserLogin(tx *dbs.Tx, userId int64, username string, password string) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	var op = NewUserOperator()
	op.Id = userId
	op.Username = username
	if len(password) > 0 {
		op.Password = stringutil.Md5(password)
	}
	err := this.Save(tx, op)
	return err
}

// CountAllEnabledUsers 计算用户数量
func (this *UserDAO) CountAllEnabledUsers(tx *dbs.Tx, clusterId int64, keyword string, isVerifying bool) (int64, error) {
	query := this.Query(tx)
	query.State(UserStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword OR mobile LIKE :keyword OR email LIKE :keyword OR tel LIKE :keyword OR remark LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if isVerifying {
		query.Where("(isVerified=0 OR (id IN (SELECT userId FROM " + SharedUserIdentityDAO.Table + " WHERE status=:identityStatus AND state=1)))")
		query.Param("identityStatus", userconfigs.UserIdentityStatusSubmitted)
	}
	return query.Count()
}

// CountAllVerifyingUsers 获取等待审核的用户数
func (this *UserDAO) CountAllVerifyingUsers(tx *dbs.Tx) (int64, error) {
	query := this.Query(tx)
	query.State(UserStateEnabled)
	query.Where("(isVerified=0 OR (id IN (SELECT userId FROM " + SharedUserIdentityDAO.Table + " WHERE status=:identityStatus AND state=1)))")
	query.Param("identityStatus", userconfigs.UserIdentityStatusSubmitted)
	return query.Count()
}

// ListEnabledUsers 列出单页用户
func (this *UserDAO) ListEnabledUsers(tx *dbs.Tx, clusterId int64, keyword string, isVerifying bool, offset int64, size int64) (result []*User, err error) {
	query := this.Query(tx)
	query.State(UserStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword OR mobile LIKE :keyword OR email LIKE :keyword OR tel LIKE :keyword OR remark LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if isVerifying {
		query.Where("(isVerified=0 OR (id IN (SELECT userId FROM " + SharedUserIdentityDAO.Table + " WHERE status=:identityStatus AND state=1)))")
		query.Param("identityStatus", userconfigs.UserIdentityStatusSubmitted)
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// ExistUser 检查用户名是否存在
func (this *UserDAO) ExistUser(tx *dbs.Tx, userId int64, username string) (bool, error) {
	return this.Query(tx).
		State(UserStateEnabled).
		Attr("username", username).
		Neq("id", userId).
		Exist()
}

// ListEnabledUserIds 列出单页的用户ID
func (this *UserDAO) ListEnabledUserIds(tx *dbs.Tx, offset, size int64) ([]int64, error) {
	ones, _, err := this.Query(tx).
		ResultPk().
		State(UserStateEnabled).
		Offset(offset).
		Limit(size).
		AscPk().
		FindOnes()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, one.GetInt64("id"))
	}
	return result, nil
}

// CheckUserPassword 检查用户名、密码
func (this *UserDAO) CheckUserPassword(tx *dbs.Tx, username string, encryptedPassword string) (int64, error) {
	if len(username) == 0 || len(encryptedPassword) == 0 {
		return 0, nil
	}
	return this.Query(tx).
		Attr("username", username).
		Attr("password", encryptedPassword).
		Attr("state", UserStateEnabled).
		Attr("isOn", true).
		ResultPk().
		FindInt64Col(0)
}

// FindUserClusterId 查找用户所在集群
func (this *UserDAO) FindUserClusterId(tx *dbs.Tx, userId int64) (int64, error) {
	return this.Query(tx).
		Pk(userId).
		Result("clusterId").
		FindInt64Col(0)
}

// UpdateUserFeatures 更新单个用户Features
func (this *UserDAO) UpdateUserFeatures(tx *dbs.Tx, userId int64, featuresJSON []byte) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	if len(featuresJSON) == 0 {
		featuresJSON = []byte("[]")
	}
	_, err := this.Query(tx).
		Pk(userId).
		Set("features", featuresJSON).
		Update()
	if err != nil {
		return err
	}
	return nil
}

// UpdateUsersFeatures 更新所有用户的Features
func (this *UserDAO) UpdateUsersFeatures(tx *dbs.Tx, featureCodes []string, overwrite bool) error {
	if featureCodes == nil {
		featureCodes = []string{}
	}
	if overwrite {
		featureCodesJSON, err := json.Marshal(featureCodes)
		if err != nil {
			return err
		}
		err = this.Query(tx).
			State(UserStateEnabled).
			Set("features", featureCodesJSON).
			UpdateQuickly()
		return err
	}

	var lastId int64
	const size = 1000
	for {
		ones, _, err := this.Query(tx).
			Result("id", "features").
			State(UserStateEnabled).
			Gt("id", lastId).
			Limit(size).
			AscPk().
			FindOnes()
		if err != nil {
			return err
		}
		for _, one := range ones {
			var userId = one.GetInt64("id")
			var userFeaturesJSON = one.GetBytes("features")
			var userFeatures = []string{}
			if len(userFeaturesJSON) > 0 {
				err = json.Unmarshal(userFeaturesJSON, &userFeatures)
				if err != nil {
					return err
				}
			}
			for _, featureCode := range featureCodes {
				if !lists.ContainsString(userFeatures, featureCode) {
					userFeatures = append(userFeatures, featureCode)
				}
			}
			userFeaturesJSON, err = json.Marshal(userFeatures)
			if err != nil {
				return err
			}
			err = this.Query(tx).
				Pk(userId).
				Set("features", userFeaturesJSON).
				UpdateQuickly()
			if err != nil {
				return err
			}
		}

		if len(ones) < size {
			break
		}

		lastId += size
	}

	return nil
}

// FindUserFeatures 查找用户Features
func (this *UserDAO) FindUserFeatures(tx *dbs.Tx, userId int64) ([]*userconfigs.UserFeature, error) {
	featuresJSON, err := this.Query(tx).
		Pk(userId).
		Result("features").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(featuresJSON) == 0 {
		return nil, nil
	}

	featureCodes := []string{}
	err = json.Unmarshal([]byte(featuresJSON), &featureCodes)
	if err != nil {
		return nil, err
	}

	// 检查是否还存在以及设置名称
	result := []*userconfigs.UserFeature{}
	if len(featureCodes) > 0 {
		for _, featureCode := range featureCodes {
			f := userconfigs.FindUserFeature(featureCode)
			if f != nil {
				result = append(result, &userconfigs.UserFeature{Name: f.Name, Code: f.Code, Description: f.Description})
			}
		}
	}

	return result, nil
}

// SumDailyUsers 获取当天用户数量
func (this *UserDAO) SumDailyUsers(tx *dbs.Tx, dayFrom string, dayTo string) (int64, error) {
	return this.Query(tx).
		Between("day", dayFrom, dayTo).
		State(UserStateEnabled).
		Count()
}

// CountDailyUsers 计算每天用户数
func (this *UserDAO) CountDailyUsers(tx *dbs.Tx, dayFrom string, dayTo string) ([]*pb.ComposeUserGlobalBoardResponse_DailyStat, error) {
	ones, _, err := this.Query(tx).
		Result("COUNT(*) AS count", "day").
		Between("day", dayFrom, dayTo).
		State(UserStateEnabled).
		Group("day").
		FindOnes()
	if err != nil {
		return nil, err
	}
	var m = map[string]*pb.ComposeUserGlobalBoardResponse_DailyStat{} // day => Stat
	for _, one := range ones {
		m[one.GetString("day")] = &pb.ComposeUserGlobalBoardResponse_DailyStat{
			Day:   one.GetString("day"),
			Count: one.GetInt64("count"),
		}
	}

	var result = []*pb.ComposeUserGlobalBoardResponse_DailyStat{}
	days, err := utils.RangeDays(dayFrom, dayTo)
	if err != nil {
		return nil, err
	}
	for _, day := range days {
		stat, ok := m[day]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &pb.ComposeUserGlobalBoardResponse_DailyStat{
				Day:   day,
				Count: 0,
			})
		}
	}

	return result, nil
}

// UpdateUserIsVerified 审核用户
func (this *UserDAO) UpdateUserIsVerified(tx *dbs.Tx, userId int64, isRejected bool, rejectReason string) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	var op = NewUserOperator()
	op.Id = userId
	op.IsRejected = isRejected
	op.RejectReason = rejectReason
	op.IsVerified = true
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, userId)
}

// RenewUserServersState 更新用户服务状态
func (this *UserDAO) RenewUserServersState(tx *dbs.Tx, userId int64) (bool, error) {
	oldServersEnabled, err := this.Query(tx).
		Pk(userId).
		Result("serversEnabled").
		FindBoolCol()
	if err != nil {
		return false, err
	}

	newServersEnabled, err := this.CheckUserServersEnabled(tx, userId)
	if err != nil {
		return false, err
	}

	if oldServersEnabled != newServersEnabled {
		err = this.Query(tx).
			Pk(userId).
			Set("serversEnabled", newServersEnabled).
			UpdateQuickly()
		if err != nil {
			return false, err
		}

		// 创建变更通知
		clusterIds, err := SharedServerDAO.FindUserServerClusterIds(tx, userId)
		if err != nil {
			return false, err
		}
		for _, clusterId := range clusterIds {
			err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, userId, 0, NodeTaskTypeUserServersStateChanged)
			if err != nil {
				return false, err
			}
		}
	}

	return newServersEnabled, nil
}

// NotifyUpdate 用户变更通知
func (this *UserDAO) NotifyUpdate(tx *dbs.Tx, userId int64) error {
	if userId <= 0 {
		return nil
	}

	// 更新用户服务状态
	_, err := this.RenewUserServersState(tx, userId)
	if err != nil {
		return err
	}

	return nil
}
