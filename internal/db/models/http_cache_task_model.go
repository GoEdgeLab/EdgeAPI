package models

// HTTPCacheTask 缓存相关任务
type HTTPCacheTask struct {
	Id          uint64 `field:"id"`          // ID
	UserId      uint32 `field:"userId"`      // 用户ID
	Type        string `field:"type"`        // 任务类型：purge|fetch
	KeyType     string `field:"keyType"`     // Key类型
	State       uint8  `field:"state"`       // 状态
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	DoneAt      uint64 `field:"doneAt"`      // 完成时间
	Day         string `field:"day"`         // 创建日期YYYYMMDD
	IsDone      bool   `field:"isDone"`      // 是否已完成
	IsOk        bool   `field:"isOk"`        // 是否完全成功
	IsReady     uint8  `field:"isReady"`     // 是否已准备好
	Description string `field:"description"` // 描述
}

type HTTPCacheTaskOperator struct {
	Id          interface{} // ID
	UserId      interface{} // 用户ID
	Type        interface{} // 任务类型：purge|fetch
	KeyType     interface{} // Key类型
	State       interface{} // 状态
	CreatedAt   interface{} // 创建时间
	DoneAt      interface{} // 完成时间
	Day         interface{} // 创建日期YYYYMMDD
	IsDone      interface{} // 是否已完成
	IsOk        interface{} // 是否完全成功
	IsReady     interface{} // 是否已准备好
	Description interface{} // 描述
}

func NewHTTPCacheTaskOperator() *HTTPCacheTaskOperator {
	return &HTTPCacheTaskOperator{}
}
