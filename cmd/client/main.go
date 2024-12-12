package main

import (
	"log"

	"jamel/internal/client"
	"jamel/internal/client/config"
	"jamel/pkg/cve"
	"jamel/pkg/rmq"
	"jamel/pkg/s3"
)

func main() {
	_config := config.Must()

	_s3, err := s3.New(
		_config.S3Conntect,
		_config.S3Username, _config.S3Password,
		_config.S3Bucket,
	)
	if err != nil {
		log.Fatalf("failed to create s3 obj: %v", err)
	}

	_rmq, err := rmq.New(
		_config.RMQConnect,
		_config.RMQUsername, _config.RMQPassword,
		rmq.TaskQueue, rmq.ResultQueue,
	)
	if err != nil {
		log.Fatalf("failed rmq connect: %v", err)
	}

	_cve, err := cve.New()
	if err != nil {
		log.Fatalf("failed to create cve checker: %v", err)
	}

	_client := client.Must(
		_s3,
		_rmq,
		_cve,
	)
	for {
		log.Println("loop started")
		if err := _client.Run(); err != nil {
			log.Printf("critical queue runtime error: %v", err)
		}
		if err := _client.Reconnect(); err != nil {
			log.Printf("failed to reconnect rmq: %v", err)
		}
	}
}
