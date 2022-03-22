package models

import "encoding/json"

func (this *SSLCert) DecodeDNSNames() []string {
	if len(this.DnsNames) == 0 {
		return nil
	}

	var result = []string{}
	var err = json.Unmarshal(this.DnsNames, &result)
	if err != nil {
		return nil
	}

	return result
}

func (this *SSLCert) DecodeCommonNames() []string {
	if len(this.CommonNames) == 0 {
		return nil
	}

	var result = []string{}
	var err = json.Unmarshal(this.CommonNames, &result)
	if err != nil {
		return nil
	}

	return result
}
