package monitoring

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/streadway/amqp"
)

var PrintTimer *prometheus.Timer
var PiTimer *prometheus.Timer

func Calculate(ch *amqp.Channel, corrId string, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read our request body into byte array
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		var message Message
		// unmarshall the request body json into our Message structure data
		json.Unmarshal(reqBody, &message)

		log.Printf("Message sent with title %s and content: %s", message.Title, message.Content)

		// start tracking our response time
		PrintTimer = prometheus.NewTimer(Latency.WithLabelValues("/"+path, "POST"))
		// publish the message to the exchange
		err = ch.Publish(
			"requests", // exchange name
			path,       // routing key
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrId,                  // correlation id indicates the id of the producer in which the message is sent from
				ReplyTo:       "amq.rabbitmq.reply-to", // set our reply-to routing key to the callback queue in RabbitMQ
				Body:          []byte(message.Content),
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}
