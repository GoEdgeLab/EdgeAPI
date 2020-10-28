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

// 执行shell命令
func (this *SSHClient) Exec(cmd string) (stdout string, stderr string, err error) {
	session, err := this.raw.NewSession()
	if err != nil {
		return "", "", err
	}
	defer func() {
		_ = session.Close()
	}()

	stdoutBuf := bytes.NewBuffer([]byte{})
	stderrBuf := bytes.NewBuffer([]byte{})
	session.Stdout = stdoutBuf
	session.Stderr = stderrBuf
	err = session.Run(cmd)
	if err != nil {
		return stdoutBuf.String(), stderrBuf.String(), err
	}
	return strings.TrimRight(stdoutBuf.String(), "\n"), stderrBuf.String(), nil
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

// 拷贝文件
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

// 获取新Session
func (this *SSHClient) NewSession() (*ssh.Session, error) {
	return this.raw.NewSession()
}

// 读取文件内容
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

// 写入文件内容
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

// 删除文件
func (this *SSHClient) Remove(path string) error {
	return this.sftp.Remove(path)
}
