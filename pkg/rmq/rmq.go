package rmq

import (
	"fmt"
	"sync"

	"github.com/streadway/amqp"
)

const (
	TaskQueue   = "jamel_task"
	ResultQueue = "jamel_result"
)

type queueset struct {
	set map[string]amqp.Queue
	s   sync.RWMutex
}

func (q *queueset) Set(name string, queue amqp.Queue) {
	q.s.Lock()
	defer q.s.Unlock()
	q.set[name] = queue
}

func (q *queueset) Get(name string) (amqp.Queue, error) {
	q.s.RLock()
	defer q.s.RUnlock()
	queue, ok := q.set[name]
	if !ok {
		return amqp.Queue{}, fmt.Errorf("failed to get queue")
	}
	return queue, nil
}

type Rmq struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queueset queueset
}

func New(
	connect string,
	username, password string,
	queue ...string,
) (_rmq *Rmq, err error) {
	if len(queue) == 0 {
		return nil, fmt.Errorf("compile error: queue names not set")
	}
	_rmq = &Rmq{
		queueset: queueset{set: make(map[string]amqp.Queue)},
	}
	_rmq.conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", username, password, connect))
	if err != nil {
		return nil, fmt.Errorf("failed to amqp connect: %w", err)
	}
	_rmq.channel, err = _rmq.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	for _, name := range queue {
		queue, err := _rmq.channel.QueueDeclare(
			name,  // Queue name
			true,  // Durable
			false, // Delete when unused
			false, // Exclusive
			false, // No-wait
			nil,   // Arguments
		)
		if err != nil {
			return nil, fmt.Errorf("failed to make queue: %w", err)
		}
		_rmq.queueset.Set(name, queue)
	}
	return _rmq, nil
}

func (r *Rmq) Close() (err error) {
	err = r.conn.Close()
	err = r.channel.Close()
	return
}

func (r *Rmq) Publish(queuename string, body []byte) error {
	queue, err := r.queueset.Get(queuename)
	if err != nil {
		return fmt.Errorf("get queue error: %w", err)
	}
	if err := r.channel.Publish(
		"",         // Default exchange
		queue.Name, // Routing key (queue name)
		true,       // Mandatory
		false,      // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	); err != nil {
		return fmt.Errorf("publish error: %w", err)
	}
	return nil
}

func (r *Rmq) Consume(queuename string, messagechan chan<- amqp.Delivery) error {
	queue, err := r.queueset.Get(queuename)
	if err != nil {
		return fmt.Errorf("get queue error: %w", err)
	}
	msgs, err := r.channel.Consume(
		queue.Name, // Queue name
		"",         // Consumer name
		true,       // Auto-ack
		false,      // Exclusive
		false,      // No-local
		false,      // No-wait
		nil,        // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to get messages chan: %w", err)
	}
	go func() {
		for msg := range msgs {
			messagechan <- msg // Send the message to the channel
		}
	}()
	return nil
}
