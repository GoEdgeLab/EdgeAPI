package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/logs"
)

// DecodePrimaryOrigins 解析主要源站
func (this *ReverseProxy) DecodePrimaryOrigins() []*serverconfigs.OriginRef {
	var refs = []*serverconfigs.OriginRef{}
	if IsNotNull(this.PrimaryOrigins) {
		err := json.Unmarshal(this.PrimaryOrigins, &refs)
		if err != nil {
			logs.Error(err)
		}
	}
	return refs
}

// DecodeBackupOrigins 解析备用源站
func (this *ReverseProxy) DecodeBackupOrigins() []*serverconfigs.OriginRef {
	var refs = []*serverconfigs.OriginRef{}
	if IsNotNull(this.BackupOrigins) {
		err := json.Unmarshal(this.BackupOrigins, &refs)
		if err != nil {
			logs.Error(err)
		}
	}
	return refs
}
