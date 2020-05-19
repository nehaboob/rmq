package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adjust/rmq/v2"
)

const (
	unackedLimit    = 1000
	batchSize       = 111
	pollDuration    = 500 * time.Millisecond
	batchTimeout    = time.Second
	consumeDuration = time.Millisecond
	shouldLog       = false
)

func main() {
	errChan := make(chan error, 10)
	go func() {
		for err := range errChan {
			switch err := err.(type) {
			case *rmq.ConsumeError:
				log.Print("consume error: ", err)
			case *rmq.HeartbeatError:
				if err.Count == rmq.HeartbeatErrorLimit {
					log.Print("heartbeat error (limit): ", err)
				} else {
					log.Print("heartbeat error: ", err)
				}
			case *rmq.DeliveryError:
				log.Print("delivery error: ", err.Delivery, err)
			default:
				log.Print("other error: ", err)
			}
		}
	}()

	connection, err := rmq.OpenConnection("consumer", "tcp", "localhost:6379", 2, errChan)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	for _, queueName := range []string{
		"things",
		"balls",
	} {
		queue, err := connection.OpenQueue(queueName)
		if err != nil {
			panic(err)
		}
		if err := queue.StartConsuming(unackedLimit, pollDuration); err != nil {
			panic(err)
		}
		if _, err := queue.AddBatchConsumer(queueName, batchSize, batchTimeout, NewBatchConsumer(ctx, queueName)); err != nil {
			panic(err)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

	c := connection.StopAllConsuming()
	// make sure Ack() and similar calls return with error so that they can be
	// handled and the active Conume() calls can finish
	cancel()
	<-c // wait for all Conume() calls to finish
}

type BatchConsumer struct {
	ctx context.Context
	tag string
}

func NewBatchConsumer(ctx context.Context, tag string) *BatchConsumer {
	return &BatchConsumer{
		ctx: ctx,
		tag: tag,
	}
}

func (consumer *BatchConsumer) Consume(batch rmq.Deliveries) {
	payloads := batch.Payloads()
	debugf("start consume %q", payloads)
	time.Sleep(consumeDuration)

	log.Printf("%s consumed %d: %s", consumer.tag, len(batch), batch[0])
	errors := batch.Ack(consumer.ctx)
	if len(errors) == 0 {
		debugf("acked %q", payloads)
		return
	}

	for i, err := range errors {
		debugf("failed to ack %q: %q", batch[i].Payload(), err)
	}
}

func debugf(format string, args ...interface{}) {
	if shouldLog {
		log.Printf(format, args...)
	}
}
