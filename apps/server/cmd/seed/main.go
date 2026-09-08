// Seed command — creates the single Reliva user and default contexts.
// Run once via Docker Compose; exits immediately if a user already exists.
//
// Required env vars: MONGO_URI, SEED_EMAIL, SEED_PASSWORD
// Optional env vars: DB_NAME (default: reliva)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load .env if present (useful for local runs; Docker injects env vars directly).
	_ = godotenv.Load()

	uri      := mustEnv("MONGO_URI")
	dbName   := envOr("DB_NAME", "reliva")
	email    := mustEnv("SEED_EMAIL")
	password := mustEnv("SEED_PASSWORD")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}

	db := client.Database(dbName)
	users := db.Collection("users")

	// Skip if a user already exists.
	count, err := users.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalf("count users: %v", err)
	}
	if count > 0 {
		fmt.Println("User already exists — skipping seed.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	_, err = users.InsertOne(ctx, bson.M{
		"email":               email,
		"password_hash":       string(hash),
		"totp_secret":         "",
		"totp_enabled":        false,
		"totp_pending_secret": "",
		"backup_codes":        []string{},
		"created_at":          time.Now(),
	})
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}

	_, err = db.Collection("contexts").InsertMany(ctx, []interface{}{
		bson.M{"name": "Personal", "slug": "personal", "color": "#6366f1", "icon": "user",      "type": "personal", "order": 0, "created_at": time.Now(), "updated_at": time.Now()},
		bson.M{"name": "Work",     "slug": "work",     "color": "#f59e0b", "icon": "briefcase", "type": "work",     "order": 1, "created_at": time.Now(), "updated_at": time.Now()},
		bson.M{"name": "Health",   "slug": "health",   "color": "#10b981", "icon": "heart",     "type": "health",   "order": 2, "created_at": time.Now(), "updated_at": time.Now()},
	})
	if err != nil {
		log.Fatalf("insert contexts: %v", err)
	}

	fmt.Printf("✓ User created: %s\n", email)
	fmt.Println("✓ Default contexts: Personal, Work, Health")
	fmt.Println("Next: log in and set up TOTP in /settings.")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
