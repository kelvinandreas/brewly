//go:build ignore

// Canonical pattern for a domain entity.
// Copy and rename Thing → your aggregate root.
// Rules: no third-party imports except github.com/google/uuid.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ThingStatus represents allowed states of a Thing.
type ThingStatus string

const (
	ThingStatusActive   ThingStatus = "active"
	ThingStatusArchived ThingStatus = "archived"
)

// Thing is a domain entity. No GORM tags — those live in repository/.
type Thing struct {
	ID        uuid.UUID
	Name      string
	Status    ThingStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ThingRepository is the interface that repository/ must implement.
// Usecase depends only on this interface, never on GORM.
type ThingRepository interface {
	Create(thing *Thing) error
	GetByID(id uuid.UUID) (*Thing, error)
	List() ([]*Thing, error)
	Update(thing *Thing) error
	SoftDelete(id uuid.UUID) error
}

// Sentinel errors — handlers map these to HTTP status codes via errors.Is.
var (
	ErrThingNotFound = errors.New("thing not found")
	ErrThingConflict = errors.New("thing already exists")
)
