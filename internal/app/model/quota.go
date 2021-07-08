package model

type QuotaType struct {
	TimeoutBlock    string
	HourlyQuota     uint64
	DailyQuota      uint64
	MonthlyQuota    uint64
	Disabled        bool
	Dynamic         bool
	Blocked         bool
	ShouldBeBlocked bool
}

func CheckNULLQuotas(setValue, deafultValue QuotaType) QuotaType {
	quotaReturned := setValue
	if setValue.DailyQuota == 0 {
		quotaReturned.DailyQuota = deafultValue.DailyQuota
	}
	if setValue.HourlyQuota == 0 {
		quotaReturned.HourlyQuota = deafultValue.HourlyQuota
	}
	if setValue.MonthlyQuota == 0 {
		quotaReturned.MonthlyQuota = deafultValue.MonthlyQuota
	}
	return quotaReturned
}

func CheckNULLQuota(q, qd uint64) uint64 {
	if q == 0 {
		q = qd
	}
	return q
}
