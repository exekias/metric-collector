package workers

import (
	"github.com/exekias/metric-collector/logging"
	"github.com/exekias/metric-collector/queue"
)

var log = logging.MustGetLogger("worker")

// RunWorker listen for messages in the given channel and process trough the given processor
func RunWorker(channel queue.Channel, q string, processor MetricDataProcessor) {

	// Listen in the queue
	metrics, err := channel.ConsumeMetrics(q)
	if err != nil {
		log.Fatal("Could not consume from queue", err)
	}

	// Listen and process all metrics
	for metric := range metrics {
		// Process in a goroutine, channel QoS will handle throttling,
		// see queue/rabbitmq.go:ConsumeMetrics
		go func(m queue.MetricMessage) {
			data, err := m.MetricData()
			if err != nil {
				log.Error("Unexpected error reading metric data:", err)
			}
			log.Debug("Processing metric", data)
			if err := processor.Process(data); err == nil {
				// We are done
				m.Ack()
			} else {
				log.Warning("Error while processing a metric, won't ACK:", err)
			}
		}(metric)
	}
	log.Error("Disconnected from metrics queue")
}
