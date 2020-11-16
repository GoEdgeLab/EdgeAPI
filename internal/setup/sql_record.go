package setup

import (
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

type SQLRecord struct {
	Id           int64             `json:"id"`
	Values       map[string]string `json:"values"`
	UniqueFields []string          `json:"uniqueFields"`
}

func (this *SQLRecord) ValuesEquals(values maps.Map) bool {
	for k, v := range values {
		if k == "id" {
			continue
		}
		vString := types.String(v)
		if this.Values[k] != vString {
			return false
		}
	}
	return true
}
