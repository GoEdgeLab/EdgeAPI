package installers

import (
	"errors"
	"github.com/iwind/TeaGo/Tea"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"golang.org/x/crypto/ssh"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BaseInstaller struct {
	client *SSHClient
}

// 登录SSH服务
func (this *BaseInstaller) Login(credentials *Credentials) error {
	var hostKeyCallback ssh.HostKeyCallback = nil

	// 检查参数
	if len(credentials.Host) == 0 {
		return errors.New("'host' should not be empty")
	}
	if credentials.Port <= 0 {
		return errors.New("'port' should be greater than 0")
	}
	if len(credentials.Password) == 0 && len(credentials.PrivateKey) == 0 {
		return errors.New("require user 'password' or 'privateKey'")
	}

	// 不使用known_hosts
	if hostKeyCallback == nil {
		hostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}
	}

	// 认证
	methods := []ssh.AuthMethod{}
	if len(credentials.Password) > 0 {
		{
			authMethod := ssh.Password(credentials.Password)
			methods = append(methods, authMethod)
		}

		{
			authMethod := ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				if len(questions) == 0 {
					return []string{}, nil
				}
				return []string{credentials.Password}, nil
			})
			methods = append(methods, authMethod)
		}
	} else {
		signer, err := ssh.ParsePrivateKey([]byte(credentials.PrivateKey))
		if err != nil {
			return errors.New("parse private key: " + err.Error())
		}
		authMethod := ssh.PublicKeys(signer)
		methods = append(methods, authMethod)
	}

	// SSH客户端
	config := &ssh.ClientConfig{
		User:            credentials.Username,
		Auth:            methods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         5 * time.Second, // TODO 后期可以设置这个超时时间
	}

	sshClient, err := ssh.Dial("tcp", credentials.Host+":"+strconv.Itoa(credentials.Port), config)
	if err != nil {
		return err
	}
	client, err := NewSSHClient(sshClient)
	if err != nil {
		return err
	}
	this.client = client
	return nil
}

// 关闭SSH服务
func (this *BaseInstaller) Close() error {
	if this.client != nil {
		return this.client.Close()
	}

	return nil
}

// 查找最新的版本的文件
func (this *BaseInstaller) LookupLatestInstaller(filePrefix string) (string, error) {
	matches, err := filepath.Glob(Tea.Root + Tea.DS + "deploy" + Tea.DS + "*.zip")
	if err != nil {
		return "", err
	}

	pattern, err := regexp.Compile(filePrefix + `-v([\d.]+)\.zip`)
	if err != nil {
		return "", err
	}

	lastVersion := ""
	result := ""
	for _, match := range matches {
		baseName := filepath.Base(match)
		if !pattern.MatchString(baseName) {
			continue
		}
		m := pattern.FindStringSubmatch(baseName)
		if len(m) < 2 {
			continue
		}
		version := m[1]
		if len(lastVersion) == 0 || stringutil.VersionCompare(version, lastVersion) > 0 {
			lastVersion = version
			result = match
		}
	}
	return result, nil
}

// 上传安装助手
func (this *BaseInstaller) InstallHelper(targetDir string) (env *Env, err error) {
	uname, _, err := this.client.Exec("uname -a")
	if err != nil {
		return env, err
	}

	osName := ""
	archName := ""
	if strings.Index(uname, "Darwin") > 0 {
		osName = "darwin"
	} else if strings.Index(uname, "Linux") >= 0 {
		osName = "linux"
	} else {
		// TODO 支持freebsd, aix ...
		return env, errors.New("installer not supported os '" + uname + "'")
	}

	if strings.Index(uname, "x86_64") > 0 {
		archName = "amd64"
	} else {
		// TODO 支持ARM和MIPS等架构
		archName = "386"
	}

	exeName := "installer-helper-" + osName + "-" + archName
	exePath := Tea.Root + "/installers/" + exeName

	err = this.client.Copy(exePath, targetDir+"/"+exeName, 0777)
	if err != nil {
		return env, err
	}

	env = &Env{
		OS:         osName,
		Arch:       archName,
		HelperName: exeName,
	}
	return env, nil
}
