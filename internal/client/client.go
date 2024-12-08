package client

import (
	"context"
	"encoding/json"
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/pkg/cve"
	"jamel/pkg/rmq"
	"os"

	"github.com/streadway/amqp"
)

type S3 interface {
	Upload(string) (string, error)
	Download(string) (string, error)
}

type Rmq interface {
	Publish(string, []byte) error
	Consume(string, chan<- amqp.Delivery) error
}

type Client struct {
	s3  S3
	rmq Rmq
}

func Must(
	_s3 S3,
	_rmq Rmq,
) *Client {
	return &Client{
		s3:  _s3,
		rmq: _rmq,
	}
}

func (c *Client) Run() error {
	var (
		taskch      = make(chan amqp.Delivery)
		errch       = make(chan error)
		ctx, cancel = context.WithCancel(context.Background())
	)
	defer close(taskch)
	defer close(errch)
	defer cancel()

	if err := c.rmq.Consume(rmq.TaskQueue, taskch); err != nil {
		return fmt.Errorf("failed to consume queue: %w", err)
	}

	go func() {
		defer close(errch) // Ensure the error channel is closed when the goroutine exits
		for data := range taskch {
			var task jamel.TaskResponse
			if err := json.Unmarshal(data.Body, &task); err != nil {
				errch <- fmt.Errorf("unmarshal task from queue error: %w", err)
				cancel()
				return
			}
			fmt.Println(task)
			if _, err := c.s3.Download(task.TaskId); err != nil {
				errch <- fmt.Errorf("download from s3 error: %w", err)
				cancel()
				return
			}
			out, err := cve.Get(fmt.Sprintf("docker-archive:%s", task.TaskId))
			if err != nil {
				errch <- fmt.Errorf("getting cves error: %w", err)
			}
			fmt.Println(string(out))
			os.Remove(task.TaskId)

		}
	}()

	select {
	case <-ctx.Done():
		return <-errch
	case err := <-errch:
		return err
	}
}
