package distinctname

import (
	"fmt"
	"time"

	"github.com/exekias/metric-collector/logging"
	"github.com/exekias/metric-collector/queue"
	"gopkg.in/redis.v3"
)

var log = logging.MustGetLogger("distinctname")

// DistinctName worker collects daily occurrences of distinct events in Redis.
// Metrics that are older than 30 days are merged into a monthly bucket, then
// cleared.
type DistinctName struct {
	client *redis.Client
}

// NewDistinctName intializes and returns a new distinct name processor
func NewDistinctName(url string) (*DistinctName, error) {
	log.Debug(fmt.Sprintf("Connecting to Redis (%s)", url))
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Error("Error connecting to Redis")
		return nil, err
	}

	return initDistinctName(client), nil
}

func initDistinctName(client *redis.Client) *DistinctName {
	var processor DistinctName
	processor.client = client
	go processor.runConsolidate()
	return &processor
}

// Process data from the queue
func (p DistinctName) Process(d queue.MetricData) error {
	// Do a ZADD (INCR mode)
	return p.insert(time.Now(), &d)
}

func (p DistinctName) insert(t time.Time, d *queue.MetricData) error {
	set := dailySetName(t)
	value := redis.Z{float64(1), d.Metric}
	_, err := p.client.ZIncr(set, value).Result()
	return err
}

// runConsolidate calls `monthlyConsolidate` in an infinite loop
// this method runs the consolidate process every 10 mins (for demostration
// purposes)
// It would be enough to run it after changing month or on a monthly basis
func (p DistinctName) runConsolidate() {
	for range time.Tick(10 * time.Minute) {
		p.monthlyConsolidate()
	}
}

// monthlyConsolidate acumulates all daily results in a monthly set
func (p DistinctName) monthlyConsolidate() {
	log.Info("Checking past month consolidation status...")
	now := time.Now()
	pastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	exists, err := p.client.Exists(monthlySetName(pastMonth)).Result()
	if err != nil {
		log.Error("Error checking if monthly set exists", err)
	}
	// Already acumulated, wait for the next month
	if exists {
		log.Info("Consolidation was already done, nothing to do")
		return
	}

	// Acumulate all days in the month
	nDays := daysIn(pastMonth.Month(), pastMonth.Year())
	days := make([]string, nDays, nDays)
	for i := 0; i < nDays; i++ {
		days[i] = dailySetName(pastMonth.Add(time.Duration(i*24) * time.Hour))
	}

	set := monthlySetName(pastMonth)
	log.Info(fmt.Sprintf("Consolidating past month into %s", set))
	log.Debug(fmt.Sprintf("Daily sets: %#v", days))
	if _, err = p.client.ZUnionStore(
		set,
		redis.ZStore{Aggregate: "SUM"},
		days...,
	).Result(); err != nil {
		log.Error("Error creating monthly set", err)
	}

	log.Info("Consolidation done, cleaning up")
	if _, err = p.client.Del(days...).Result(); err != nil {
		log.Warning("Could not cleanup daily data after consolidation", err)
	}
}

// dailySetName returns daily set name for a given date
func dailySetName(t time.Time) string {
	return fmt.Sprintf("metrics:dn:%d:%d", t.Month(), t.Day())
}

// monthlySetName
func monthlySetName(t time.Time) string {
	return fmt.Sprintf("metrics:dn:%d", t.Month())
}

// daysIn return the number of days in a given month
func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
