package models

func (this *ServerDailyStat) AsUserBandwidthStat() *UserBandwidthStat {
	var timeAt = "0000"
	if len(this.TimeFrom) >= 4 {
		timeAt = this.TimeFrom[:4]
	} else if len(this.Hour) > 8 {
		timeAt = this.Hour[8:] + "00"
	}
	return &UserBandwidthStat{
		Id:                  0,
		UserId:              uint64(this.UserId),
		RegionId:            this.RegionId,
		Day:                 this.Day,
		TimeAt:              timeAt,
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
	var timeAt = "0000"
	if len(this.TimeFrom) >= 4 {
		timeAt = this.TimeFrom[:4]
	} else if len(this.Hour) > 8 {
		timeAt = this.Hour[8:] + "00"
	}
	return &ServerBandwidthStat{
		Id:                  0,
		UserId:              uint64(this.UserId),
		ServerId:            uint64(this.ServerId),
		RegionId:            this.RegionId,
		Day:                 this.Day,
		TimeAt:              timeAt,
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
