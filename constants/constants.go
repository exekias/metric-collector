package constants

const (
	// Exchange name
	Exchange = "metric_collector"

	// DistinctName metric queue
	DistinctName = "distinctName"

	// HourlyLog metric queue
	HourlyLog = "hourlyLog"

	// AccountName metric queue
	AccountName = "accountName"
)

// Queues supported
var Queues = [...]string{DistinctName, HourlyLog, AccountName}
