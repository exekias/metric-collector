package distinctname

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/redis.v3"

	"github.com/exekias/metric-collector/queue"
)

func TestDistinctName(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		t.Skip("redis server not available")
	}

	processor, err := NewDistinctName("localhost:6379")
	if err != nil {
		t.Error(err)
	}

	if err = processor.Process(queue.MetricData{"user1", 5, "metric1"}); err != nil {
		t.Error("Processing a metric", err)
	}

	if err = processor.Process(queue.MetricData{"user1", 1, "metric1"}); err != nil {
		t.Error("Processing a metric", err)
	}

	if err = processor.Process(queue.MetricData{"user1", 7, "metric2"}); err != nil {
		t.Error("Processing a metric", err)
	}

	now := time.Now()
	set := fmt.Sprintf("metrics:dn:%d:%d", now.Month(), now.Day())
	res, err := client.ZRangeWithScores(set, 0, -1).Result()
	if err != nil {
		t.Error("Could not get result from redis", err)
	}

	metrics := 0
	for _, v := range res {
		switch v.Member {
		case "metric1":
			if v.Score != 2 {
				t.Error("Wrong data for 'metric1'")
			}
			metrics++
			continue
		case "metric2":
			if v.Score != 1 {
				t.Error("Wrong data for 'metric1'")
			}
			metrics++
			continue
		default:
			t.Errorf("Unknown member in the set: %#v", v)
		}
	}

	if metrics != 2 {
		t.Error("Metrics missing in the set")
	}
}

func TestDistinctNameConsolidation(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		t.Skip("redis server not available")
	}

	processor, err := NewDistinctName("localhost:6379")
	if err != nil {
		t.Error(err)
	}

	// Insert values from past month
	now := time.Now()
	date := time.Date(now.Year(), now.Month()-1, 5, 0, 0, 0, 0, time.UTC)
	if err = processor.insert(date, &queue.MetricData{"user1", 5, "metric1"}); err != nil {
		t.Error("Processing a metric", err)
	}

	// one day later...
	date = date.Add(24 * time.Hour)
	if err = processor.insert(date, &queue.MetricData{"user1", 1, "metric1"}); err != nil {
		t.Error("Processing a metric", err)
	}

	date = date.Add(24 * time.Hour)
	if err = processor.insert(date, &queue.MetricData{"user1", 7, "metric2"}); err != nil {
		t.Error("Processing a metric", err)
	}

	processor.monthlyConsolidate()
	set := fmt.Sprintf("metrics:dn:%d:%d", now.Month(), now.Day())
	res, err := client.ZRangeWithScores(set, 0, -1).Result()
	if err != nil {
		t.Error("Could not get result from redis", err)
	}

	metrics := 0
	for _, v := range res {
		switch v.Member {
		case "metric1":
			if v.Score != 2 {
				t.Error("Wrong data for 'metric1'")
			}
			metrics++
			continue
		case "metric2":
			if v.Score != 1 {
				t.Error("Wrong data for 'metric1'")
			}
			metrics++
			continue
		default:
			t.Errorf("Unknown member in the set: %#v", v)
		}
	}

	if metrics != 2 {
		t.Error("Metrics missing in the set")
	}
}

func TestDaysIn(t *testing.T) {
	var tests = []struct {
		m      time.Month
		y, res int
	}{
		{1, time.Now().Year(), 31},
		{3, time.Now().Year(), 31},
		{4, time.Now().Year(), 30},
		{5, time.Now().Year(), 31},
		{6, time.Now().Year(), 30},
		{7, time.Now().Year(), 31},
		{8, time.Now().Year(), 31},
		{9, time.Now().Year(), 30},
		{10, time.Now().Year(), 31},
		{11, time.Now().Year(), 30},
		{12, time.Now().Year(), 31},
		// Feb:
		{2, 2000, 29},
		{2, 2015, 28},
		{2, 2016, 29},
	}

	for _, v := range tests {
		if x := daysIn(v.m, v.y); x != v.res {
			t.Errorf("Incorrect number of days in %d/%d: %d (%d expected)",
				v.m, v.y, x, v.res)
		}
	}
}
