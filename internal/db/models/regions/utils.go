package regions

import "sync"

var SharedCacheLocker = sync.RWMutex{}
