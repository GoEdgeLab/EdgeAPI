package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"strings"
)

type LogDAO dbs.DAO

func NewLogDAO() *LogDAO {
	return dbs.NewDAO(&LogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeLogs",
			Model:  new(Log),
			PkName: "id",
		},
	}).(*LogDAO)
}

var SharedLogDAO *LogDAO

func init() {
	dbs.OnReady(func() {
		SharedLogDAO = NewLogDAO()
	})
}

// 创建管理员日志
func (this *LogDAO) CreateLog(adminType string, adminId int64, level string, description string, action string, ip string) error {
	op := NewLogOperator()
	op.Type = adminType
	op.AdminId, op.Level, op.Description, op.Action, op.Ip = adminId, level, description, action, ip
	op.Day = timeutil.Format("Ymd")
	op.Type = LogTypeAdmin
	_, err := this.Save(op)
	return err
}

// 计算所有日志数量
func (this *LogDAO) CountLogs(dayFrom string, dayTo string, keyword string) (int64, error) {
	dayFrom = this.formatDay(dayFrom)
	dayTo = this.formatDay(dayTo)

	query := this.Query()

	if len(dayFrom) > 0 {
		query.Gte("day", dayFrom)
	}
	if len(dayTo) > 0 {
		query.Lte("day", dayTo)
	}
	if len(keyword) > 0 {
		query.Where("(description LIKE :keyword OR ip LIKE :keyword OR action LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}

	return query.Count()
}

// 列出单页日志
func (this *LogDAO) ListLogs(offset int64, size int64, dayFrom string, dayTo string, keyword string) (result []*Log, err error) {
	dayFrom = this.formatDay(dayFrom)
	dayTo = this.formatDay(dayTo)

	query := this.Query()
	if len(dayFrom) > 0 {
		query.Gte("day", dayFrom)
	}
	if len(dayTo) > 0 {
		query.Lte("day", dayTo)
	}
	if len(keyword) > 0 {
		query.Where("(description LIKE :keyword OR ip LIKE :keyword OR action LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}

	_, err = query.
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// 格式化日期
func (this *LogDAO) formatDay(day string) string {
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(day) {
		return ""
	}
	day = strings.ReplaceAll(day, "-", "")
	return day
}
