package installers

import (
	"bytes"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"strings"
)

type SSHClient struct {
	raw  *ssh.Client
	sftp *sftp.Client

	sudo         bool
	sudoPassword string
}

func NewSSHClient(raw *ssh.Client) (*SSHClient, error) {
	c := &SSHClient{
		raw: raw,
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		_ = c.Close()
		return nil, err
	}
	c.sftp = sftpClient

	return c, nil
}

// Sudo 设置使用Sudo
func (this *SSHClient) Sudo(password string) {
	this.sudo = true
	this.sudoPassword = password
}

// Exec 执行shell命令
func (this *SSHClient) Exec(cmd string) (stdout string, stderr string, err error) {
	if this.raw.User() != "root" && this.sudo {
		return this.execSudo(cmd, this.sudoPassword)
	}

	session, err := this.raw.NewSession()
	if err != nil {
		return "", "", err
	}
	defer func() {
		_ = session.Close()
	}()

	var stdoutBuf = &bytes.Buffer{}
	var stderrBuf = &bytes.Buffer{}
	session.Stdout = stdoutBuf
	session.Stderr = stderrBuf
	err = session.Run(cmd)
	if err != nil {
		return stdoutBuf.String(), stderrBuf.String(), err
	}
	return strings.TrimRight(stdoutBuf.String(), "\n"), stderrBuf.String(), nil
}

// execSudo 使用sudo执行shell命令
func (this *SSHClient) execSudo(cmd string, password string) (stdout string, stderr string, err error) {
	session, err := this.raw.NewSession()
	if err != nil {
		return "", "", err
	}
	defer func() {
		_ = session.Close()
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0, // disable echo
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return "", "", err
	}

	var stderrBuf = &bytes.Buffer{}
	session.Stderr = stderrBuf

	pipeIn, err := session.StdinPipe()
	if err != nil {
		return "", "", err
	}

	pipeOut, err := session.StdoutPipe()
	if err != nil {
		return "", "", err
	}

	var resultErr error
	var stdoutBuf = bytes.NewBuffer([]byte{})

	go func() {
		var buf = make([]byte, 512)
		for {
			n, err := pipeOut.Read(buf)
			if n > 0 {
				if strings.Contains(string(buf[:n]), "[sudo] password for") {
					_, err = pipeIn.Write([]byte(password + "\n"))
					if err != nil {
						resultErr = err
						return
					}
					continue
				}
				stdoutBuf.Write(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	err = session.Run("sudo " + cmd)

	stdout = strings.TrimSpace(stdoutBuf.String())
	stderr = strings.TrimSpace(stderrBuf.String())

	if err != nil {
		return stdout, stderr, err
	}

	if resultErr != nil {
		return stdout, stderr, resultErr
	}
	return stdout, stderr, nil
}

func (this *SSHClient) Listen(network string, addr string) (net.Listener, error) {
	return this.raw.Listen(network, addr)
}

func (this *SSHClient) Dial(network string, addr string) (net.Conn, error) {
	return this.raw.Dial(network, addr)
}

func (this *SSHClient) Close() error {
	if this.sftp != nil {
		_ = this.sftp.Close()
	}
	return this.raw.Close()
}

func (this *SSHClient) OpenFile(path string, flags int) (*sftp.File, error) {
	return this.sftp.OpenFile(path, flags)
}

func (this *SSHClient) Stat(path string) (os.FileInfo, error) {
	return this.sftp.Stat(path)
}

func (this *SSHClient) Mkdir(path string) error {
	return this.sftp.Mkdir(path)
}

func (this *SSHClient) MkdirAll(path string) error {
	return this.sftp.MkdirAll(path)
}

func (this *SSHClient) Chmod(path string, mode os.FileMode) error {
	return this.sftp.Chmod(path, mode)
}

// Copy 拷贝文件
func (this *SSHClient) Copy(localPath string, remotePath string, mode os.FileMode) error {
	localFp, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = localFp.Close()
	}()
	remoteFp, err := this.sftp.OpenFile(remotePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
	if err != nil {
		return err
	}
	defer func() {
		_ = remoteFp.Close()
	}()
	_, err = io.Copy(remoteFp, localFp)
	if err != nil {
		return err
	}

	return this.Chmod(remotePath, mode)
}

// NewSession 获取新Session
func (this *SSHClient) NewSession() (*ssh.Session, error) {
	return this.raw.NewSession()
}

// ReadFile 读取文件内容
func (this *SSHClient) ReadFile(path string) ([]byte, error) {
	fp, err := this.sftp.OpenFile(path, 0444)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = fp.Close()
	}()

	buffer := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buffer, fp)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// WriteFile 写入文件内容
func (this *SSHClient) WriteFile(path string, data []byte) (n int, err error) {
	fp, err := this.sftp.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = fp.Close()
	}()

	n, err = fp.Write(data)
	return
}

// Remove 删除文件
func (this *SSHClient) Remove(path string) error {
	return this.sftp.Remove(path)
}

// User 用户名
func (this *SSHClient) User() string {
	return this.raw.User()
}

// UserHome 用户地址
func (this *SSHClient) UserHome() string {
	homeStdout, _, err := this.Exec("echo $HOME")
	if err != nil {
		return this.defaultUserHome()
	}

	var home = strings.TrimSpace(homeStdout)
	if len(home) > 0 {
		return home
	}

	return this.defaultUserHome()
}

func (this *SSHClient) defaultUserHome() string {
	var user = this.raw.User()
	if user == "root" {
		return "/root"
	}
	return "/home/" + user
}
