package acme

import (
	"crypto/x509"
	"encoding/base64"
)

func ParsePrivateKeyFromBase64(base64String string) (interface{}, error) {
	data, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS8PrivateKey(data)
}
