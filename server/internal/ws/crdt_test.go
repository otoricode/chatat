package ws

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLWWRegister_Merge(t *testing.T) {
	nodeA := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	nodeB := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	t.Run("higher timestamp wins", func(t *testing.T) {
		reg := LWWRegister{Value: "old", Timestamp: 100, NodeID: nodeA}
		accepted := reg.Merge(LWWRegister{Value: "new", Timestamp: 200, NodeID: nodeB})
		assert.True(t, accepted)
		assert.Equal(t, "new", reg.Value)
		assert.Equal(t, int64(200), reg.Timestamp)
	})

	t.Run("lower timestamp loses", func(t *testing.T) {
		reg := LWWRegister{Value: "current", Timestamp: 200, NodeID: nodeA}
		accepted := reg.Merge(LWWRegister{Value: "old", Timestamp: 100, NodeID: nodeB})
		assert.False(t, accepted)
		assert.Equal(t, "current", reg.Value)
	})

	t.Run("same timestamp higher node ID wins", func(t *testing.T) {
		reg := LWWRegister{Value: "a-value", Timestamp: 100, NodeID: nodeA}
		accepted := reg.Merge(LWWRegister{Value: "b-value", Timestamp: 100, NodeID: nodeB})
		assert.True(t, accepted)
		assert.Equal(t, "b-value", reg.Value)
	})

	t.Run("same timestamp lower node ID loses", func(t *testing.T) {
		reg := LWWRegister{Value: "b-value", Timestamp: 100, NodeID: nodeB}
		accepted := reg.Merge(LWWRegister{Value: "a-value", Timestamp: 100, NodeID: nodeA})
		assert.False(t, accepted)
		assert.Equal(t, "b-value", reg.Value)
	})
}

func TestDocumentCRDT_ApplyUpdate(t *testing.T) {
	docID := uuid.New()
	blockID := uuid.New()
	nodeA := uuid.New()
	nodeB := uuid.New()

	t.Run("first update is accepted", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		accepted := crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID,
			BlockID:    blockID,
			Field:      "content",
			Value:      "hello",
			Timestamp:  100,
			NodeID:     nodeA,
		})
		assert.True(t, accepted)
		state := crdt.GetBlockState(blockID)
		require.NotNil(t, state)
		assert.Equal(t, "hello", state.Content.Value)
	})

	t.Run("later update wins over earlier", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "content",
			Value: "first", Timestamp: 100, NodeID: nodeA,
		})
		accepted := crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "content",
			Value: "second", Timestamp: 200, NodeID: nodeB,
		})
		assert.True(t, accepted)
		assert.Equal(t, "second", crdt.GetBlockState(blockID).Content.Value)
	})

	t.Run("earlier update loses to later", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "content",
			Value: "later", Timestamp: 200, NodeID: nodeA,
		})
		accepted := crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "content",
			Value: "earlier", Timestamp: 100, NodeID: nodeB,
		})
		assert.False(t, accepted)
		assert.Equal(t, "later", crdt.GetBlockState(blockID).Content.Value)
	})

	t.Run("update rejected on deleted block", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		crdt.ApplyDelete(CRDTDeleteEvent{
			DocumentID: docID, BlockID: blockID, Timestamp: 100, NodeID: nodeA,
		})
		accepted := crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "content",
			Value: "revive", Timestamp: 200, NodeID: nodeB,
		})
		assert.False(t, accepted)
	})

	t.Run("concurrent edits different blocks", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		block1 := uuid.New()
		block2 := uuid.New()
		crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: block1, Field: "content",
			Value: "block1-by-A", Timestamp: 100, NodeID: nodeA,
		})
		crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: block2, Field: "content",
			Value: "block2-by-B", Timestamp: 100, NodeID: nodeB,
		})
		assert.Equal(t, "block1-by-A", crdt.GetBlockState(block1).Content.Value)
		assert.Equal(t, "block2-by-B", crdt.GetBlockState(block2).Content.Value)
	})

	t.Run("checked field update", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		accepted := crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "checked",
			Value: "true", Timestamp: 100, NodeID: nodeA,
		})
		assert.True(t, accepted)
		assert.Equal(t, "true", crdt.GetBlockState(blockID).Checked.Value)
	})

	t.Run("unknown field rejected", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		accepted := crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "unknown",
			Value: "val", Timestamp: 100, NodeID: nodeA,
		})
		assert.False(t, accepted)
	})
}

