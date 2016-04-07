package workers

import "github.com/exekias/metric-collector/queue"

// MetricDataProcessor processes/stores MetricData messages
type MetricDataProcessor interface {

	// Process given MetricData
	Process(data queue.MetricData) error
}
