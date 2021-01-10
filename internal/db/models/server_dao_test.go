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
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := SharedServerDAO.ComposeServerConfig(tx, 1)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config, t)
}

func TestServerDAO_ComposeServerConfig_AliasServerNames(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := SharedServerDAO.ComposeServerConfig(tx, 14)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config.AliasServerNames, t)
}

func TestServerDAO_UpdateServerConfig(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := SharedServerDAO.ComposeServerConfig(tx, 1)
	if err != nil {
		t.Fatal(err)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = SharedServerDAO.UpdateServerConfig(tx, 1, configJSON, false)
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
	var tx *dbs.Tx
	dnsName, err := SharedServerDAO.genDNSName(tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("dnsName:", dnsName)
}

func TestServerDAO_FindAllServerDNSNamesWithDNSDomainId(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	dnsNames, err := SharedServerDAO.FindAllServerDNSNamesWithDNSDomainId(tx, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("dnsNames:", dnsNames)
}
