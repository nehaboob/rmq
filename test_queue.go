package rmq

import "time"

type TestQueue struct {
	name           string
	LastDeliveries []string
}

func NewTestQueue(name string) *TestQueue {
	queue := &TestQueue{name: name}
	queue.Reset()
	return queue
}

func (queue *TestQueue) String() string {
	return queue.name
}

func (queue *TestQueue) Publish(payload ...string) error {
	queue.LastDeliveries = append(queue.LastDeliveries, payload...)
	return nil
}

func (queue *TestQueue) PublishBytes(payload ...[]byte) error {
	stringifiedBytes := make([]string, len(payload))
	for i, b := range payload {
		stringifiedBytes[i] = string(b)
	}
	return queue.Publish(stringifiedBytes...)
}

func (queue *TestQueue) SetPushQueue(pushQueue Queue) {
}

func (queue *TestQueue) StartConsuming(prefetchLimit int64, pollDuration time.Duration) error {
	return nil
}

func (queue *TestQueue) StopConsuming() <-chan struct{} {
	return nil
}

func (queue *TestQueue) AddConsumer(tag string, consumer Consumer) (string, error) {
	return "", nil
}

func (queue *TestQueue) AddConsumerFunc(tag string, consumerFunc ConsumerFunc) (string, error) {
	return "", nil
}

func (queue *TestQueue) AddBatchConsumer(tag string, batchSize int64, consumer BatchConsumer) (string, error) {
	return "", nil
}

func (queue *TestQueue) AddBatchConsumerWithTimeout(tag string, batchSize int64, timeout time.Duration, consumer BatchConsumer) (string, error) {
	return "", nil
}

func (queue *TestQueue) ReturnRejected(count int64) (int64, error) {
	return 0, nil
}

func (queue *TestQueue) ReturnAllRejected() (int64, error) {
	return 0, nil
}

func (queue *TestQueue) PurgeReady() (int64, error) {
	return 0, nil
}

func (queue *TestQueue) PurgeRejected() (int64, error) {
	return 0, nil
}

func (queue *TestQueue) Close() (bool, error) {
	return false, nil
}

func (queue *TestQueue) Reset() {
	queue.LastDeliveries = []string{}
}
