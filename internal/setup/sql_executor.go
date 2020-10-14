package setup

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/setup/sqls"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io/ioutil"
	"sort"
	"strings"
	"time"
)

// 安装或升级SQL执行器
type SQLExecutor struct {
	dbConfig *dbs.DBConfig
}

func NewSQLExecutor(dbConfig *dbs.DBConfig) *SQLExecutor {
	return &SQLExecutor{
		dbConfig: dbConfig,
	}
}

func NewSQLExecutorFromCmd() (*SQLExecutor, error) {
	// 执行SQL
	config := &dbs.Config{}
	configData, err := ioutil.ReadFile(Tea.ConfigFile("db.yaml"))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err
	}
	return NewSQLExecutor(config.DBs[Tea.Env]), nil
}

func (this *SQLExecutor) Run() error {
	db, err := dbs.NewInstanceFromConfig(this.dbConfig)
	if err != nil {
		return err
	}

	tableNames, err := db.TableNames()
	if err != nil {
		return err
	}

	// 检查已有数据
	if lists.ContainsString(tableNames, "edgeVersions") {
		stmt, err := db.Prepare("SELECT version FROM " + db.TablePrefix() + "Versions")
		if err != nil {
			return err
		}
		defer func() {
			_ = stmt.Close()
		}()

		rows, err := stmt.Query()
		if err != nil {
			return err
		}
		defer func() {
			_ = rows.Close()
		}()

		// 对比版本
		oldVersion := ""
		if rows.Next() {
			err = rows.Scan(&oldVersion)
			if err != nil {
				return errors.New("query version failed: " + err.Error())
			}
		}
		if oldVersion == teaconst.Version {
			err = this.checkData(db)
			if err != nil {
				return err
			}

			return nil
		}
		oldVersion = strings.Replace(oldVersion, "_", ".", -1)

		upgradeVersions := []string{}
		sqlMap := map[string]string{} // version => sql

		for _, m := range sqls.SQLVersions {
			version, _ := m["version"]
			if version == "full" {
				continue
			}
			version = strings.Replace(version, "_", ".", -1)
			if len(oldVersion) == 0 || stringutil.VersionCompare(version, oldVersion) > 0 {
				upgradeVersions = append(upgradeVersions, version)
				sql, _ := m["sql"]
				sqlMap[version] = sql
			}
		}

		// 如果没有可以升级的版本，直接返回
		if len(upgradeVersions) == 0 {
			err = this.checkData(db)
			if err != nil {
				return err
			}
			return nil
		}

		sort.Slice(upgradeVersions, func(i, j int) bool {
			return stringutil.VersionCompare(upgradeVersions[i], upgradeVersions[j]) < 0
		})

		// 执行升级的SQL
		for _, version := range upgradeVersions {
			sql := sqlMap[version]
			_, err = db.Exec(sql)
			if err != nil {
				if !this.canIgnoreError(err) {
					return errors.New("exec upgrade sql for version '" + version + "' failed: " + err.Error())

				}
			}
			err = this.updateVersion(db, version)
			if err != nil {
				return err
			}
		}

		// 检查数据
		err = this.checkData(db)
		if err != nil {
			return err
		}

		return nil
	}

	// 全新安装
	fullSQL, found := this.findFullSQL()
	if !found {
		return errors.New("not found full setup sql")
	}

	// 执行SQL
	_, err = db.Exec(fullSQL)
	if err != nil {
		return errors.New("create tables failed: " + err.Error())
	}

	// 检查数据
	err = this.checkData(db)
	if err != nil {
		return err
	}

	return nil
}

// 查找完整的SQL
func (this *SQLExecutor) findFullSQL() (sql string, found bool) {
	for _, m := range sqls.SQLVersions {
		code, _ := m["version"]
		if code == "full" {
			sql, found = m["sql"]
			return
		}
	}
	return
}

// 检查数据
func (this *SQLExecutor) checkData(db *dbs.DB) error {
	// 检查管理员
	err := this.checkAdmin(db)
	if err != nil {
		return err
	}

	// 检查管理员平台节点
	err = this.checkAdminNode(db)
	if err != nil {
		return err
	}

	// 检查集群配置
	err = this.checkCluster(db)

	// 更新版本号
	err = this.updateVersion(db, teaconst.Version)
	if err != nil {
		return err
	}

	return nil
}

// 检查管理员
func (this *SQLExecutor) checkAdmin(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeAdmins")
	if err != nil {
		return errors.New("check admin failed: " + err.Error())
	}
	defer func() {
		_ = stmt.Close()
	}()
	col, err := stmt.FindCol(0)
	if err != nil {
		return errors.New("check admin failed: " + err.Error())
	}
	count := types.Int(col)
	if count == 0 {
		_, err = db.Exec("INSERT INTO edgeAdmins (username, password, fullname, isSuper, createdAt, state) VALUES (?, ?, ?, ?, ?, ?)", "admin", stringutil.Md5("123456"), "管理员", 1, time.Now().Unix(), 1)
		if err != nil {
			return errors.New("create admin failed: " + err.Error())
		}
	}
	return nil
}

// 检查管理员平台节点
func (this *SQLExecutor) checkAdminNode(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeAPITokens WHERE role='admin'")
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	col, err := stmt.FindCol(0)
	if err != nil {
		return err
	}
	count := types.Int(col)
	if count > 0 {
		return nil
	}

	nodeId := rands.HexString(32)
	secret := rands.String(32)
	_, err = db.Exec("INSERT INTO edgeAPITokens (nodeId, secret, role) VALUES (?, ?, ?)", nodeId, secret, "admin")
	if err != nil {
		return err
	}

	return nil
}

// 检查集群配置
func (this *SQLExecutor) checkCluster(db *dbs.DB) error {
	/// 检查是否有集群数字
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeNodeClusters")
	if err != nil {
		return errors.New("query clusters failed: " + err.Error())
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return errors.New("query clusters failed: " + err.Error())
	}
	count := types.Int(col)
	if count > 0 {
		return nil
	}

	// 创建默认集群
	_, err = db.Exec("INSERT INTO edgeNodeClusters (name, useAllAPINodes, state) VALUES (?, ?, ?)", "默认集群", 1, 1)
	if err != nil {
		return err
	}

	return nil
}

// 更新版本号
func (this *SQLExecutor) updateVersion(db *dbs.DB, version string) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeVersions")
	if err != nil {
		return errors.New("query version failed: " + err.Error())
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return errors.New("query version failed: " + err.Error())
	}
	count := types.Int(col)
	if count > 0 {
		_, err = db.Exec("UPDATE edgeVersions SET version=?", version)
		if err != nil {
			return errors.New("update version failed: " + err.Error())
		}
		return nil
	}

	_, err = db.Exec("INSERT edgeVersions (version) VALUES (?)", version)
	if err != nil {
		return errors.New("create version failed: " + err.Error())
	}

	return nil
}

// 判断某个错误是否可以忽略
func (this *SQLExecutor) canIgnoreError(err error) bool {
	if err == nil {
		return true
	}

	// Error 1050: Table 'xxx' already exists
	if strings.Contains(err.Error(), "Error 1050") {
		return true
	}

	return false
}
