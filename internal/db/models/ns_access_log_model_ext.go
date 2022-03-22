package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ToPB 转换成PB对象
func (this *NSAccessLog) ToPB() (*pb.NSAccessLog, error) {
	p := &pb.NSAccessLog{}
	err := json.Unmarshal(this.Content, p)
	if err != nil {
		return nil, err
	}
	p.RequestId = this.RequestId
	return p, nil
}
