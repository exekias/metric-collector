package queue

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

// RabbitMQChannel implements queue.Channel
type RabbitMQChannel struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// RabbitMQMetricMessage implementes queue.Metric
type RabbitMQMetricMessage struct {
	d    amqp.Delivery
	data MetricData
}

// RabbitMQ creates a Rabbit MQ connection to the given URL and returns a channel
func RabbitMQ(url string) (*RabbitMQChannel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQChannel{
		conn:    conn,
		channel: ch,
	}, nil
}

// DeclareExchange creates a exchnage (fanout type)
func (c *RabbitMQChannel) DeclareExchange(exchange string, durable bool) error {
	return c.channel.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		durable,  // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
}

// DeclareQueue creates a queue, binds it to the exchange and sets its durability settings
func (c *RabbitMQChannel) DeclareQueue(exchange string, queue string, durable bool) error {
	_, err := c.channel.QueueDeclare(
		queue,   // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		queue,    // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil)

	return err
}

// PublishMetric to the given exchange
func (c *RabbitMQChannel) PublishMetric(exchange string, metric *MetricData) error {
	msg, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	return c.channel.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         msg,
		})
}

// ConsumeMetrics returns a channel receiving metrics from the given queue
func (c *RabbitMQChannel) ConsumeMetrics(queue string) (<-chan MetricMessage, error) {
	// Set QoS settings, TODO: benchmark and finetune
	err := c.channel.Qos(
		3,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}

	msgs, err := c.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	// Create generator consumer goroutine
	res := make(chan MetricMessage)
	go func() {
		for d := range msgs {
			res <- &RabbitMQMetricMessage{d: d}
		}
	}()

	return res, nil
}

// Close the connection, must be called when no longer necessary
func (c *RabbitMQChannel) Close() error {
	// Close channel and connection
	if err := c.channel.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}

	return nil
}

// METRIC MESSAGE METHODS:

func (m *RabbitMQMetricMessage) unmarshalIfNeeded() error {
	return json.Unmarshal(m.d.Body, &m.data)
}

// MetricData in the metric
func (m *RabbitMQMetricMessage) MetricData() (MetricData, error) {
	err := m.unmarshalIfNeeded()
	return m.data, err
}

// Ack acknowledges metric processed (and stored) correctly
func (m *RabbitMQMetricMessage) Ack() error {
	return m.d.Ack(false)
}
