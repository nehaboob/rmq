package rmq

import (
	"context"
	"fmt"
	"time"
)

type Delivery interface {
	Payload() string

	AckWithRetry(context.Context, chan<- error) error
	RejectWithRetry(context.Context, chan<- error) error
	PushWithRetry(context.Context, chan<- error) error

	// NOTE: These are simple low level functions, but potentially dangerous if
	// used naively. It's strongly recommended to use the functions above
	// (AckWithRetry() etc.) instead unless you know what you are doing.
	Ack() error
	Reject() error
	Push() error
}

type redisDelivery struct {
	payload     string
	unackedKey  string
	rejectedKey string
	pushKey     string
	redisClient RedisClient
}

func newDelivery(payload, unackedKey, rejectedKey, pushKey string, redisClient RedisClient) *redisDelivery {
	return &redisDelivery{
		payload:     payload,
		unackedKey:  unackedKey,
		rejectedKey: rejectedKey,
		pushKey:     pushKey,
		redisClient: redisClient,
	}
}

func (delivery *redisDelivery) String() string {
	return fmt.Sprintf("[%s %s]", delivery.payload, delivery.unackedKey)
}

func (delivery *redisDelivery) Payload() string {
	return delivery.payload
}

// blocking versions of the functions below with the following behavior:
// 1. return immediately if the operation succeeded or failed with ErrorNotFound
// 2. in case of other redis errors, send them to the errors chan and retry after a sleep
// 3. if the context is cancalled or its timeout exceeded, context.Cancelled or
//    context.DeadlineExceeded will be returned

func (delivery *redisDelivery) AckWithRetry(ctx context.Context, errChan chan<- error) error {
	return delivery.ackWithRetry(ctx, errChan, 0)
}

func (delivery *redisDelivery) ackWithRetry(ctx context.Context, errChan chan<- error, errorCount int) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		switch err := delivery.Ack(); err {
		case nil, ErrorNotFound:
			return err
		default: // redis error
			errorCount++

			select { // try to add error to channel, but don't block
			case errChan <- &DeliveryError{Delivery: delivery, RedisErr: err, Count: errorCount}:
			default:
			}

			if err := ctx.Err(); err != nil {
				return err
			}

			time.Sleep(time.Second)
		}
	}
}

func (delivery *redisDelivery) RejectWithRetry(ctx context.Context, errChan chan<- error) error {
	return delivery.moveWithRetry(ctx, errChan, delivery.rejectedKey)
}

func (delivery *redisDelivery) PushWithRetry(ctx context.Context, errChan chan<- error) error {
	if delivery.pushKey == "" {
		return delivery.RejectWithRetry(ctx, errChan) // fall back to rejecting
	}

	return delivery.moveWithRetry(ctx, errChan, delivery.pushKey)
}

func (delivery *redisDelivery) moveWithRetry(ctx context.Context, errChan chan<- error, key string) error {
	errorCount := 0
	for {
		_, err := delivery.redisClient.LPush(key, delivery.payload)
		if err == nil { // success
			break
		}
		// error

		errorCount++

		select { // try to add error to channel, but don't block
		case errChan <- &DeliveryError{Delivery: delivery, RedisErr: err, Count: errorCount}:
		default:
		}

		time.Sleep(time.Second)
	}

	return delivery.ackWithRetry(ctx, errChan, errorCount)
}

// lower level functions which don't retry but just return the first error

func (delivery *redisDelivery) Ack() error {
	count, err := delivery.redisClient.LRem(delivery.unackedKey, 1, delivery.payload)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrorNotFound
	}
	return nil
}

func (delivery *redisDelivery) Reject() error {
	return delivery.move(delivery.rejectedKey)
}

func (delivery *redisDelivery) Push() error {
	if delivery.pushKey == "" {
		return delivery.Reject() // fall back to rejecting
	}

	return delivery.move(delivery.pushKey)
}

func (delivery *redisDelivery) move(key string) error {
	if _, err := delivery.redisClient.LPush(key, delivery.payload); err != nil {
		return err
	}

	return delivery.Ack()
}
