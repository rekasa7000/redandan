package handlers

import (
	"go.mongodb.org/mongo-driver/v2/mongo"

	"reliva/server/internal/config"
)

// Handler holds shared dependencies injected into every route handler.
// All handler methods are defined on this struct.
type Handler struct {
	db  *mongo.Database
	cfg *config.Config
}

// New creates a Handler with the given database and config.
func New(db *mongo.Database, cfg *config.Config) *Handler {
	return &Handler{db: db, cfg: cfg}
}

// col is a shorthand for getting a MongoDB collection by name.
func (h *Handler) col(name string) *mongo.Collection {
	return h.db.Collection(name)
}
