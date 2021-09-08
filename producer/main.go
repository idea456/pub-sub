package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/idea456/pub-sub/monitoring"
	"github.com/idea456/pub-sub/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
)

// simple function that prints out the error to the logs
func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// connect to the RabbitMQ container
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	logError(err)
	// close the connection if there is an error and program exits early
	defer conn.Close()

	// create a new channel
	ch, err := conn.Channel()
	logError(err)
	// close the channel if there is an error
	defer ch.Close()

	// declare new exchange
	err = ch.ExchangeDeclare(
		"requests", // exchange name
		"direct",   // exchange type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	logError(err)

	// consume messages from the reply-to callback queue
	// here we can obtain our responses from the consumers
	msgs, err := ch.Consume(
		"amq.rabbitmq.reply-to", // use direct reply-to
		"",                      // consumer
		true,                    // auto-ack
		false,                   // exclusive: When the connection that declared it closes, the queue will be deleted because it is declared as exclusive
		false,                   // no-local
		false,                   // no-wait
		nil,                     // args
	)
	logError(err)

	// initialize and register our Prometheus metrics
	if err := monitoring.Init(); err != nil {
		log.Fatal("unable to init monitoring, err: ", err.Error())
	}

	// create a new router to handle our requests
	router := mux.NewRouter()
	// track incoming requests and responses to Prometheus using a middleware
	router.Use(monitoring.Middleware)

	// Prometheus Handler
	// We declare Prometheus metrics endpoint and listen on the server port
	corrId := "abc"
	router.Path("/prometheus").Handler(promhttp.Handler())

	// define handlers for each path
	router.HandleFunc("/print", monitoring.Calculate(ch, corrId, "print")).Methods(http.MethodPost)
	router.HandleFunc("/pi", monitoring.Calculate(ch, corrId, "pi")).Methods(http.MethodPost)
	router.HandleFunc("/fibonacci", monitoring.Calculate(ch, corrId, "fibonacci")).Methods(http.MethodPost)
	router.HandleFunc("/prime", monitoring.Calculate(ch, corrId, "prime")).Methods(http.MethodPost)

	// initialize a goroutine to listen for responses in the msgs channel
	monitoring.ObserveResponse(corrId, msgs)

	serverConfig := server.Config{
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Port:         9000,
	}
	// start our server
	server.Serve(serverConfig, router)
}
