package main

import (
	"time"

	"github.com/streadway/amqp"
)

func PrintSmth(msg amqp.Delivery) {
	time.Sleep(time.Duration(1) * time.Second)
}
