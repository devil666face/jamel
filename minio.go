package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := ""        // MinIO server from Docker Compose
	accessKeyID := ""     // Set in docker-compose.yml
	secretAccessKey := "" // Set in docker-compose.yml
	useSSL := false       // MinIO in Docker Compose does not use SSL

	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Define bucket and object names
	bucket := "jamel"

	// Create the bucket (optional, checks if it exists)
	ctx := context.Background()
	err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			fmt.Printf("Bucket %s already exists.\n", bucket)
		} else {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	} else {
		fmt.Printf("Bucket %s created successfully.\n", bucket)
	}

	objectName := "ubuntu.tar"
	localFilePath := "ubuntu.tar"
	downloadedFilePath := "ubuntu.2.tar"

	// Upload the file from disk to MinIO
	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		log.Fatalf("Failed to access file %s: %v", localFilePath, err)
	}
	file, err := os.Open(localFilePath)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", localFilePath, err)
	}
	defer file.Close()

	info, err := client.PutObject(ctx, bucket, objectName, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
	fmt.Printf("File %s uploaded successfully to bucket %s as %s.\n", localFilePath, bucket, objectName)
	fmt.Println(info.ETag)

	// Download the file from MinIO and save it to disk
	object, err := client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalf("Failed to retrieve file: %v", err)
	}
	defer object.Close()

	outFile, err := os.Create(downloadedFilePath)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", downloadedFilePath, err)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, object); err != nil {
		log.Fatalf("Failed to save file to disk: %v", err)
	}
	fmt.Printf("File downloaded from bucket %s and saved as %s.\n", bucket, downloadedFilePath)
}
