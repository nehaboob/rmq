package main

import (
	"fmt"
	"log"
	"time"

	"github.com/adjust/rmq/v2"
)

const (
	numDeliveries = 100000000
	batchSize     = 10000
)

func main() {
	connection, err := rmq.OpenConnection("producer", "tcp", "localhost:6379", 2)
	if err != nil {
		panic(err)
	}

	things, err := connection.OpenQueue("things")
	if err != nil {
		panic(err)
	}
	balls, err := connection.OpenQueue("balls")
	if err != nil {
		panic(err)
	}

	var before time.Time
	for i := 0; i < numDeliveries; i++ {
		delivery := fmt.Sprintf("delivery %d", i)
		if _, err := things.Publish(delivery); err != nil {
			log.Printf("failed to publish: %s", err)
		}

		if i%batchSize == 0 {
			duration := time.Now().Sub(before)
			before = time.Now()
			perSecond := time.Second / (duration / batchSize)
			log.Printf("produced %d %s %d", i, delivery, perSecond)
			if _, err := balls.Publish("ball"); err != nil {
				log.Printf("failed to publish: %s", err)
			}
		}
	}
}
