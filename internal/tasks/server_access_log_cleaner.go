package tasks

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"strings"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewServerAccessLogCleaner(6 * time.Hour).Start()
		})
	})
}

// ServerAccessLogCleaner 服务访问日志自动清理
type ServerAccessLogCleaner struct {
	BaseTask

	ticker *time.Ticker
}

func NewServerAccessLogCleaner(duration time.Duration) *ServerAccessLogCleaner {
	return &ServerAccessLogCleaner{
		ticker: time.NewTicker(duration),
	}
}

func (this *ServerAccessLogCleaner) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("[TASK][ServerAccessLogCleaner]", err.Error())
		}
	}
}

func (this *ServerAccessLogCleaner) Loop() error {
	// 当前设置
	configJSON, err := models.SharedSysSettingDAO.ReadSetting(nil, systemconfigs.SettingCodeDatabaseConfigSetting)
	if err != nil {
		return err
	}
	if len(configJSON) == 0 {
		return nil
	}
	var config = &systemconfigs.DatabaseConfig{}
	err = json.Unmarshal(configJSON, config)
	if err != nil {
		return err
	}
	if config.ServerAccessLog.Clean.Days <= 0 {
		return nil
	}
	var days = config.ServerAccessLog.Clean.Days
	var endDay = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days+1))

	// 当前连接的数据库
	db, err := dbs.Default()
	if err != nil {
		return err
	}
	err = this.cleanDB(db, endDay)
	if err != nil {
		return err
	}

	// 日志数据库节点
	nodes, err := models.SharedDBNodeDAO.FindAllEnabledAndOnDBNodes(nil)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		err := func(node *models.DBNode) error {
			var dbConfig = node.DBConfig()
			nodeDB, err := dbs.NewInstanceFromConfig(dbConfig)
			if err != nil {
				return err
			}

			defer func() {
				_ = nodeDB.Close()
			}()

			err = this.cleanDB(nodeDB, endDay)
			if err != nil {
				return err
			}

			return nil
		}(node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *ServerAccessLogCleaner) cleanDB(db *dbs.DB, endDay string) error {
	ones, columnNames, err := db.FindPreparedOnes("SHOW TABLES")
	if err != nil {
		return err
	}
	if len(columnNames) != 1 {
		return errors.New("invalid column names: " + strings.Join(columnNames, ", "))
	}
	var columnName = columnNames[0]
	var reg = regexp.MustCompile(`^(?i)(edgeHTTPAccessLogs|edgeNSAccessLogs)_(\d{8})(_\d{4})?$`)
	for _, one := range ones {
		var tableName = one.GetString(columnName)
		if len(tableName) == 0 {
			continue
		}
		if !reg.MatchString(tableName) {
			continue
		}
		var matches = reg.FindStringSubmatch(tableName)
		var day = matches[2]

		if day < endDay {
			_, err = db.Exec("DROP TABLE " + tableName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
