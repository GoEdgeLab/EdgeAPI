package models

// 并发锁
type SysLocker struct {
	Id        uint64 `field:"id"`        // ID
	Key       string `field:"key"`       // 键值
	Version   uint64 `field:"version"`   // 版本号
	TimeoutAt uint64 `field:"timeoutAt"` // 超时时间
}

type SysLockerOperator struct {
	Id        interface{} // ID
	Key       interface{} // 键值
	Version   interface{} // 版本号
	TimeoutAt interface{} // 超时时间
}

func NewSysLockerOperator() *SysLockerOperator {
	return &SysLockerOperator{}
}
