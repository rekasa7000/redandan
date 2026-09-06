package db

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	client  *mongo.Client
	once    sync.Once
	initErr error
)

// Connect establishes a singleton MongoDB connection.
// Safe to call multiple times — only connects once.
func Connect(uri string) (*mongo.Client, error) {
	once.Do(func() {
		c, err := mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			initErr = err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err = c.Ping(ctx, nil); err != nil {
			initErr = err
			return
		}

		client = c
	})
	return client, initErr
}
