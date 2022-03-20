package main

import (
	"fmt"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:8080//dev")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Created connection")

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	fmt.Println("Created channel")

	err = ch.Publish("example-exchange", "my.routing.key", true, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte("Hello World"),
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("Published message")

	ch.Close()
	fmt.Println("closed channel")
}
