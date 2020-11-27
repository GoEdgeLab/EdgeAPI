package setup

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
)

type upgradeVersion struct {
	version string
	f       func(db *dbs.DB) error
}

var upgradeFuncs = []*upgradeVersion{
	{
		"0.0.3", upgradeV0_0_3,
	},
}

// 升级SQL数据
func UpgradeSQLData(db *dbs.DB) error {
	version, err := db.FindCol(0, "SELECT version FROM edgeVersions")
	if err != nil {
		return err
	}
	versionString := types.String(version)
	if len(versionString) > 0 {
		for _, f := range upgradeFuncs {
			if stringutil.VersionCompare(versionString, f.version) >= 0 {
				continue
			}
			err = f.f(db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// v0.0.3
func upgradeV0_0_3(db *dbs.DB) error {
	// 获取第一个管理员
	adminIdCol, err := db.FindCol(0, "SELECT id FROM edgeAdmins ORDER BY id ASC LIMIT 1")
	if err != nil {
		return err
	}
	adminId := types.Int64(adminIdCol)
	if adminId <= 0 {
		return errors.New("'edgeAdmins' table should not be empty")
	}

	// 升级edgeDNSProviders
	_, err = db.Exec("UPDATE edgeDNSProviders SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeDNSDomains
	_, err = db.Exec("UPDATE edgeDNSDomains SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeSSLCerts
	_, err = db.Exec("UPDATE edgeSSLCerts SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeNodeClusters
	_, err = db.Exec("UPDATE edgeNodeClusters SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeNodes
	_, err = db.Exec("UPDATE edgeNodes SET adminId=? WHERE adminId=0 AND userId=0", adminId)
	if err != nil {
		return err
	}

	// 升级edgeNodeGrants
	_, err = db.Exec("UPDATE edgeNodeGrants SET adminId=? WHERE adminId=0", adminId)
	if err != nil {
		return err
	}

	return nil
}
