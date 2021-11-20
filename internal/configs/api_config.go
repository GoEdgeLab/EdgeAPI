package configs

import (
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"io/ioutil"
	"os"
	"path/filepath"
)

var sharedAPIConfig *APIConfig = nil
var PaddingId string

// APIConfig API节点配置
type APIConfig struct {
	NodeId string `yaml:"nodeId" json:"nodeId"`
	Secret string `yaml:"secret" json:"secret"`

	numberId int64 // 数字ID
}

// SharedAPIConfig 获取共享配置
func SharedAPIConfig() (*APIConfig, error) {
	sharedLocker.Lock()
	defer sharedLocker.Unlock()

	if sharedAPIConfig != nil {
		return sharedAPIConfig, nil
	}

	// 候选文件
	localFile := Tea.ConfigFile("api.yaml")
	isFromLocal := false
	paths := []string{localFile}
	homeDir, homeErr := os.UserHomeDir()
	if homeErr == nil {
		paths = append(paths, homeDir+"/."+teaconst.ProcessName+"/api.yaml")
	}
	paths = append(paths, "/etc/"+teaconst.ProcessName+"/api.yaml")

	// 依次检查文件
	var data []byte
	var err error
	for _, path := range paths {
		data, err = ioutil.ReadFile(path)
		if err == nil {
			if path == localFile {
				isFromLocal = true
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// 解析内容
	config := &APIConfig{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	if !isFromLocal {
		// 恢复文件
		_ = ioutil.WriteFile(localFile, data, 0666)
	}

	// 恢复数据库文件
	{
		dbConfigFile := Tea.ConfigFile("db.yaml")
		_, err := os.Stat(dbConfigFile)
		if err != nil {
			paths := []string{}
			homeDir, homeErr := os.UserHomeDir()
			if homeErr == nil {
				paths = append(paths, homeDir+"/."+teaconst.ProcessName+"/db.yaml")
			}
			paths = append(paths, "/etc/"+teaconst.ProcessName+"/db.yaml")
			for _, path := range paths {
				_, err := os.Stat(path)
				if err == nil {
					data, err := ioutil.ReadFile(path)
					if err == nil {
						_ = ioutil.WriteFile(dbConfigFile, data, 0666)
						break
					}
				}
			}
		}
	}

	sharedAPIConfig = config
	return config, nil
}

// SetNumberId 设置数字ID
func (this *APIConfig) SetNumberId(numberId int64) {
	this.numberId = numberId
	teaconst.NodeId = numberId
	PaddingId = fmt.Sprintf("%08d", numberId)
}

// NumberId 获取数字ID
func (this *APIConfig) NumberId() int64 {
	return this.numberId
}

// WriteFile 保存到文件
func (this *APIConfig) WriteFile(path string) error {
	data, err := yaml.Marshal(this)
	if err != nil {
		return err
	}

	// 生成备份文件
	filename := filepath.Base(path)
	homeDir, _ := os.UserHomeDir()
	backupDirs := []string{"/etc/edge-api"}
	if len(homeDir) > 0 {
		backupDirs = append(backupDirs, homeDir+"/.edge-api")
	}
	for _, backupDir := range backupDirs {
		stat, err := os.Stat(backupDir)
		if err == nil && stat.IsDir() {
			_ = ioutil.WriteFile(backupDir+"/"+filename, data, 0666)
		} else if err != nil && os.IsNotExist(err) {
			err = os.Mkdir(backupDir, 0777)
			if err == nil {
				_ = ioutil.WriteFile(backupDir+"/"+filename, data, 0666)
			}
		}
	}

	return ioutil.WriteFile(path, data, 0666)
}
