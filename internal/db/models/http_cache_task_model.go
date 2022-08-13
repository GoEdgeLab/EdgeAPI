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
	IsReady     bool   `field:"isReady"`     // 是否已准备好
	Description string `field:"description"` // 描述
}

type HTTPCacheTaskOperator struct {
	Id          any // ID
	UserId      any // 用户ID
	Type        any // 任务类型：purge|fetch
	KeyType     any // Key类型
	State       any // 状态
	CreatedAt   any // 创建时间
	DoneAt      any // 完成时间
	Day         any // 创建日期YYYYMMDD
	IsDone      any // 是否已完成
	IsOk        any // 是否完全成功
	IsReady     any // 是否已准备好
	Description any // 描述
}

func NewHTTPCacheTaskOperator() *HTTPCacheTaskOperator {
	return &HTTPCacheTaskOperator{}
}
