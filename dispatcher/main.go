// Dispatcher app, sends random metrics to the exchange, only for testing
// as this should come from a different app
package main

import (
	"math/rand"
	"time"

	"github.com/exekias/metric-collector/constants"
	"github.com/exekias/metric-collector/logging"
	"github.com/exekias/metric-collector/queue"
	"github.com/exekias/metric-collector/util"
)

const (
	// MaxRetries when sending messages to a queue
	MaxRetries = 3

	// RabbitMQURL server URL
	RabbitMQURL = "amqp://guest:guest@localhost:5672/"

	// RateLimit sent messages to ~10/second
	RateLimit = time.Second / 10
)

var log = logging.MustGetLogger("dispatcher")

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	var ch queue.Channel

	// Connect to RabbitMQ
	log.Info("Connecting to RabbitMQ")
	ch, err := queue.RabbitMQ(RabbitMQURL)
	if err != nil {
		log.Fatal("Error connecting to RabbitMQ: ", err)
	}

	// Init queues
	log.Debug("Declaring exchange '%s'", constants.Exchange)
	if err := ch.DeclareExchange(constants.Exchange, true); err != nil {
		log.Fatal("Could not declare queue exchange: ", err)
	}
	for _, queue := range constants.Queues {
		log.Debug("Declaring queue '%s'", queue)
		if err := ch.DeclareQueue(constants.Exchange, queue, true); err != nil {
			log.Fatal("Could not declare queue: ", err)
		}
	}

	// Send random metrics, forever
	log.Info("Initialization done, sending random messages")
	var data queue.MetricData
	for range time.Tick(RateLimit) {

		// Generate random data
		data.Username = random(usernames)
		data.Count = rand.Int63n(10000)
		data.Metric = random(metrics)
		log.Debug("Sending %#v", data)

		// Send it
		retries := 0
		backoff := util.ExponentialBackoff(300 * time.Millisecond)
		for retries < MaxRetries {
			if err := ch.PublishMetric(constants.Exchange, &data); err != nil {
				log.Warning("Could not publish message, retrying (%d)", retries)
				retries++
				backoff()
			} else {
				break // Everything ok
			}
		}

		if retries >= MaxRetries {
			log.Fatal("Error delivering a message, dying")
		}
	}
}

// Useful helpers:

func random(source []string) string {
	return source[rand.Intn(len(source))]
}

var usernames = []string{
	"cihangir",
	"didemacet",
	"fatihacet",
	"gokmen",
	"mehmetalisavas",
	"rjeczalik",
	"sinan",
	"usirin",
	"exekias",
}

// Got this using  github.com/docker/docker/pkg/namesgenerator:
var metrics = []string{
	"prickly_swanson",
	"distracted_noyce",
	"jolly_borg",
	"stupefied_goodall",
	"nauseous_swartz",
	"sharp_easley",
	"serene_archimedes",
	"drunk_dubinsky",
	"kickass_cray",
	"admiring_bhaskara",
	"reverent_kilby",
	"gloomy_yalow",
	"hopeful_boyd",
	"determined_sinoussi",
	"grave_swanson",
	"big_cray",
	"elated_chandrasekhar",
	"hopeful_colden",
	"clever_cray",
	"pensive_bhaskara",
	"mad_wozniak",
	"backstabbing_jang",
	"agitated_kowalevski",
	"elegant_elion",
}
