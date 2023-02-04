package models

import "time"

func (this *LoginSession) IsAvailable() bool {
	return this.ExpiresAt == 0 || int64(this.ExpiresAt) > time.Now().Unix()
}
