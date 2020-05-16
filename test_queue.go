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

func (queue *TestQueue) Publish(payload ...string) (total int64, err error) {
	queue.LastDeliveries = append(queue.LastDeliveries, payload...)
	return int64(len(queue.LastDeliveries)), nil
}

func (queue *TestQueue) PublishBytes(payload ...[]byte) (total int64, err error) {
	stringifiedBytes := make([]string, len(payload))
	for i, b := range payload {
		stringifiedBytes[i] = string(b)
	}
	return queue.Publish(stringifiedBytes...)
}

func (queue *TestQueue) SetPushQueue(_ Queue)                         { panic(errorNotSupported) }
func (queue *TestQueue) StartConsuming(int64, time.Duration) error    { panic(errorNotSupported) }
func (queue *TestQueue) StopConsuming() <-chan struct{}               { panic(errorNotSupported) }
func (queue *TestQueue) AddConsumer(string, Consumer) (string, error) { panic(errorNotSupported) }
func (queue *TestQueue) AddConsumerFunc(string, ConsumerFunc) (string, error) {
	panic(errorNotSupported)
}
func (queue *TestQueue) AddBatchConsumer(string, int64, BatchConsumer) (string, error) {
	panic(errorNotSupported)
}
func (queue *TestQueue) AddBatchConsumerWithTimeout(string, int64, time.Duration, BatchConsumer) (string, error) {
	panic(errorNotSupported)
}
func (queue *TestQueue) ReturnRejected(int64) (int64, error) { panic(errorNotSupported) }
func (queue *TestQueue) ReturnAllUnacked() (int64, error)    { panic(errorNotSupported) }
func (queue *TestQueue) ReturnAllRejected() (int64, error)   { panic(errorNotSupported) }
func (queue *TestQueue) PurgeReady() (int64, error)          { panic(errorNotSupported) }
func (queue *TestQueue) PurgeRejected() (int64, error)       { panic(errorNotSupported) }
func (queue *TestQueue) Destroy() (int64, int64, error)      { panic(errorNotSupported) }
func (queue *TestQueue) closeInConnection()                  { panic(errorNotSupported) }
func (queue *TestQueue) readyCount() (int64, error)          { panic(errorNotSupported) }
func (queue *TestQueue) unackedCount() (int64, error)        { panic(errorNotSupported) }
func (queue *TestQueue) rejectedCount() (int64, error)       { panic(errorNotSupported) }
func (queue *TestQueue) getConsumers() ([]string, error)     { panic(errorNotSupported) }
func (queue *TestQueue) removeAllConsumers() (int64, error)  { panic(errorNotSupported) }
func (queue *TestQueue) removeConsumer(string) (bool, error) { panic(errorNotSupported) }

// test helper

func (queue *TestQueue) Reset() {
	queue.LastDeliveries = []string{}
}
