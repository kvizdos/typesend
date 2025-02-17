package typesend_db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

// TestDatabase is an in-memory implementation of TypeSendDatabase.
type TestDatabase struct {
	mu        sync.Mutex
	connected bool
	items     []*typesend_schemas.TypeSendEnvelope
}

// Connect simply sets the database as connected.
func (db *TestDatabase) Connect(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.connected = true
	db.items = make([]*typesend_schemas.TypeSendEnvelope, 0)
	return nil
}

func (db *TestDatabase) Items() []*typesend_schemas.TypeSendEnvelope {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.items
}

func (db *TestDatabase) GetEnvelopeByID(id string) *typesend_schemas.TypeSendEnvelope {
	for _, item := range db.items {
		if item.ID == id {
			return item
		}
	}
	return nil
}

func (db *TestDatabase) Insert(envelope *typesend_schemas.TypeSendEnvelope) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.connected {
		return fmt.Errorf("database not connected")
	}

	db.items = append(db.items, envelope)
	return nil
}

func (db *TestDatabase) GetMessagesReadyToSend(ctx context.Context, timestamp time.Time) (chan *typesend_schemas.TypeSendEnvelope, error) {
	ch := make(chan *typesend_schemas.TypeSendEnvelope)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
			return
		default:
		}
		for _, envelope := range db.Items() {
			if envelope.Status == typesend_schemas.TypeSendStatus_UNSENT && !envelope.ScheduledFor.After(timestamp) {
				select {
				case <-ctx.Done():
					return
				case ch <- envelope:
				}
			}
		}
	}()

	return ch, nil
}

func (db *TestDatabase) UpdateEnvelopeStatus(ctx context.Context, envelopeID string, toStatus typesend_schemas.TypeSendStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Iterate over the items to find the envelope with the matching ID.
	for _, envelope := range db.items {
		if envelope.ID == envelopeID {
			envelope.Status = toStatus
			return nil
		}
	}

	return fmt.Errorf("envelope with ID %s not found", envelopeID)
}
