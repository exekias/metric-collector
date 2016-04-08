package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/exekias/metric-collector/constants"
	"github.com/exekias/metric-collector/logging"
	"github.com/exekias/metric-collector/queue"
	"github.com/exekias/metric-collector/util"
	"github.com/exekias/metric-collector/workers"
	"github.com/exekias/metric-collector/workers/accountname"
	"github.com/exekias/metric-collector/workers/distinctname"
	"github.com/exekias/metric-collector/workers/hourlylog"
)

var log = logging.MustGetLogger("main")
var debug = flag.Bool("debug", false, "Enable debug")

const (
	// MongoDatabase to use
	MongoDatabase = "metrics"
	// MongoCollection to use
	MongoCollection = "hourly"
)

// MongoURL to connect to
var MongoURL = util.Getenv("MONGO_URL", "localhost:27017")

// RabbitMQURL server URL
var RabbitMQURL = util.Getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

// RedisURL server URL
var RedisURL = util.Getenv("REDIS_URL", "localhost:6379")

// PostgresURL server URL
var PostgresURL = util.Getenv("POSTGRES_URL", "postgres://metrics:foobar@localhost/metrics?sslmode=disable")

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "Args:")
		fmt.Fprintln(os.Stderr, "  -hourlylog - runs hourly log worker")
		fmt.Fprintln(os.Stderr, "  -distinctname - runs distinct name worker")
		fmt.Fprintln(os.Stderr, "  -accountname - runs account name worker")
	}
}

func main() {
	flag.Parse()
	if *debug {
		logging.SetLevel(logging.DEBUG, "")
	}

	if flag.NArg() != 1 {
		flag.Usage()
		return
	}

	// Init processor
	log.Info(fmt.Sprintf("Initializing metric '%s' processor", flag.Arg(0)))
	processor, queue := initProcessor(flag.Arg(0))

	// Init queue consumer
	log.Info(fmt.Sprintf("Initializing queue consumer %s", flag.Arg(0)))
	channel := initConsumer()

	log.Info("Serving stats in http://localhost:8080/debug/vars")
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	log.Info("Starting worker")
	workers.RunWorker(channel, queue, workers.Stats(processor, true))
	os.Exit(1)
}

func initProcessor(name string) (workers.MetricDataProcessor, string) {
	var processor workers.MetricDataProcessor
	var queue string
	var err error

	switch name {
	case "hourlylog":
		processor, err = hourlylog.NewHourlyLog(MongoURL, MongoDatabase, MongoCollection)
		queue = constants.HourlyLog

	case "distinctname":
		processor, err = distinctname.NewDistinctName(RedisURL)
		queue = constants.DistinctName

	case "accountname":
		processor, err = accountname.NewAccountName(PostgresURL)
		queue = constants.AccountName

	default:
		fmt.Printf("Unkown processor '%s'\n\n", name)
		flag.Usage()
		os.Exit(2)
	}

	if err != nil {
		log.Fatal("Error initializing processor: ", err)
	}

	return processor, queue
}

func initConsumer() queue.Channel {
	var ch queue.Channel

	// Connect to RabbitMQ
	log.Info("Connecting to RabbitMQ")
	ch, err := queue.RabbitMQ(RabbitMQURL)
	if err != nil {
		log.Fatal("Error connecting to RabbitMQ: ", err)
	}

	// Init queues
	log.Debug(fmt.Sprintf("Declaring exchange '%s'", constants.Exchange))
	if err = ch.DeclareExchange(constants.Exchange, true); err != nil {
		log.Fatal("Could not declare queue exchange: ", err)
	}
	for _, queue := range constants.Queues {
		log.Debug(fmt.Sprintf("Declaring queue '%s'", queue))
		if err = ch.DeclareQueue(constants.Exchange, queue, true); err != nil {
			log.Fatal("Could not declare queue: ", err)
		}
	}

	return ch
}
