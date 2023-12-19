package logger

import (
	"context"
	"log"

	"cloud.google.com/go/logging"
)

func NewLogger() *logging.Logger {

	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "kaigofika-poc01"

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Sets the name of the log to write to.
	logName := "my-log"
	// Selects the log to write to.
	logger := client.Logger(logName)
	defer logger.Flush()

	return logger
}
