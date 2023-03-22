package models

func (this *ServerDailyStat) AsUserBandwidthStat() *UserBandwidthStat {
	return &UserBandwidthStat{
		Id:                  0,
		UserId:              uint64(this.UserId),
		RegionId:            this.RegionId,
		Day:                 this.Day,
		TimeAt:              this.TimeFrom[:4],
		Bytes:               this.Bytes / 300,
		TotalBytes:          this.Bytes,
		AvgBytes:            this.Bytes / 300,
		CachedBytes:         this.CachedBytes,
		AttackBytes:         this.AttackBytes,
		CountRequests:       this.CountRequests,
		CountCachedRequests: this.CountCachedRequests,
		CountAttackRequests: this.CountAttackRequests,
	}
}

func (this *ServerDailyStat) AsServerBandwidthStat() *ServerBandwidthStat {
	return &ServerBandwidthStat{
		Id:                  0,
		UserId:              uint64(this.UserId),
		ServerId:            uint64(this.ServerId),
		RegionId:            this.RegionId,
		Day:                 this.Day,
		TimeAt:              this.TimeFrom[:4],
		Bytes:               this.Bytes / 300,
		TotalBytes:          this.Bytes,
		AvgBytes:            this.Bytes / 300,
		CachedBytes:         this.CachedBytes,
		AttackBytes:         this.AttackBytes,
		CountRequests:       this.CountRequests,
		CountCachedRequests: this.CountCachedRequests,
		CountAttackRequests: this.CountAttackRequests,
	}
}
