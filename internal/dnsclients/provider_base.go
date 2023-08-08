package dnsclients

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/types"
)

type BaseProvider struct{}

// WrapError 封装解析相关错误
func (this *BaseProvider) WrapError(err error, domain string, record *dnstypes.Record) error {
	if err == nil {
		return nil
	}

	if record == nil {
		return err
	}

	var fullname string
	if len(record.Name) == 0 {
		fullname = domain
	} else {
		fullname = record.Name + "." + domain
	}
	return errors.New("record operation failed: '" + fullname + " " + record.Type + " " + record.Value + " " + types.String(record.TTL) + "': " + err.Error())
}
