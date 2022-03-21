package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strconv"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewSSLCertExpireCheckExecutor().Start()
		})
	})
}

// SSLCertExpireCheckExecutor 证书检查任务
type SSLCertExpireCheckExecutor struct {
}

func NewSSLCertExpireCheckExecutor() *SSLCertExpireCheckExecutor {
	return &SSLCertExpireCheckExecutor{}
}

// Start 启动任务
func (this *SSLCertExpireCheckExecutor) Start() {
	seconds := int64(3600)
	ticker := time.NewTicker(time.Duration(seconds) * time.Second)
	for range ticker.C {
		err := this.loop(seconds)
		if err != nil {
			logs.Println("[ERROR][SSLCertExpireCheckExecutor]" + err.Error())
		}
	}
}

// 单次执行
func (this *SSLCertExpireCheckExecutor) loop(seconds int64) error {
	// 检查上次运行时间，防止重复运行
	settingKey := "sslCertExpiringCheckLoop"
	timestamp := time.Now().Unix()
	c, err := models.SharedSysSettingDAO.CompareInt64Setting(nil, settingKey, timestamp-seconds)
	if err != nil {
		return err
	}
	if c > 0 {
		return nil
	}

	// 记录时间
	err = models.SharedSysSettingDAO.UpdateSetting(nil, settingKey, []byte(numberutils.FormatInt64(timestamp)))
	if err != nil {
		return err
	}

	// 查找需要自动更新的证书
	// 30, 14 ... 是到期的天数
	for _, days := range []int{30, 14, 7} {
		certs, err := models.SharedSSLCertDAO.FindAllExpiringCerts(nil, days)
		if err != nil {
			return err
		}
		for _, cert := range certs {
			// 发送消息
			subject := "SSL证书\"" + cert.Name + "\"在" + strconv.Itoa(days) + "天后将到期，"
			msg := "SSL证书\"" + cert.Name + "\"（" + string(cert.DnsNames) + "）在" + strconv.Itoa(days) + "天后将到期，"

			// 是否有自动更新任务
			if cert.AcmeTaskId > 0 {
				task, err := acme.SharedACMETaskDAO.FindEnabledACMETask(nil, int64(cert.AcmeTaskId))
				if err != nil {
					return err
				}
				if task != nil {
					if task.AutoRenew == 1 {
						msg += "此证书是免费申请的证书，且已设置了自动续期，将会在到期前三天自动尝试续期。"
					} else {
						msg += "此证书是免费申请的证书，没有设置自动续期，请在到期前手动执行续期任务。"
					}
				}
			} else {
				msg += "请及时更新证书。"
			}

			err = models.SharedMessageDAO.CreateMessage(nil, int64(cert.AdminId), int64(cert.UserId), models.MessageTypeSSLCertExpiring, models.MessageLevelWarning, subject, msg, maps.Map{
				"certId":     cert.Id,
				"acmeTaskId": cert.AcmeTaskId,
			}.AsJSON())
			if err != nil {
				return err
			}

			// 设置最后通知时间
			err = models.SharedSSLCertDAO.UpdateCertNotifiedAt(nil, int64(cert.Id))
			if err != nil {
				return err
			}
		}
	}

	// 自动续期
	for _, days := range []int{3, 2, 1} {
		certs, err := models.SharedSSLCertDAO.FindAllExpiringCerts(nil, days)
		if err != nil {
			return err
		}
		for _, cert := range certs {
			// 发送消息
			subject := "SSL证书\"" + cert.Name + "\"在" + strconv.Itoa(days) + "天后将到期，"
			msg := "SSL证书\"" + cert.Name + "\"（" + string(cert.DnsNames) + "）在" + strconv.Itoa(days) + "天后将到期，"

			// 是否有自动更新任务
			if cert.AcmeTaskId > 0 {
				task, err := acme.SharedACMETaskDAO.FindEnabledACMETask(nil, int64(cert.AcmeTaskId))
				if err != nil {
					return err
				}
				if task != nil {
					if task.AutoRenew == 1 {
						isOk, errMsg, _ := acme.SharedACMETaskDAO.RunTask(nil, int64(cert.AcmeTaskId))
						if isOk {
							// 发送成功通知
							subject := "系统已成功为你自动更新了证书\"" + cert.Name + "\""
							msg = "系统已成功为你自动更新了证书\"" + cert.Name + "\"（" + string(cert.DnsNames) + "）。"
							err = models.SharedMessageDAO.CreateMessage(nil, int64(cert.AdminId), int64(cert.UserId), models.MessageTypeSSLCertACMETaskSuccess, models.MessageLevelSuccess, subject, msg, maps.Map{
								"certId":     cert.Id,
								"acmeTaskId": cert.AcmeTaskId,
							}.AsJSON())

							// 更新通知时间
							err = models.SharedSSLCertDAO.UpdateCertNotifiedAt(nil, int64(cert.Id))
							if err != nil {
								return err
							}
						} else {
							// 发送失败通知
							subject := "系统在尝试自动更新证书\"" + cert.Name + "\"时发生错误"
							msg = "系统在尝试自动更新证书\"" + cert.Name + "\"（" + string(cert.DnsNames) + "）时发生错误：" + errMsg + "。请检查系统设置并修复错误。"
							err = models.SharedMessageDAO.CreateMessage(nil, int64(cert.AdminId), int64(cert.UserId), models.MessageTypeSSLCertACMETaskFailed, models.MessageLevelError, subject, msg, maps.Map{
								"certId":     cert.Id,
								"acmeTaskId": cert.AcmeTaskId,
							}.AsJSON())

							// 更新通知时间
							err = models.SharedSSLCertDAO.UpdateCertNotifiedAt(nil, int64(cert.Id))
							if err != nil {
								return err
							}
						}

						// 中止不发送消息
						continue

					} else {
						msg += "此证书是免费申请的证书，没有设置自动续期，请在到期前手动执行续期任务。"
					}
				}
			} else {
				msg += "请及时更新证书。"
			}

			err = models.SharedMessageDAO.CreateMessage(nil, int64(cert.AdminId), int64(cert.UserId), models.MessageTypeSSLCertExpiring, models.MessageLevelWarning, subject, msg, maps.Map{
				"certId":     cert.Id,
				"acmeTaskId": cert.AcmeTaskId,
			}.AsJSON())
			if err != nil {
				return err
			}

			// 设置最后通知时间
			err = models.SharedSSLCertDAO.UpdateCertNotifiedAt(nil, int64(cert.Id))
			if err != nil {
				return err
			}
		}
	}

	// 当天过期
	for _, days := range []int{0} {
		certs, err := models.SharedSSLCertDAO.FindAllExpiringCerts(nil, days)
		if err != nil {
			return err
		}
		for _, cert := range certs {
			// 发送消息
			today := timeutil.Format("Y-m-d")
			subject := "SSL证书\"" + cert.Name + "\"在今天（" + today + "）过期"
			msg := "SSL证书\"" + cert.Name + "\"（" + string(cert.DnsNames) + "）在今天（" + today + "）过期，请及时更新证书，之后将不再重复提醒。"
			err = models.SharedMessageDAO.CreateMessage(nil, int64(cert.AdminId), int64(cert.UserId), models.MessageTypeSSLCertExpiring, models.MessageLevelWarning, subject, msg, maps.Map{
				"certId":     cert.Id,
				"acmeTaskId": cert.AcmeTaskId,
			}.AsJSON())
			if err != nil {
				return err
			}

			// 设置最后通知时间
			err = models.SharedSSLCertDAO.UpdateCertNotifiedAt(nil, int64(cert.Id))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
