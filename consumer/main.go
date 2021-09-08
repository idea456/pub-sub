package main

import (
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PerformOperation(tick int, msg amqp.Delivery) {
	log.Printf("Performing operation for %d second...", tick)
	time.Sleep(time.Duration(tick) * time.Second)
}

func main() {
	// connect to the RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	logError(err)
	defer conn.Close()

	// create a new channel
	ch, err := conn.Channel()
	logError(err)
	defer ch.Close()

	// declare the same exchange as the producer
	err = ch.ExchangeDeclare(
		"requests",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	logError(err)

	// declare the same queue as the producer since the consumer could be started before the producer
	// and the same queue should be ready and available for both sides
	q, err := ch.QueueDeclare(
		"",    // name: when queue name is empty, auto generate a queue name
		false, // durable
		false, // delete when unused
		true,  // exclusive: When the connection that declared it closes, the queue will be deleted because it is declared as exclusive
		false, // no-wait
		nil,   // arguments
	)
	logError(err)

	// define our quality of service
	err = ch.Qos(
		1,     // prefetch count (here At Least Once): how many acknowledgements should it receive before it can send further messages
		0,     // prefetch size: how many bytes of deliveries to keep before receiving acknowledgements
		false, // global: apply to all queues on the SAME CHANNEL
	)
	logError(err)

	// create the binding connection between the queue and the exchange
	// based on our routing key explicitly stated in our command-line argument
	for _, routingKey := range os.Args[1:] {
		// We want to tell the exchange to send our message to the queue
		// the relationship between an exchange and a queue is called binding
		err = ch.QueueBind(
			q.Name,
			routingKey,
			"requests",
			false,
			nil,
		)
		logError(err)
	}

	// consume messages received from the producer
	msgs, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer
		false,  // auto-ack
		false,
		false,
		false,
		nil,
	)
	logError(err)

	forever := make(chan bool)
	// use goroutines to listen for incoming messages from the msgs channel
	go func() {
		for msg := range msgs {
			var reply string
			log.Printf("Message received %s!", msg.Body)
			switch msg.RoutingKey {
			case "print":
				PerformOperation(1, msg)
				reply = "Printed message!"
			case "pi":
				PerformOperation(2, msg)
				reply = "Calculated Pi!"
			case "fibonacci":
				PerformOperation(3, msg)
				reply = "Calculated Fibonacci!"
			case "prime":
				PerformOperation(4, msg)
				reply = "Calculated Prime!"
			default:
				time.Sleep(time.Duration(1) * time.Second)
				log.Printf("[Consumer] Undefined routing key!\n")
			}

			err = ch.Publish(
				"",          // exchange
				msg.ReplyTo, // routing key
				false,       // mandatory: when the routing key doesnt match any queue, discard that message if mandatory is false
				false,       // immediate: if there are no consumers connected or ready to accept the message, discard it if immediate is true
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: msg.CorrelationId,
					Body:          []byte(reply),
				})
			logError(err)
			// acknowledge the delivery
			msg.Ack(false)
		}
		log.Printf("[Consumer][*] Waiting for new messages...\n")
	}()

	log.Printf("[Consumer][*] Waiting for new messages...\n")
	<-forever
}
