package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// ServerHTTPFirewallDailyStatService WAF统计
type ServerHTTPFirewallDailyStatService struct {
	BaseService
}

// ComposeServerHTTPFirewallDashboard 组合Dashboard
func (this *ServerHTTPFirewallDailyStatService) ComposeServerHTTPFirewallDashboard(ctx context.Context, req *pb.ComposeServerHTTPFirewallDashboardRequest) (*pb.ComposeServerHTTPFirewallDashboardResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		if req.UserId > 0 && req.UserId != userId {
			return nil, this.PermissionError()
		}
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		userId = req.UserId
	}

	day := req.Day
	if len(day) != 8 {
		day = timeutil.Format("Ymd")
	}

	date := time.Date(types.Int(day[:4]), time.Month(types.Int(day[4:6])), types.Int(day[6:]), 0, 0, 0, 0, time.Local)
	var w = types.Int(timeutil.Format("w", date))
	if w == 0 {
		w = 7
	}
	weekFrom := timeutil.Format("Ymd", date.AddDate(0, 0, -w+1))
	weekTo := timeutil.Format("Ymd", date.AddDate(0, 0, -w+7))

	var d = types.Int(timeutil.Format("d"))
	monthFrom := timeutil.Format("Ymd", date.AddDate(0, 0, -d+1))
	monthTo := timeutil.Format("Ymd", date.AddDate(0, 1, -d))

	var tx = this.NullTx()

	countDailyLog, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, userId, req.ServerId, "log", day, day)
	if err != nil {
		return nil, err
	}

	countDailyBlock, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, userId, req.ServerId, "block", day, day)
	if err != nil {
		return nil, err
	}

	countDailyCaptcha, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, userId, req.ServerId, "captcha", day, day)
	if err != nil {
		return nil, err
	}

	countWeeklyBlock, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, userId, req.ServerId, "block", weekFrom, weekTo)
	if err != nil {
		return nil, err
	}

	countMonthlyBlock, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, userId, req.ServerId, "block", monthFrom, monthTo)
	if err != nil {
		return nil, err
	}

	resp := &pb.ComposeServerHTTPFirewallDashboardResponse{
		CountDailyLog:     countDailyLog,
		CountDailyBlock:   countDailyBlock,
		CountDailyCaptcha: countDailyCaptcha,
		CountWeeklyBlock:  countWeeklyBlock,
		CountMonthlyBlock: countMonthlyBlock,
	}

	// 规则分组
	groupStats, err := stats.SharedServerHTTPFirewallDailyStatDAO.GroupDailyCount(tx, userId, req.ServerId, monthFrom, monthTo, 0, 10)
	if err != nil {
		return nil, err
	}
	for _, stat := range groupStats {
		ruleGroupName, err := models.SharedHTTPFirewallRuleGroupDAO.FindHTTPFirewallRuleGroupName(tx, int64(stat.HttpFirewallRuleGroupId))
		if err != nil {
			return nil, err
		}
		if len(ruleGroupName) == 0 {
			continue
		}

		resp.HttpFirewallRuleGroups = append(resp.HttpFirewallRuleGroups, &pb.ComposeServerHTTPFirewallDashboardResponse_HTTPFirewallRuleGroupStat{
			HttpFirewallRuleGroup: &pb.HTTPFirewallRuleGroup{Id: int64(stat.HttpFirewallRuleGroupId), Name: ruleGroupName},
			Count:                 int64(stat.Count),
		})
	}

	// 每日趋势
	dayBefore := timeutil.Format("Ymd", date.AddDate(0, 0, -14))
	days, err := utils.RangeDays(dayBefore, day)
	if err != nil {
		return nil, err
	}
	{
		statList, err := stats.SharedServerHTTPFirewallDailyStatDAO.FindDailyStats(tx, userId, req.ServerId, "log", dayBefore, day)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Day] = int64(stat.Count)
		}
		for _, day := range days {
			resp.LogDailyStats = append(resp.LogDailyStats, &pb.ComposeServerHTTPFirewallDashboardResponse_DailyStat{Day: day, Count: m[day]})
		}
	}
	{
		statList, err := stats.SharedServerHTTPFirewallDailyStatDAO.FindDailyStats(tx, userId, req.ServerId, "block", dayBefore, day)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Day] = int64(stat.Count)
		}
		for _, day := range days {
			resp.BlockDailyStats = append(resp.BlockDailyStats, &pb.ComposeServerHTTPFirewallDashboardResponse_DailyStat{Day: day, Count: m[day]})
		}
	}
	{
		statList, err := stats.SharedServerHTTPFirewallDailyStatDAO.FindDailyStats(tx, userId, req.ServerId, "captcha", dayBefore, day)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Day] = int64(stat.Count)
		}
		for _, day := range days {
			resp.CaptchaDailyStats = append(resp.CaptchaDailyStats, &pb.ComposeServerHTTPFirewallDashboardResponse_DailyStat{Day: day, Count: m[day]})
		}
	}

	return resp, nil
}
