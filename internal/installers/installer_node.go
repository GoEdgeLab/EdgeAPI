package installers

import (
	"bytes"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"path/filepath"
	"regexp"
)

type NodeInstaller struct {
	BaseInstaller
}

func (this *NodeInstaller) Install(dir string, params interface{}, installStatus *models.NodeInstallStatus) error {
	if params == nil {
		return errors.New("'params' required for node installation")
	}
	nodeParams, ok := params.(*NodeParams)
	if !ok {
		return errors.New("'params' should be *NodeParams")
	}
	err := nodeParams.Validate()
	if err != nil {
		return errors.New("params validation: " + err.Error())
	}

	// 检查目标目录是否存在
	_, err = this.client.Stat(dir)
	if err != nil {
		err = this.client.MkdirAll(dir)
		if err != nil {
			installStatus.ErrorCode = "CREATE_ROOT_DIRECTORY_FAILED"
			return errors.New("create directory  '" + dir + "' failed: " + err.Error())
		}
	}

	// 安装助手
	env, err := this.InstallHelper(dir, nodeconfigs.NodeRoleNode)
	if err != nil {
		installStatus.ErrorCode = "INSTALL_HELPER_FAILED"
		return err
	}

	// 上传安装文件
	filePrefix := "edge-node-" + env.OS + "-" + env.Arch
	zipFile, err := this.LookupLatestInstaller(filePrefix)
	if err != nil {
		return err
	}
	if len(zipFile) == 0 {
		return errors.New("can not find installer file for " + env.OS + "/" + env.Arch)
	}
	targetZip := dir + "/" + filepath.Base(zipFile)
	err = this.client.Copy(zipFile, targetZip, 0777)
	if err != nil {
		return err
	}

	// 测试运行环境
	// 升级的节点暂时不列入测试
	if !nodeParams.IsUpgrading {
		_, stderr, err := this.client.Exec(dir + "/" + env.HelperName + " -cmd=test")
		if err != nil {
			return errors.New("test failed: " + err.Error())
		}
		if len(stderr) > 0 {
			return errors.New("test failed: " + stderr)
		}
	}

	// 如果是升级则优雅停止先前的进程
	exePath := dir + "/edge-node/bin/edge-node"
	if nodeParams.IsUpgrading {
		_, err = this.client.Stat(exePath)
		if err == nil {
			_, _, _ = this.client.Exec(exePath + " quit")
		}

		// 删除可执行文件防止冲突
		err = this.client.Remove(exePath)
		if err != nil {
			return errors.New("remove old file failed: " + err.Error())
		}
	}

	// 解压
	_, stderr, err := this.client.Exec(dir + "/" + env.HelperName + " -cmd=unzip -zip=\"" + targetZip + "\" -target=\"" + dir + "\"")
	if err != nil {
		return err
	}
	if len(stderr) > 0 {
		return errors.New("unzip installer failed: " + stderr)
	}

	// 修改配置文件
	{
		configFile := dir + "/edge-node/configs/api.yaml"

		// sudo之后我们需要修改配置目录才能写入文件
		if this.client.sudo {
			_, _, _ = this.client.Exec("chown " + this.client.User() + " " + filepath.Dir(configFile))
		}

		var data = []byte(`rpc:
  endpoints: [ ${endpoints} ]
nodeId: "${nodeId}"
secret: "${nodeSecret}"`)

		data = bytes.ReplaceAll(data, []byte("${endpoints}"), []byte(nodeParams.QuoteEndpoints()))
		data = bytes.ReplaceAll(data, []byte("${nodeId}"), []byte(nodeParams.NodeId))
		data = bytes.ReplaceAll(data, []byte("${nodeSecret}"), []byte(nodeParams.Secret))

		_, err = this.client.WriteFile(configFile, data)
		if err != nil {
			return errors.New("write '" + configFile + "': " + err.Error())
		}
	}

	// 测试
	_, stderr, err = this.client.Exec(dir + "/edge-node/bin/edge-node test")
	if err != nil {
		installStatus.ErrorCode = "TEST_FAILED"
		return errors.New("test edge node failed: " + err.Error())
	}
	if len(stderr) > 0 {
		if regexp.MustCompile(`(?i)rpc`).MatchString(stderr) {
			installStatus.ErrorCode = "RPC_TEST_FAILED"
		}

		return errors.New("test edge node failed: " + stderr)
	}

	// 启动
	_, stderr, err = this.client.Exec(dir + "/edge-node/bin/edge-node start")
	if err != nil {
		return errors.New("start edge node failed: " + err.Error())
	}

	if len(stderr) > 0 {
		return errors.New("start edge node failed: " + stderr)
	}

	return nil
}
