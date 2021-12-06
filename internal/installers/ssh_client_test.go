// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package installers

import (
	"golang.org/x/crypto/ssh"
	"net"
	"testing"
	"time"
)

func testSSHClient(t *testing.T, username string, password string) *SSHClient {
	methods := []ssh.AuthMethod{}
	{
		authMethod := ssh.Password(password)
		methods = append(methods, authMethod)
	}

	{
		authMethod := ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			if len(questions) == 0 {
				return []string{}, nil
			}
			return []string{password}, nil
		})
		methods = append(methods, authMethod)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: methods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 5 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", "192.168.2.31:22", config)
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewSSHClient(sshClient)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestSSHClient_Home(t *testing.T) {
	var client = testSSHClient(t, "root", "123456")
	t.Log(client.UserHome())
}

func TestSSHClient_Exec(t *testing.T) {
	var client = testSSHClient(t, "liuxiangchao", "123456")
	stdout, stderr, err := client.Exec("echo 'Hello'")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("stdout:", stdout, "stderr:", stderr)
}

func TestSSHClient_SudoExec(t *testing.T) {
	var client = testSSHClient(t, "liuxiangchao", "123456")
	client.Sudo("123456")
	stdout, stderr, err := client.Exec("echo 'Hello'")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("stdout:", stdout, "stderr:", stderr)
}

func TestSSHClient_SudoExec2(t *testing.T) {
	var client = testSSHClient(t, "liuxiangchao", "123456")
	client.Sudo("123456")
	stdout, stderr, err := client.Exec("/home/liuxiangchao/edge-node/edge-node/bin/edge-node start")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("stdout:", stdout, "stderr:", stderr)
}
