package extract

import (
	"math"
	"time"
)

// BuildPlanAnalysis analyzes cost savings per plan period and current billing cycle.
func BuildPlanAnalysis(planHistory []PlanConfig, dailyCosts []map[string]any, sessionList []SessionOutput) PlanAnalysis {
	today := time.Now().UTC().Format("2006-01-02")

	var periods []PlanPeriod
	for _, ph := range planHistory {
		start := ph.Start
		end := today
		if ph.End != nil {
			end = *ph.End
		}

		// Sum API costs in this period
		apiCost := 0.0
		for _, dc := range dailyCosts {
			d, _ := dc["date"].(string)
			if d >= start && d <= end {
				if t, ok := dc["total"].(float64); ok {
					apiCost += t
				}
			}
		}

		// Count sessions and messages in period
		sessionCount := 0
		messageCount := 0
		daysActive := make(map[string]bool)
		for _, s := range sessionList {
			if s.Date >= start && s.Date <= end {
				sessionCount++
				messageCount += s.Messages
				daysActive[s.Date] = true
			}
		}

		startDt, _ := time.Parse("2006-01-02", start)
		endDt, _ := time.Parse("2006-01-02", end)
		totalDays := int(endDt.Sub(startDt).Hours()/24) + 1

		savings := apiCost - ph.CostUSD
		roiFactor := 0.0
		if ph.CostUSD > 0 {
			roiFactor = math.Round(apiCost/ph.CostUSD*10) / 10
		}
		costPerDay := 0.0
		if totalDays > 0 {
			costPerDay = math.Round(apiCost/float64(totalDays)*100) / 100
		}

		periods = append(periods, PlanPeriod{
			Plan:        ph.Plan,
			Start:       start,
			End:         end,
			TotalDays:   totalDays,
			DaysActive:  len(daysActive),
			PlanCostEUR: ph.CostEUR,
			PlanCostUSD: ph.CostUSD,
			APICost:     math.Round(apiCost*100) / 100,
			Savings:     math.Round(savings*100) / 100,
			ROIFactor:   roiFactor,
			Sessions:    sessionCount,
			Messages:    messageCount,
			CostPerDay:  costPerDay,
		})
	}

	// Current billing period
	var currentBilling *CurrentBilling
	if len(planHistory) > 0 {
		currentPlan := planHistory[len(planHistory)-1]
		billingDay := currentPlan.BillingDay
		if billingDay == 0 {
			billingDay = 1
		}

		todayDt := time.Now().UTC()

		// Clamp billing day to last day of month (fix Python bug)
		billingStart := clampedDate(todayDt.Year(), todayDt.Month(), billingDay)
		if todayDt.Day() < billingDay {
			// Previous month
			prevMonth := todayDt.AddDate(0, -1, 0)
			billingStart = clampedDate(prevMonth.Year(), prevMonth.Month(), billingDay)
		}

		// Next billing date
		nextMonth := billingStart.AddDate(0, 1, 0)
		billingEnd := clampedDate(nextMonth.Year(), nextMonth.Month(), billingDay)

		billingStartStr := billingStart.Format("2006-01-02")
		billingEndStr := billingEnd.Format("2006-01-02")

		currentAPICost := 0.0
		for _, dc := range dailyCosts {
			d, _ := dc["date"].(string)
			if d >= billingStartStr && d <= today {
				if t, ok := dc["total"].(float64); ok {
					currentAPICost += t
				}
			}
		}

		daysElapsed := int(todayDt.Sub(billingStart).Hours()/24) + 1
		daysTotal := int(billingEnd.Sub(billingStart).Hours() / 24)
		daysRemaining := daysTotal - daysElapsed
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		projectedCost := 0.0
		if daysElapsed > 0 {
			projectedCost = currentAPICost / float64(daysElapsed) * float64(daysTotal)
		}

		var currentSessions int
		var currentMessages int
		for _, s := range sessionList {
			if s.Date >= billingStartStr && s.Date <= today {
				currentSessions++
				currentMessages += s.Messages
			}
		}

		roiFactor := 0.0
		if currentPlan.CostUSD > 0 {
			roiFactor = math.Round(currentAPICost/currentPlan.CostUSD*10) / 10
		}
		costPerDay := 0.0
		if daysElapsed > 0 {
			costPerDay = math.Round(currentAPICost/float64(daysElapsed)*100) / 100
		}

		currentBilling = &CurrentBilling{
			Plan:          currentPlan.Plan,
			PeriodStart:   billingStartStr,
			PeriodEnd:     billingEndStr,
			DaysElapsed:   daysElapsed,
			DaysTotal:     daysTotal,
			DaysRemaining: daysRemaining,
			PlanCostEUR:   currentPlan.CostEUR,
			PlanCostUSD:   currentPlan.CostUSD,
			APICost:       math.Round(currentAPICost*100) / 100,
			ProjectedCost: math.Round(projectedCost*100) / 100,
			Savings:       math.Round((currentAPICost-currentPlan.CostUSD)*100) / 100,
			ROIFactor:     roiFactor,
			Sessions:      currentSessions,
			Messages:      currentMessages,
			CostPerDay:    costPerDay,
		}
	}

	totalAPI := 0.0
	totalPlan := 0.0
	for _, p := range periods {
		totalAPI += p.APICost
		totalPlan += p.PlanCostUSD
	}

	overallROI := 0.0
	if totalPlan > 0 {
		overallROI = math.Round(totalAPI/totalPlan*10) / 10
	}

	return PlanAnalysis{
		Periods:        periods,
		CurrentBilling: currentBilling,
		TotalAPICost:   math.Round(totalAPI*100) / 100,
		TotalPlanCost:  math.Round(totalPlan*100) / 100,
		TotalSavings:   math.Round((totalAPI-totalPlan)*100) / 100,
		OverallROI:     overallROI,
	}
}

// clampedDate creates a date clamping day to the last valid day of the month.
func clampedDate(year int, month time.Month, day int) time.Time {
	// Get last day of month
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
