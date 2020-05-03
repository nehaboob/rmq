package main

import (
	"log"

	"github.com/adjust/rmq/v2"
)

func main() {
	connection, err := rmq.OpenConnection("returner", "tcp", "localhost:6379", 2)
	if err != nil {
		panic(err)
	}

	queue, err := connection.OpenQueue("things")
	if err != nil {
		panic(err)
	}
	returned, err := queue.ReturnAllRejected()
	if err != nil {
		panic(err)
	}

	log.Printf("queue returner returned %d rejected deliveries", returned)
}