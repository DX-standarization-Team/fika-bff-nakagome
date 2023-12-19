package logger

import (
	"context"
	"log"

	"cloud.google.com/go/logging"
)

func NewLogger() *logging.Logger {

	log.Println("NewLogger entering")
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "kaigofika-poc01"

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// defer client.Close()
	log.Printf("client: %v", client)

	// Sets the name of the log to write to.
	logName := "my-log"
	// Selects the log to write to.
	logger := client.Logger(logName)
	log.Printf("logger: %v", logger)
	// defer logger.Flush()

	return logger
}
