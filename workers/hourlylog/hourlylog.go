package hourlylog

import (
	"time"

	"gopkg.in/mgo.v2"
	_ "gopkg.in/tomb.v2" // to ensure dependency is satisfied

	"github.com/exekias/metric-collector/logging"
	"github.com/exekias/metric-collector/queue"
)

var log = logging.MustGetLogger("hourlylog")

// HourlyLog worker collects all items that occurred in the last hour into
// MongoDB
type HourlyLog struct {
	collection *mgo.Collection
}

type mongoMetric struct {
	queue.MetricData `bson:",inline"`
	Time             time.Time
}

// NewHourlyLog intializes and returns a new hourly log processor
func NewHourlyLog(session *mgo.Session, db, collection string) (*HourlyLog, error) {
	var processor HourlyLog

	processor.collection = session.DB(db).C(collection)

	// Setup index
	index := mgo.Index{
		Key:         []string{"time"},
		ExpireAfter: 1 * time.Hour,
	}
	if err := processor.collection.EnsureIndex(index); err != nil {
		log.Error("Error configuring MongoDB indexes")
		return nil, err
	}

	return &processor, nil
}

// Process data from the queue
func (h HourlyLog) Process(d queue.MetricData) error {
	return h.insert(&mongoMetric{d, time.Now()})
}

func (h HourlyLog) insert(data *mongoMetric) error {
	if err := h.collection.Insert(data); err != nil {
		log.Error("Error storing data in MongoDB:", err)
		return err
	}
	return nil
}
