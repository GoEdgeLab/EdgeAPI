package installers

import (
	"bytes"
	"errors"
	"path/filepath"
)

type NodeInstaller struct {
	BaseInstaller
}

func (this *NodeInstaller) Install(dir string, params interface{}) error {
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

	// 安装助手
	env, err := this.InstallHelper(dir)
	if err != nil {
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
		templateFile := dir + "/edge-node/configs/api.template.yaml"
		configFile := dir + "/edge-node/configs/api.yaml"
		data, err := this.client.ReadFile(templateFile)
		if err != nil {
			return err
		}

		data = bytes.ReplaceAll(data, []byte("${endpoints}"), []byte(nodeParams.QuoteEndpoints()))
		data = bytes.ReplaceAll(data, []byte("${nodeId}"), []byte(nodeParams.NodeId))
		data = bytes.ReplaceAll(data, []byte("${nodeSecret}"), []byte(nodeParams.Secret))

		_, err = this.client.WriteFile(configFile, data)
		if err != nil {
			return errors.New("write 'configs/api.yaml': " + err.Error())
		}
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
