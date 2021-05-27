package nameservers

import "encoding/json"

func (this *NSRecord) DecodeRouteIds() []int64 {
	routeIds := []int64{}
	if len(this.RouteIds) > 0 {
		_ = json.Unmarshal([]byte(this.RouteIds), &routeIds)
	}
	return routeIds
}
