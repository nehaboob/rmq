package main

import (
	"time"

	"github.com/adjust/rmq/v2"
)

func main() {
	connection, err := rmq.OpenConnection("cleaner", "tcp", "localhost:6379", 2)
	if err != nil {
		panic(err)
	}

	cleaner := rmq.NewCleaner(connection)

	for _ = range time.Tick(time.Second) {
		cleaner.Clean()
	}
}
