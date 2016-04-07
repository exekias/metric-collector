package queue

import "errors"

// DummyChannel implements in memory queue.Channel
type DummyChannel struct {
	// map exchanges -> queues -> listener channels
	exchanges map[string]map[string][]chan MetricMessage
}

// DummyMetricMessage implementes queue.Metric
type DummyMetricMessage struct {
	data  MetricData
	Acked bool
}

// Dummy creates an in memory queue channel (for testing)
func Dummy() *DummyChannel {
	return &DummyChannel{
		exchanges: make(map[string]map[string][]chan MetricMessage),
	}
}

// DeclareExchange creates a exchnage (fanout type), no durable (ignores param)
func (c *DummyChannel) DeclareExchange(exchange string, durable bool) error {
	if c.exchanges[exchange] == nil {
		c.exchanges[exchange] = make(map[string][]chan MetricMessage)
	}
	return nil
}

// DeclareQueue creates a queue, binds it to the exchange, no durable (ignores param)
func (c *DummyChannel) DeclareQueue(exchange string, queue string, durable bool) error {
	if c.exchanges[exchange][queue] == nil {
		c.exchanges[exchange][queue] = make([]chan MetricMessage, 0)
	}
	return nil
}

// PublishMetric to the given exchange
func (c *DummyChannel) PublishMetric(exchange string, metric *MetricData) error {
	for queue := range c.exchanges[exchange] {
		for _, listener := range c.exchanges[exchange][queue] {
			listener <- &DummyMetricMessage{data: *metric, Acked: false}
		}
	}
	return nil
}

// ConsumeMetrics returns a channel receiving metrics from the given queue
func (c *DummyChannel) ConsumeMetrics(queue string) (<-chan MetricMessage, error) {
	for _, queues := range c.exchanges {
		if queues[queue] != nil {
			ch := make(chan MetricMessage)
			queues[queue] = append(queues[queue], ch)
			return ch, nil
		}
	}
	return nil, errors.New("Queue not found")
}

// Close the connection, must be called when no longer necessary
func (c *DummyChannel) Close() error {
	for _, queues := range c.exchanges {
		for _, chs := range queues {
			for _, ch := range chs {
				close(ch)
			}
		}
	}
	return nil
}

// METRIC MESSAGE METHODS:

// MetricData in the metric
func (m *DummyMetricMessage) MetricData() (MetricData, error) {
	return m.data, nil
}

// Ack acknowledges metric processed (and stored) correctly
func (m *DummyMetricMessage) Ack() error {
	m.Acked = true
	return nil
}
