package setup

import (
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

type SQLRecord struct {
	Id           int64             `json:"id"`
	Values       map[string]string `json:"values"`
	UniqueFields []string          `json:"uniqueFields"`
	ExceptFields []string          `json:"exceptFields"`
}

func (this *SQLRecord) ValuesEquals(values maps.Map) bool {
	for k, v := range values {
		// 跳过ID
		if k == "id" {
			continue
		}

		// 需要排除的字段
		if lists.ContainsString(this.ExceptFields, k) {
			continue
		}

		var vString = types.String(v)
		if this.Values[k] != vString {
			return false
		}
	}
	return true
}
