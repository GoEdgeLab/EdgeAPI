package tasks

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"strings"
	"time"
)

func init() {
	dbs.OnReady(func() {
		task := NewServerAccessLogCleaner()
		go task.Start()
	})
}

// 服务访问日志自动清理
type ServerAccessLogCleaner struct {
}

func NewServerAccessLogCleaner() *ServerAccessLogCleaner {
	return &ServerAccessLogCleaner{}
}

func (this *ServerAccessLogCleaner) Start() {
	ticker := time.NewTicker(12 * time.Hour)
	for range ticker.C {
		err := this.Loop()
		if err != nil {
			logs.Println("[TASK][ServerAccessLogCleaner]Error: " + err.Error())
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
	config := &systemconfigs.DatabaseConfig{}
	err = json.Unmarshal(configJSON, config)
	if err != nil {
		return err
	}
	if config.ServerAccessLog.Clean.Days <= 0 {
		return nil
	}
	days := config.ServerAccessLog.Clean.Days
	endDay := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days+1))

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
		dbConfig := node.DBConfig()
		db, err := dbs.NewInstanceFromConfig(dbConfig)
		if err != nil {
			return err
		}
		err = this.cleanDB(db, endDay)
		if err != nil {
			_ = db.Close()
			return err
		}

		_ = db.Close()
	}

	return nil
}

func (this *ServerAccessLogCleaner) cleanDB(db *dbs.DB, endDay string) error {
	ones, columnNames, err := db.FindOnes("SHOW TABLES")
	if err != nil {
		return err
	}
	if len(columnNames) != 1 {
		return errors.New("invalid column names: " + strings.Join(columnNames, ", "))
	}
	columnName := columnNames[0]
	for _, one := range ones {
		tableName := one.GetString(columnName)
		if len(tableName) == 0 {
			continue
		}
		ok, err := regexp.MatchString(`^edgeHTTPAccessLogs_(\d{8})$`, tableName)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		index := strings.LastIndex(tableName, "_")
		if index < 0 {
			continue
		}
		day := tableName[index+1:]
		if day < endDay {
			_, err = db.Exec("DROP TABLE " + tableName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
