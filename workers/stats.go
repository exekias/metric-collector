package workers

import (
	"expvar"
	"sync"
	"time"

	"github.com/exekias/metric-collector/queue"
)

var (
	numMetrics     = expvar.NewInt("n_metrics")
	numErrors      = expvar.NewInt("n_errors")
	avgCount       = expvar.NewFloat("avg_count")
	avgProcessTime = expvar.NewFloat("avg_process_time") // in seconds
)

// StatsProcessor wraps another processor and stores stats of it
// How many requests were handled? How long did they take? What are our
// average values? etc.
type StatsProcessor struct {
	// Mutex *for writting*, we can affort unconsistent reads
	sync.Mutex
	// NumMetrics processed
	NumMetrics int64
	// NumErrors that happened
	NumErrors int64
	// AvgCount from all metrics
	AvgCount float64
	// AvgProcessTime in seconds
	AvgProcessTime float64
	// Publish stats in expvars? Only one statsprocessor should do this at a time
	public bool

	processor MetricDataProcessor
}

// Stats wraps a given processor and stores stats on processed metrics
func Stats(p MetricDataProcessor, public bool) MetricDataProcessor {
	return &StatsProcessor{
		processor: p,
		public:    public,
	}
}

// Process using wrapped worker and store stats on the result
func (stats *StatsProcessor) Process(d queue.MetricData) error {
	start := time.Now()

	// Wrapped process
	err := stats.processor.Process(d)

	// Store stats
	stats.Lock()
	stats.AvgCount = nextAvg(stats.AvgCount, stats.NumMetrics, float64(d.Count))
	stats.AvgProcessTime = nextAvg(stats.AvgProcessTime, stats.NumMetrics, time.Since(start).Seconds())
	stats.NumMetrics++
	stats.Unlock()

	// Copy to expvar
	stats.publish()

	return err
}

func nextAvg(current float64, count int64, newVal float64) float64 {
	return (current*float64(count) + newVal) / float64(count+1)
}

func (stats *StatsProcessor) publish() {
	numMetrics.Set(stats.NumMetrics)
	numErrors.Set(stats.NumErrors)
	avgCount.Set(stats.AvgCount)
	avgProcessTime.Set(stats.AvgProcessTime)
}
