package clients

// NSRouteCode NS线路代号
func (this *ClientAgent) NSRouteCode() string {
	return "agent:" + this.Code
}
