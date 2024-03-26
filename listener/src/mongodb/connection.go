package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/dappnode/validator-monitoring/listener/src/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB(uri string) (*mongo.Client, error) {
	// The URI includes the credentials
	var client *mongo.Client
	var err error // Declare err here to ensure it's accessible outside the loop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for attempt := 1; attempt <= 5; attempt++ {
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			logger.Warn(fmt.Sprintf("Attempt %d: Failed to initiate connection to MongoDB: %v", attempt, err))
			time.Sleep(time.Second * 5) // Wait for 5 seconds before retrying
			continue                    // Proceed to the next iteration of the loop
		}

		// Attempt to ping the database
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = client.Ping(pingCtx, nil)
		pingCancel()

		if err == nil {
			logger.Info("Connected to MongoDB!")
			return client, nil // Successfully connected and pinged MongoDB
		}

		logger.Warn(fmt.Sprintf("Attempt %d: Failed to ping MongoDB: %v", attempt, err))
		time.Sleep(time.Second * 5) // Wait for 5 seconds before retrying
	}

	// After exiting the loop, either connection initiation or ping has consistently failed
	return nil, fmt.Errorf("after several attempts, failed to connect to MongoDB: %v", err) // Keeping fmt for error return
}
