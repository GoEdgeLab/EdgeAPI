package models

import (
	"crypto/md5"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestServerDAO_ComposeServerConfig(t *testing.T) {
	config, err := SharedServerDAO.ComposeServerConfig(1)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config, t)
}

func TestServerDAO_UpdateServerConfig(t *testing.T) {
	config, err := SharedServerDAO.ComposeServerConfig(1)
	if err != nil {
		t.Fatal(err)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = SharedServerDAO.UpdateServerConfig(1, configJSON, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNewServerDAO_md5(t *testing.T) {
	m := md5.New()
	_, err := m.Write([]byte("123456"))
	if err != nil {
		t.Fatal(err)
	}
	h := m.Sum(nil)
	t.Logf("%x", h)
}

func TestServerDAO_genDNSName(t *testing.T) {
	dbs.NotifyReady()
	dnsName, err := SharedServerDAO.genDNSName()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("dnsName:", dnsName)
}
