package client

import (
	"context"
	"encoding/json"
	"fmt"
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
	Connect() error
}

type Cve interface {
	GetUnwrap(string, string) (string, string, string, error)
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
	jamel.TaskType_SBOM:           "sbom",
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

	if err := c.rmq.Consume(ctx, rmq.TaskQueue, taskch); err != nil {
		return fmt.Errorf("failed to consume queue: %w", err)
	}

	go func() {
		for data := range taskch {
			func() {
				var (
					resp = jamel.TaskResponse{}
					err  error
				)
				if err := json.Unmarshal(data.Body, &resp); err != nil {
					errch <- fmt.Errorf("unmarshal task from queue error: %w", err)
					return
				}
				switch resp.TaskType {
				case jamel.TaskType_DOCKER:
					if resp.Json, resp.Report, resp.Sbom, err = c.NewTaskFromImage(&resp); err != nil {
						resp.Error = err.Error()
					}
				default:
					if resp.Json, resp.Report, resp.Sbom, err = c.NewTaskFromFile(&resp); err != nil {
						resp.Error = err.Error()
					}
				}
				data, err := json.Marshal(&resp)
				if err != nil {
					resp.Error = fmt.Errorf("failed to marshal result before set in queue: %w", err).Error()
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

func (c *Client) NewTaskFromFile(task *jamel.TaskResponse) (string, string, string, error) {
	if _, err := c.s3.Download(task.TaskId); err != nil {
		return "", "", "", fmt.Errorf("download from s3 error: %w", err)
	}
	defer func() {
		if err := os.Remove(task.TaskId); err != nil {
			return
		}
	}()
	return c.cve.GetUnwrap(TaskTypeMap[task.TaskType], task.TaskId)
}

func (c *Client) NewTaskFromImage(task *jamel.TaskResponse) (string, string, string, error) {
	return c.cve.GetUnwrap(TaskTypeMap[task.TaskType], task.Name)
}

func (c *Client) Reconnect() error {
	return c.rmq.Connect()
}
