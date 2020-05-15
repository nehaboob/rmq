package rmq

import (
	"testing"
	"time"

	// TODO: use testify
	. "github.com/adjust/gocheck"
)

func TestCleanerSuite(t *testing.T) {
	TestingSuiteT(&CleanerSuite{}, t)
}

type CleanerSuite struct{}

func (suite *CleanerSuite) TestCleaner(c *C) {
	flushConn, err := OpenConnection("cleaner-flush", "tcp", "localhost:6379", 1)
	c.Check(err, IsNil)
	flushConn.flushDb()
	flushConn.stopHeartbeat()

	conn, err := OpenConnection("cleaner-conn1", "tcp", "localhost:6379", 1)
	c.Check(err, IsNil)
	queues, err := conn.GetOpenQueues()
	c.Check(err, IsNil)
	c.Check(queues, HasLen, 0)
	queue := conn.OpenQueue("q1")
	queues, err = conn.GetOpenQueues()
	c.Check(err, IsNil)
	c.Check(queues, HasLen, 1)
	conn.OpenQueue("q2")
	queues, err = conn.GetOpenQueues()
	c.Check(err, IsNil)
	c.Check(queues, HasLen, 2)

	count, err := queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(0))
	queue.Publish("del1")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(1))
	queue.Publish("del2")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(2))
	queue.Publish("del3")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))
	queue.Publish("del4")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(4))
	queue.Publish("del5")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(5))
	queue.Publish("del6")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(6))

	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(0))
	queue.StartConsuming(2, time.Millisecond)
	time.Sleep(time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(2))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(4))

	consumer := NewTestConsumer("c-A")
	consumer.AutoFinish = false
	consumer.AutoAck = false

	queue.AddConsumer("consumer1", consumer)
	time.Sleep(10 * time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))

	c.Assert(consumer.LastDelivery, NotNil)
	c.Check(consumer.LastDelivery.Payload(), Equals, "del1")
	err = consumer.LastDelivery.Ack()
	c.Check(err, IsNil)
	time.Sleep(10 * time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(2))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))

	consumer.Finish()
	time.Sleep(10 * time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(2))
	c.Check(consumer.LastDelivery.Payload(), Equals, "del2")

	queue.StopConsuming()
	conn.stopHeartbeat()
	time.Sleep(time.Millisecond)

	conn, err = OpenConnection("cleaner-conn1", "tcp", "localhost:6379", 1)
	c.Check(err, IsNil)
	queue = conn.OpenQueue("q1")

	queue.Publish("del7")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))
	queue.Publish("del7")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(4))
	queue.Publish("del8")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(5))
	queue.Publish("del9")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(6))
	queue.Publish("del10")
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(7))

	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(0))
	queue.StartConsuming(2, time.Millisecond)
	time.Sleep(time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(2))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(5))

	consumer = NewTestConsumer("c-B")
	consumer.AutoFinish = false
	consumer.AutoAck = false

	queue.AddConsumer("consumer2", consumer)
	time.Sleep(10 * time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(4))
	c.Check(consumer.LastDelivery.Payload(), Equals, "del5")

	consumer.Finish() // unacked
	time.Sleep(10 * time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(4))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))

	c.Check(consumer.LastDelivery.Payload(), Equals, "del6")
	err = consumer.LastDelivery.Ack()
	c.Check(err, IsNil)
	time.Sleep(10 * time.Millisecond)
	count, err = queue.unackedCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(3))

	queue.StopConsuming()
	conn.stopHeartbeat()
	time.Sleep(time.Millisecond)

	cleanerConn, err := OpenConnection("cleaner-conn", "tcp", "localhost:6379", 1)
	c.Check(err, IsNil)
	cleaner := NewCleaner(cleanerConn)
	c.Check(cleaner.Clean(), IsNil)
	count, err = queue.readyCount()
	c.Check(err, IsNil)
	c.Check(count, Equals, int64(9)) // 2 of 11 were acked above
	queues, err = conn.GetOpenQueues()
	c.Check(err, IsNil)
	c.Check(queues, HasLen, 2)

	conn, err = OpenConnection("cleaner-conn1", "tcp", "localhost:6379", 1)
	c.Check(err, IsNil)
	queue = conn.OpenQueue("q1")
	queue.StartConsuming(10, time.Millisecond)
	consumer = NewTestConsumer("c-C")

	queue.AddConsumer("consumer3", consumer)
	time.Sleep(10 * time.Millisecond)
	c.Check(consumer.LastDeliveries, HasLen, 9)

	queue.StopConsuming()
	conn.stopHeartbeat()
	time.Sleep(time.Millisecond)

	c.Check(cleaner.Clean(), IsNil)
	cleanerConn.stopHeartbeat()
}
