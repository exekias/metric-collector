package workers

import (
	"testing"
	"time"

	"github.com/exekias/metric-collector/queue"
)

// Helper debug processor with channel notification
type Debug struct {
	processed chan queue.MetricData
}

func (c Debug) Process(data queue.MetricData) error {
	c.processed <- data
	return nil
}

func TestRunWorkerCallsProcess(t *testing.T) {
	channel := queue.Dummy()
	channel.DeclareExchange("foo", true)
	channel.DeclareQueue("foo", "bar", true)
	debug := Debug{processed: make(chan queue.MetricData, 0)}

	go func() {
		channel.PublishMetric("foo", &queue.MetricData{"user", 0, "sample_metric"})
		channel.Close()
	}()

	RunWorker(channel, "bar", debug)

	select {
	case <-debug.processed:
		return
	case <-time.After(500 * time.Millisecond):
		t.Error("Data sent trough the queue was not processed")
	}
}
