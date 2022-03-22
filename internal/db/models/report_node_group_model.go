package models

// ReportNodeGroup 监控终端区域
type ReportNodeGroup struct {
	Id    uint32 `field:"id"`    // ID
	Name  string `field:"name"`  // 名称
	State uint8  `field:"state"` // 状态
	IsOn  bool   `field:"isOn"`  // 是否启用
}

type ReportNodeGroupOperator struct {
	Id    interface{} // ID
	Name  interface{} // 名称
	State interface{} // 状态
	IsOn  interface{} // 是否启用
}

func NewReportNodeGroupOperator() *ReportNodeGroupOperator {
	return &ReportNodeGroupOperator{}
}
