package rmq

import "fmt"

type Cleaner struct {
	connection Connection
}

func NewCleaner(connection Connection) *Cleaner {
	return &Cleaner{connection: connection}
}

func (cleaner *Cleaner) Clean() error {
	cleanerConnection, ok := cleaner.connection.(*redisConnection)
	if !ok {
		return nil
	}
	connectionNames, err := cleanerConnection.GetConnections()
	if err != nil {
		return err
	}

	for _, connectionName := range connectionNames {
		connection := cleanerConnection.hijackConnection(connectionName)
		ok, err := connection.Check()
		if err != nil {
			return err
		}

		if ok {
			continue // skip active connections!
		}

		if err := CleanConnection(connection); err != nil {
			return err
		}
	}

	return nil
}

func CleanConnection(connection *redisConnection) error {
	queueNames, err := connection.GetConsumingQueues()
	if err != nil {
		return err
	}

	for _, queueName := range queueNames {
		queue, ok := connection.OpenQueue(queueName).(*redisQueue)
		// TODO: these would merge if we returned redis.nil
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("rmq cleaner failed to open queue %s", queueName)
		}

		CleanQueue(queue)
	}

	ok, err := connection.Close()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("rmq cleaner failed to close connection %s", connection)
	}

	if err := connection.CloseAllQueuesInConnection(); err != nil {
		return fmt.Errorf("rmq cleaner failed to close all queues %s %s", connection, err)
	}

	// log.Printf("rmq cleaner cleaned connection %s", connection)
	return nil
}

func CleanQueue(queue *redisQueue) error {
	returned, err := queue.ReturnAllUnacked()
	if err != nil {
		return err
	}
	queue.CloseInConnection()
	_ = returned
	// log.Printf("rmq cleaner cleaned queue %s %d", queue, returned)
	return nil
}
