package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"jamel/gen/go/jamel"
	"jamel/pkg/rmq"

	"github.com/streadway/amqp"
)

type S3 interface {
	Download(string) (string, error)
}

type Rmq interface {
	Publish(string, []byte) error
	Consume(context.Context, string, chan<- amqp.Delivery) error
}

type Cve interface {
	Get(string, string) ([]byte, error)
}

type Client struct {
	s3  S3
	rmq Rmq
	cve Cve
}

func Must(
	_s3 S3,
	_rmq Rmq,
	_cve Cve,
) *Client {
	return &Client{
		s3:  _s3,
		rmq: _rmq,
		cve: _cve,
	}
}

var TaskTypeMap = map[jamel.TaskType]string{
	jamel.TaskType_DOCKER:         "docker",
	jamel.TaskType_DOCKER_ARCHIVE: "docker-archive",
	jamel.TaskType_DIR:            "dir",
	jamel.TaskType_SBOM:           "sbom",
}

func (c *Client) Run() error {
	log.Println("loop started")
	defer log.Println("loop stopped")

	var (
		taskch      = make(chan amqp.Delivery)
		errch       = make(chan error)
		ctx, cancel = context.WithCancel(context.Background())
	)
	defer close(taskch)
	defer close(errch)
	defer cancel()

	if err := c.rmq.Consume(ctx, rmq.TaskQueue, taskch); err != nil {
		return fmt.Errorf("failed to consume queue: %w", err)
	}

	go func() {
		for data := range taskch {
			func() {
				var task = jamel.TaskResponse{}
				if err := json.Unmarshal(data.Body, &task); err != nil {
					errch <- fmt.Errorf("unmarshal task from queue error: %w", err)
					return
				}
				if _, err := c.s3.Download(task.TaskId); err != nil {
					errch <- fmt.Errorf("download from s3 error: %w", err)
					return
				}
				defer func() {
					if err := os.Remove(task.TaskId); err != nil {
						// errch <- fmt.Errorf("failed to remove: %w", err)
						return
					}
				}()
				out, err := c.cve.Get(TaskTypeMap[task.TaskType], task.TaskId)
				if err != nil {
					errch <- fmt.Errorf("getting cves error: %w", err)
					return
				}
				task.Report = string(out)
				data, err := json.Marshal(&task)
				if err != nil {
					errch <- fmt.Errorf("failed to marshal result before set in queue: %w", err)
					return
				}
				if err := c.rmq.Publish(rmq.ResultQueue, data); err != nil {
					errch <- fmt.Errorf("failed to set in result queue: %w", err)
					return
				}
			}()

		}
	}()

	for err := range errch {
		if err != nil {
			return fmt.Errorf("task queue error: %w", err)
		}
	}
	return nil
}
