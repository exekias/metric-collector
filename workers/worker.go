package workers

import (
	"fmt"

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

	// Results channel, true = success, false = error
	results := make(chan bool)

	// Will get true when too many errors happened
	tooManyErrors := make(chan bool)

	go errorCheck(results, tooManyErrors)

	// Listen and process all metrics
	for metric := range metrics {
		// Process in a goroutine, channel QoS will handle throttling,
		// see queue/rabbitmq.go:ConsumeMetrics
		go func(m queue.MetricMessage, results chan<- bool) {
			data, err := m.MetricData()
			if err != nil {
				log.Error("Unexpected error reading metric data:", err)
			}
			log.Debug(fmt.Sprintf("Processing metric %#v", data))
			if err := processor.Process(data); err == nil {
				// We are done
				m.Ack()
				results <- true
			} else {
				log.Warning("Error while processing a metric, won't ACK:", err)
				m.Nack()
				results <- false
			}
		}(metric, results)

		select {
		case <-tooManyErrors:
			log.Error("Processor had too many errors, quitting...")
			return
		default:
		}
	}
	log.Error("Disconnected from metrics queue")
}

// errorCheck notifies of too many errors after 5 errors in a row
func errorCheck(results <-chan bool, tooManyErrors chan<- bool) {
	errors := 0
	for success := range results {
		if success {
			errors = 0
		} else {
			errors++
		}

		if errors >= 5 {
			tooManyErrors <- true
		}
	}
}
