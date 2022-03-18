package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestSSLCertDAO_ListCertsToUpdateOCSP(t *testing.T) {
	var dao = models.NewSSLCertDAO()
	certs, err := dao.ListCertsToUpdateOCSP(nil, 3600, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, cert := range certs {
		t.Log(cert.Id, cert.Name, "updatedAt:", cert.OcspUpdatedAt, timeutil.FormatTime("Y-m-d H:i:s", int64(cert.OcspUpdatedAt)), cert.OcspIsUpdated)
	}
}