func TestDocumentCRDT_ApplyDelete(t *testing.T) {
	docID := uuid.New()
	blockID := uuid.New()
	nodeA := uuid.New()

	t.Run("delete marks block as deleted", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		crdt.ApplyUpdate(CRDTUpdateEvent{
			DocumentID: docID, BlockID: blockID, Field: "content",
			Value: "hello", Timestamp: 100, NodeID: nodeA,
		})
		accepted := crdt.ApplyDelete(CRDTDeleteEvent{
			DocumentID: docID, BlockID: blockID, Timestamp: 200, NodeID: nodeA,
		})
		assert.True(t, accepted)
		assert.True(t, crdt.GetBlockState(blockID).Deleted)
	})

	t.Run("duplicate delete is rejected", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		crdt.ApplyDelete(CRDTDeleteEvent{
			DocumentID: docID, BlockID: blockID, Timestamp: 200, NodeID: nodeA,
		})
		accepted := crdt.ApplyDelete(CRDTDeleteEvent{
			DocumentID: docID, BlockID: blockID, Timestamp: 100, NodeID: nodeA,
		})
		assert.False(t, accepted)
	})

	t.Run("delete on nonexistent block creates it deleted", func(t *testing.T) {
		crdt := NewDocumentCRDT(docID)
		newBlock := uuid.New()
		accepted := crdt.ApplyDelete(CRDTDeleteEvent{
			DocumentID: docID, BlockID: newBlock, Timestamp: 100, NodeID: nodeA,
		})
		assert.True(t, accepted)
		state := crdt.GetBlockState(newBlock)
		require.NotNil(t, state)
		assert.True(t, state.Deleted)
	})
}

func TestDocumentCRDT_Clock(t *testing.T) {
	docID := uuid.New()
	crdt := NewDocumentCRDT(docID)

	t.Run("tick advances clock", func(t *testing.T) {
		ts1 := crdt.Tick()
		ts2 := crdt.Tick()
		assert.True(t, ts2 > ts1)
	})

	t.Run("receive tick updates clock", func(t *testing.T) {
		farFuture := int64(9999999999999)
		ts := crdt.ReceiveTick(farFuture)
		assert.True(t, ts > farFuture)
	})

	t.Run("tick else branch when clock is ahead of now", func(t *testing.T) {
		// Set clock far into the future so time.Now().UnixMilli() < Clock
		crdt2 := NewDocumentCRDT(uuid.New())
		crdt2.Clock = int64(99999999999999) // far future
		prev := crdt2.Clock
		ts := crdt2.Tick()
		assert.Equal(t, prev+1, ts, "should increment clock by 1 when now <= clock")
	})
}

func TestDocumentCRDTManager(t *testing.T) {
	manager := NewDocumentCRDTManager()
	docID := uuid.New()

	t.Run("get or create", func(t *testing.T) {
		crdt := manager.GetOrCreate(docID)
		require.NotNil(t, crdt)
		assert.Equal(t, docID, crdt.DocumentID)
		crdt2 := manager.GetOrCreate(docID)
		assert.Equal(t, crdt, crdt2)
	})

	t.Run("has", func(t *testing.T) {
		assert.True(t, manager.Has(docID))
		assert.False(t, manager.Has(uuid.New()))
	})

	t.Run("remove", func(t *testing.T) {
		manager.Remove(docID)
		assert.False(t, manager.Has(docID))
	})
}
