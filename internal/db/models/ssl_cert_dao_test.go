package models_test

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/assert"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestSSLCertDAO_ListCertsToUpdateOCSP(t *testing.T) {
	var dao = models.NewSSLCertDAO()
	certs, err := dao.ListCertsToUpdateOCSP(nil, 3, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, cert := range certs {
		t.Log(cert.Id, cert.Name, "updatedAt:", cert.OcspUpdatedAt, timeutil.FormatTime("Y-m-d H:i:s", int64(cert.OcspUpdatedAt)), cert.OcspIsUpdated, string(cert.Ocsp))
	}
}

func TestSSLCertDAO_Update_Blob(t *testing.T) {
	var a = assert.NewAssertion(t)

	var certId = 2

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("ocsp", 123).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).Ocsp) == "123")
		}
	}

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("ocsp", 456).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).Ocsp) == "456")
		}
	}

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("ocsp", []byte("789")).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).Ocsp) == "789")
		}
	}

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("ocsp", []byte("")).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).Ocsp) == "")
		}
	}

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("ocsp", nil).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).Ocsp) == "")
		}
	}

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("ocsp", []byte("1.2.3")).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).Ocsp) == "1.2.3")
		}
	}
}

func TestSSLCertDAO_Update_JSON(t *testing.T) {
	var a = assert.NewAssertion(t)

	var certId = 2

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("commonNames", []byte("null")).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).CommonNames) == "null")
		}
	}

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("commonNames", dbs.JSON(`["a","b"]`)).
			UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}

		{
			one, _ := dao.Query(nil).Pk(certId).Find()
			if one != nil {
				a.IsTrue(string(one.(*models.SSLCert).CommonNames) == `["a", "b"]`)
			}
		}
		{
			commonNames, _ := dao.Query(nil).Pk(certId).Result("commonNames").FindBytesCol()
			t.Log("commonNames:", commonNames)
			a.IsTrue(string(commonNames) == `["a", "b"]`)
		}
	}

	{
		var op = models.NewSSLCertOperator()
		op.Id = certId
		op.CommonNames = dbs.JSON(`["a", "b"]`)

		var dao = models.NewSSLCertDAO()
		err := dao.Save(nil, op)
		if err != nil {
			t.Fatal(err)
		}

		{
			commonNames, _ := dao.Query(nil).Pk(certId).Result("commonNames").FindBytesCol()
			t.Log("commonNames:", commonNames)
			a.IsTrue(string(commonNames) == `["a", "b"]`)
		}
	}

	{
		var op = models.NewSSLCertOperator()
		op.Id = certId
		op.CommonNames = []byte(`["a", "b"]`)

		var dao = models.NewSSLCertDAO()
		err := dao.Save(nil, op)
		if err != nil {
			t.Fatal(err)
		}

		{
			commonNames, _ := dao.Query(nil).Pk(certId).Result("commonNames").FindBytesCol()
			t.Log("commonNames:", commonNames)
			a.IsTrue(string(commonNames) == `["a", "b"]`)
		}
	}

	{
		var dao = models.NewSSLCertDAO()
		err := dao.Query(nil).
			Pk(certId).
			Set("commonNames", []byte("")).
			UpdateQuickly()
		a.IsTrue(err != nil)
		if err != nil {
			a.Log("expected has error:", err.Error())
		}

		one, _ := dao.Query(nil).Pk(certId).Find()
		if one != nil {
			a.IsTrue(string(one.(*models.SSLCert).CommonNames) == `["a", "b"]`)
		}
	}

	{
		var commonNames = []string{}
		err := json.Unmarshal([]byte("null"), &commonNames)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(commonNames)
	}

	{
		var cert = &models.SSLCert{}
		err := json.Unmarshal([]byte("null"), &cert)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(cert)
	}
}
