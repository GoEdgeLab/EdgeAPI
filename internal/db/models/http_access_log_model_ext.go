package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 转换成PB对象
func (this *HTTPAccessLog) ToPB() (*pb.HTTPAccessLog, error) {
	p := &pb.HTTPAccessLog{}
	err := json.Unmarshal([]byte(this.Content), p)
	if err != nil {
		return nil, err
	}
	p.RequestId = this.RequestId
	return p, nil
}
