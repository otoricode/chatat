package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// LWWRegister implements a Last-Writer-Wins Register CRDT.
// Each block's content is a register that keeps the value with the highest timestamp.
type LWWRegister struct {
	Value     string    `json:"value"`
	Timestamp int64     `json:"timestamp"` // Lamport timestamp (hybrid logical clock)
	NodeID    uuid.UUID `json:"nodeId"`    // Tie-breaker: higher UUID wins
}

// Merge merges a remote register value using LWW semantics.
// Returns true if the remote value was accepted (wins).
func (r *LWWRegister) Merge(remote LWWRegister) bool {
	if remote.Timestamp > r.Timestamp {
		r.Value = remote.Value
		r.Timestamp = remote.Timestamp
		r.NodeID = remote.NodeID
		return true
	}
	if remote.Timestamp == r.Timestamp {
		// Tie-break: higher node ID wins
		if remote.NodeID.String() > r.NodeID.String() {
			r.Value = remote.Value
			r.Timestamp = remote.Timestamp
			r.NodeID = remote.NodeID
			return true
		}
	}
	return false
}

// BlockCRDT tracks the CRDT state for a single block.
type BlockCRDT struct {
	BlockID   uuid.UUID   `json:"blockId"`
	Content   LWWRegister `json:"content"`
	Checked   LWWRegister `json:"checked"`
	Deleted   bool        `json:"deleted"`
	DeletedAt int64       `json:"deletedAt"`
	DeletedBy uuid.UUID   `json:"deletedBy"`
}

// DocumentCRDT manages CRDT state for an entire document.
// It maintains a map of block CRDTs and provides merge operations.
type DocumentCRDT struct {
	DocumentID uuid.UUID                `json:"documentId"`
	Blocks     map[uuid.UUID]*BlockCRDT `json:"blocks"`
	Clock      int64                    `json:"clock"` // Lamport clock for this document
	mu         sync.RWMutex
}

// NewDocumentCRDT creates a new CRDT state for a document.
func NewDocumentCRDT(docID uuid.UUID) *DocumentCRDT {
	return &DocumentCRDT{
		DocumentID: docID,
		Blocks:     make(map[uuid.UUID]*BlockCRDT),
		Clock:      time.Now().UnixMilli(),
	}
}

// Tick advances the Lamport clock and returns the new timestamp.
func (d *DocumentCRDT) Tick() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now().UnixMilli()
	if now > d.Clock {
		d.Clock = now
	} else {
		d.Clock++
	}
	return d.Clock
}

// ReceiveTick updates the clock based on a received timestamp (Lamport clock rule).
func (d *DocumentCRDT) ReceiveTick(remoteTs int64) int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	if remoteTs > d.Clock {
		d.Clock = remoteTs
	}
	d.Clock++
	return d.Clock
}

// CRDTUpdateEvent represents a CRDT update from a client.
type CRDTUpdateEvent struct {
	DocumentID uuid.UUID `json:"documentId"`
	BlockID    uuid.UUID `json:"blockId"`
	Field      string    `json:"field"` // "content", "checked"
	Value      string    `json:"value"`
	Timestamp  int64     `json:"timestamp"`
	NodeID     uuid.UUID `json:"nodeId"`
}

// CRDTDeleteEvent represents a block deletion via CRDT.
type CRDTDeleteEvent struct {
	DocumentID uuid.UUID `json:"documentId"`
	BlockID    uuid.UUID `json:"blockId"`
	Timestamp  int64     `json:"timestamp"`
	NodeID     uuid.UUID `json:"nodeId"`
}

// ApplyUpdate merges a block field update into the document CRDT.
// Returns true if the update was accepted (remote wins).
func (d *DocumentCRDT) ApplyUpdate(event CRDTUpdateEvent) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	block, exists := d.Blocks[event.BlockID]
	if !exists {
		block = &BlockCRDT{BlockID: event.BlockID}
		d.Blocks[event.BlockID] = block
	}

	if block.Deleted {
		return false
	}

	remote := LWWRegister{
		Value:     event.Value,
		Timestamp: event.Timestamp,
		NodeID:    event.NodeID,
	}

	switch event.Field {
	case "content":
		return block.Content.Merge(remote)
	case "checked":
		return block.Checked.Merge(remote)
	default:
		return false
	}
}

// ApplyDelete marks a block as deleted in the CRDT.
// Returns true if the delete was accepted.
func (d *DocumentCRDT) ApplyDelete(event CRDTDeleteEvent) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	block, exists := d.Blocks[event.BlockID]
	if !exists {
		block = &BlockCRDT{BlockID: event.BlockID}
		d.Blocks[event.BlockID] = block
	}

	if block.Deleted && event.Timestamp <= block.DeletedAt {
		return false
	}

	block.Deleted = true
	block.DeletedAt = event.Timestamp
	block.DeletedBy = event.NodeID
	return true
}

// GetBlockState returns the current CRDT state for a block.
func (d *DocumentCRDT) GetBlockState(blockID uuid.UUID) *BlockCRDT {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Blocks[blockID]
}

// DocumentCRDTManager manages CRDT instances for multiple documents.
type DocumentCRDTManager struct {
	documents map[uuid.UUID]*DocumentCRDT
	mu        sync.RWMutex
}

// NewDocumentCRDTManager creates a new CRDT manager.
func NewDocumentCRDTManager() *DocumentCRDTManager {
	return &DocumentCRDTManager{
		documents: make(map[uuid.UUID]*DocumentCRDT),
	}
}

// GetOrCreate returns the CRDT for a document, creating one if needed.
func (m *DocumentCRDTManager) GetOrCreate(docID uuid.UUID) *DocumentCRDT {
	m.mu.Lock()
	defer m.mu.Unlock()

	if crdt, ok := m.documents[docID]; ok {
		return crdt
	}

	crdt := NewDocumentCRDT(docID)
	m.documents[docID] = crdt
	return crdt
}

// Remove removes a document CRDT (e.g., when all clients leave).
func (m *DocumentCRDTManager) Remove(docID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.documents, docID)
}

// Has checks if a document CRDT exists.
func (m *DocumentCRDTManager) Has(docID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.documents[docID]
	return ok
}
