package nameservers

import (
	"encoding/json"
	"github.com/iwind/TeaGo/types"
)

func (this *NSRecord) DecodeRouteIds() []string {
	var routeIds = []string{}
	if len(this.RouteIds) > 0 {
		err := json.Unmarshal(this.RouteIds, &routeIds)
		if err != nil {
			// 检查是否有旧的数据
			var oldRouteIds = []int64{}
			err = json.Unmarshal(this.RouteIds, &oldRouteIds)
			if err != nil {
				return []string{}
			}
			routeIds = []string{}
			for _, routeId := range oldRouteIds {
				routeIds = append(routeIds, "id:"+types.String(routeId))
			}
		}
	}
	return routeIds
}
