package installers

import (
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
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

// Login 登录SSH服务
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
	var methods = []ssh.AuthMethod{}
	if credentials.Method == "user" {
		{
			var authMethod = ssh.Password(credentials.Password)
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
	} else if credentials.Method == "privateKey" {
		var signer ssh.Signer
		var err error
		if len(credentials.Passphrase) > 0 {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(credentials.PrivateKey), []byte(credentials.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(credentials.PrivateKey))
		}
		if err != nil {
			return fmt.Errorf("parse private key: %w", err)
		}
		authMethod := ssh.PublicKeys(signer)
		methods = append(methods, authMethod)
	} else {
		return errors.New("invalid method '" + credentials.Method + "'")
	}

	// SSH客户端
	if len(credentials.Username) == 0 {
		credentials.Username = "root"
	}
	var config = &ssh.ClientConfig{
		User:            credentials.Username,
		Auth:            methods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         5 * time.Second, // TODO 后期可以设置这个超时时间
	}

	sshClient, err := ssh.Dial("tcp", configutils.QuoteIP(credentials.Host)+":"+strconv.Itoa(credentials.Port), config)
	if err != nil {
		return err
	}
	client, err := NewSSHClient(sshClient)
	if err != nil {
		return err
	}

	if credentials.Sudo {
		client.Sudo(credentials.Password)
	}

	this.client = client

	return nil
}

// Close 关闭SSH服务
func (this *BaseInstaller) Close() error {
	if this.client != nil {
		return this.client.Close()
	}

	return nil
}

// LookupLatestInstaller 查找最新的版本的文件
func (this *BaseInstaller) LookupLatestInstaller(filePrefix string) (string, error) {
	matches, err := filepath.Glob(Tea.Root + Tea.DS + "deploy" + Tea.DS + "*.zip")
	if err != nil {
		return "", err
	}

	pattern, err := regexp.Compile(filePrefix + `-v([\d.]+)\.zip`)
	if err != nil {
		return "", err
	}

	var lastVersion = ""
	var result = ""
	for _, match := range matches {
		var baseName = filepath.Base(match)
		if !pattern.MatchString(baseName) {
			continue
		}
		var m = pattern.FindStringSubmatch(baseName)
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

// InstallHelper 上传安装助手
func (this *BaseInstaller) InstallHelper(targetDir string, role nodeconfigs.NodeRole) (env *Env, err error) {
	var uname = this.uname()

	var osName string
	var archName string
	if strings.Contains(uname, "Darwin") {
		osName = "darwin"
	} else if strings.Contains(uname, "Linux") {
		osName = "linux"
	} else {
		// TODO 支持freebsd, aix ...
		return env, errors.New("installer not supported os '" + uname + "'")
	}

	if strings.Contains(uname, "aarch64") || strings.Contains(uname, "armv8") {
		archName = "arm64"
	} else if strings.Contains(uname, "aarch64_be") {
		archName = "arm64be"
	} else if strings.Contains(uname, "mips64el") {
		archName = "mips64le"
	} else if strings.Contains(uname, "mips64") {
		archName = "mips64"
	} else if strings.Contains(uname, "x86_64") {
		archName = "amd64"
	} else {
		archName = "386"
	}

	var exeName = "edge-installer-helper-" + osName + "-" + archName
	switch role {
	case nodeconfigs.NodeRoleDNS:
		exeName = "edge-installer-dns-helper-" + osName + "-" + archName
	}
	var exePath = Tea.Root + "/installers/" + exeName

	var realHelperPath = ""

	var firstCopyErr error
	for _, path := range []string{
		targetDir + "/" + exeName,
		this.client.UserHome() + "/" + exeName,
		"/tmp/" + exeName,
	} {
		err = this.client.Copy(exePath, path, 0777)
		if err != nil {
			if firstCopyErr == nil {
				firstCopyErr = err
			}
		} else {
			err = nil
			firstCopyErr = nil
			realHelperPath = path
			break
		}
	}
	if firstCopyErr != nil {
		return env, errors.New("copy '" + exeName + "' to '" + targetDir + "' failed: " + firstCopyErr.Error())
	}

	env = &Env{
		OS:         osName,
		Arch:       archName,
		HelperPath: realHelperPath,
	}
	return env, nil
}

func (this *BaseInstaller) uname() (uname string) {
	var unameRetries = 3

	for i := 0; i < unameRetries; i++ {
		for _, unameExe := range []string{"uname", "/bin/uname", "/usr/bin/uname"} {
			uname, _, _ = this.client.Exec(unameExe + " -a")
			if len(uname) > 0 {
				return
			}
		}
	}

	return "x86_64 GNU/Linux"
}
