package providers_testing

import (
	"sync"

	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type TestMessage struct {
	Subject string
	Content string
}

// TestingProvider implements TypeSendProvider for testing purposes.
// It stores envelopes internally in a map keyed by envelope ID.
type TestingProvider struct {
	mu       sync.Mutex
	messages map[string]*TestMessage
	// SendError, if non-nil, forces Deliver to return that error.
	SendError error
}

// NewTestingProvider creates a new instance of TestingProvider.
func NewTestingProvider() *TestingProvider {
	return &TestingProvider{
		messages: make(map[string]*TestMessage),
	}
}

// Deliver stores the provided envelope internally.
// It returns t.SendError if set, otherwise stores the envelope and returns nil.
func (t *TestingProvider) Deliver(e *typesend_schemas.TypeSendEnvelope, filledTemplate *typesend_schemas.TypeSendTemplate) error {
	if t.SendError != nil {
		return t.SendError
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.messages[e.ID] = &TestMessage{
		Subject: filledTemplate.Subject,
		Content: filledTemplate.Content,
	}
	return nil
}

// GetProviderName returns a fixed provider name.
func (t *TestingProvider) GetProviderName() string {
	return "TestingProvider"
}

// GetEnvelopeByID retrieves a stored envelope by its ID.
// Returns nil if no envelope is found.
func (t *TestingProvider) GetMessageByEnvelopeID(id string) *TestMessage {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.messages[id]
}

// ListMessages returns a slice containing all stored messages.
func (t *TestingProvider) ListMessages() []*TestMessage {
	t.mu.Lock()
	defer t.mu.Unlock()

	msgs := make([]*TestMessage, 0, len(t.messages))
	for _, msg := range t.messages {
		msgs = append(msgs, msg)
	}
	return msgs
}
