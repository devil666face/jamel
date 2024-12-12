package rmq

import (
	"context"
	"fmt"
	"log"
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

func (q *queueset) Delete() {
	q.s.Lock()
	defer q.s.Unlock()
	for key := range q.set {
		delete(q.set, key)
	}
}

type Rmq struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queueset queueset

	username, password, connect string
	queue                       []string
}

func New(
	_connect string,
	_username, _password string,
	_queue ...string,
) (*Rmq, error) {
	if len(_queue) == 0 {
		return nil, fmt.Errorf("compile error: queue names not set")
	}
	_rmq := &Rmq{
		queueset: queueset{set: make(map[string]amqp.Queue)},
		username: _username,
		password: _password,
		connect:  _connect,
		queue:    _queue,
	}
	if err := _rmq.Connect(); err != nil {
		return nil, fmt.Errorf("connect error: %w", err)
	}
	return _rmq, nil
}

func (r *Rmq) Connect() error {
	var err error
	r.conn, err = amqp.Dial(
		fmt.Sprintf("amqp://%s:%s@%s/", r.username, r.password, r.connect),
	)
	if err != nil {
		return fmt.Errorf("failed to amqp connect: %w", err)
	}
	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	for _, name := range r.queue {
		queue, err := r.channel.QueueDeclare(
			name,  // Queue name
			true,  // Durable
			false, // Delete when unused
			false, // Exclusive
			false, // No-wait
			nil,   // Arguments
		)
		if err != nil {
			return fmt.Errorf("failed to make queue: %w", err)
		}
		r.queueset.Set(name, queue)
	}
	return nil
}

func (r *Rmq) Close() error {
	r.queueset.Delete()
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close: %w", err)
	}
	return nil
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
	log.Printf("set in %s: %v\n", queuename, string(body))
	return nil
}

func (r *Rmq) Consume(ctx context.Context, queuename string, messagechan chan<- amqp.Delivery) error {
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
		// defer r.channel.Close()
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				log.Printf("get from %s: %v\n", queuename, string(msg.Body))
				messagechan <- msg
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
