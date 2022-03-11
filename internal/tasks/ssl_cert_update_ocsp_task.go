// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
	"golang.org/x/crypto/ocsp"
	"io/ioutil"
	"net/http"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewSSLCertUpdateOCSPTask().Start()
		})
	})
}

type SSLCertUpdateOCSPTask struct {
	ticker *time.Ticker
}

func NewSSLCertUpdateOCSPTask() *SSLCertUpdateOCSPTask {
	return &SSLCertUpdateOCSPTask{
		ticker: time.NewTicker(1 * time.Minute),
	}
}

func (this *SSLCertUpdateOCSPTask) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			remotelogs.Error("SSLCertUpdateOCSPTask", err.Error())
		}
	}
}

func (this *SSLCertUpdateOCSPTask) Loop() error {
	ok, err := models.SharedSysLockerDAO.Lock(nil, "ssl_cert_update_ocsp_task", 60-1) // 假设执行时间为1秒
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var tx *dbs.Tx
	// TODO 将来可以设置单次任务条数
	var size int64 = 20
	certs, err := models.SharedSSLCertDAO.ListCertsToUpdateOCSP(tx, size)
	if err != nil {
		return errors.New("list certs failed: " + err.Error())
	}
	if len(certs) == 0 {
		return nil
	}

	for _, cert := range certs {
		ocspData, err := this.UpdateCertOCSP(cert)
		var errString = ""
		if err != nil {
			errString = err.Error()

			remotelogs.Warn("SSLCertUpdateOCSPTask", "update ocsp failed: "+errString)
		}
		err = models.SharedSSLCertDAO.UpdateCertOSCP(tx, int64(cert.Id), ocspData, errString)
		if err != nil {
			return errors.New("update ocsp failed: " + err.Error())
		}
	}

	return nil
}

func (this *SSLCertUpdateOCSPTask) UpdateCertOCSP(certOne *models.SSLCert) (ocspData []byte, err error) {
	if certOne.IsCA == 1 || len(certOne.CertData) == 0 || len(certOne.KeyData) == 0 {
		return
	}

	keyPair, err := tls.X509KeyPair([]byte(certOne.CertData), []byte(certOne.KeyData))
	if err != nil {
		return nil, errors.New("parse certificate failed: " + err.Error())
	}
	if len(keyPair.Certificate) == 0 {
		return nil, nil
	}

	var certData = keyPair.Certificate[0]
	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, errors.New("parse certificate block failed: " + err.Error())
	}

	// 是否已过期
	var now = time.Now()
	if cert.NotBefore.After(now) || cert.NotAfter.Before(now) {
		return nil, nil
	}

	if len(cert.IssuingCertificateURL) == 0 || len(cert.OCSPServer) == 0 {
		return nil, nil
	}

	if len(cert.DNSNames) == 0 {
		return nil, nil
	}

	var issuerURL = cert.IssuingCertificateURL[0]
	var ocspServerURL = cert.OCSPServer[0]

	var httpClient = utils.SharedHttpClient(5 * time.Second)
	issuerReq, err := http.NewRequest(http.MethodGet, issuerURL, nil)
	if err != nil {
		return nil, errors.New("request issuer certificate failed: " + err.Error())
	}
	issuerReq.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)
	issuerResp, err := httpClient.Do(issuerReq)
	if err != nil {
		return nil, errors.New("request issuer certificate failed: '" + issuerURL + "': " + err.Error())
	}
	defer func() {
		_ = issuerResp.Body.Close()
	}()

	issuerData, err := ioutil.ReadAll(issuerResp.Body)
	if err != nil {
		return nil, errors.New("read issuer certificate failed: '" + issuerURL + "': " + err.Error())
	}
	issuerCert, err := x509.ParseCertificate(issuerData)
	if err != nil {
		return nil, errors.New("parse issuer certificate failed: '" + issuerURL + "': " + err.Error())
	}

	buf, err := ocsp.CreateRequest(cert, issuerCert, &ocsp.RequestOptions{
		Hash: crypto.SHA1,
	})
	if err != nil {
		return nil, errors.New("create ocsp request failed: " + err.Error())
	}
	ocspReq, err := http.NewRequest(http.MethodPost, ocspServerURL, bytes.NewBuffer(buf))
	if err != nil {
		return nil, errors.New("request ocsp failed: " + err.Error())
	}
	ocspReq.Header.Set("Content-Type", "application/ocsp-request")
	ocspReq.Header.Set("Accept", "application/ocsp-response")

	ocspResp, err := httpClient.Do(ocspReq)
	if err != nil {
		return nil, errors.New("request ocsp failed: '" + ocspServerURL + "': " + err.Error())
	}

	defer func() {
		_ = ocspResp.Body.Close()
	}()

	respData, err := ioutil.ReadAll(ocspResp.Body)
	if err != nil {
		return nil, errors.New("read ocsp failed: '" + ocspServerURL + "': " + err.Error())
	}

	ocspResult, err := ocsp.ParseResponse(respData, issuerCert)
	if err != nil {
		return nil, errors.New("decode ocsp failed: " + err.Error())
	}

	// 只返回Good的ocsp
	if ocspResult.Status == ocsp.Good {
		return respData, nil
	}
	return nil, nil
}
