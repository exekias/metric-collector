package queue

// Channel offers operations for defining queues, and sending / receiving tasks
type Channel interface {

	// DeclareExchange creates a exchnage (fanout type)
	DeclareExchange(exchange string, durable bool) error

	// DeclareQueue creates a queue, binds it to the exchange and sets its durability settings
	DeclareQueue(exchange string, queue string, durable bool) error

	// PublishMetric to the given exchnage
	PublishMetric(exchange string, metric *MetricData) error

	// ConsumeMetrics returns a channel receiving metrics from the given queue
	ConsumeMetrics(queue string) (<-chan MetricMessage, error)

	// Close the connection, must be called when no longer necessary
	Close() error
}

// MetricData delivered by the app
type MetricData struct {
	Username string `json:"username"`
	Count    int64  `json:"count"`
	Metric   string `json:"metric"`
}

// MetricMessage job sent trough a queue
type MetricMessage interface {
	// MetricData extracted from this delivery
	MetricData() (MetricData, error)

	// Ack acknowledges metric processed (and stored) correctly
	Ack() error

	// Nack negatively acknowledges the message, forcing requeuing
	Nack() error
}
