package committee

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/oasislabs/oasis-core/go/common/crypto/hash"
	"github.com/oasislabs/oasis-core/go/common/crypto/signature"
	roothash "github.com/oasislabs/oasis-core/go/roothash/api"
	"github.com/oasislabs/oasis-core/go/roothash/api/block"
	"github.com/oasislabs/oasis-core/go/runtime/transaction"
	"github.com/oasislabs/oasis-core/go/worker/common/host/protocol"
)

// StateName is a symbolic state without the attached values.
type StateName string

const (
	// NotReady is the name of StateNotReady.
	NotReady StateName = "NotReady"
	// WaitingForBatch is the name of StateWaitingForBatch.
	WaitingForBatch = "WaitingForBatch"
	// WaitingForBlock is the name of StateWaitingForBlock.
	WaitingForBlock = "WaitingForBlock"
	// WaitingForEvent is the name of StateWaitingForEvent.
	WaitingForEvent = "WaitingForEvent"
	// ProcessingBatch is the name of StateProcessingBatch.
	ProcessingBatch = "ProcessingBatch"
	// WaitingForFinalize is the name of StateWaitingForFinalize.
	WaitingForFinalize = "WaitingForFinalize"
)

// Valid state transitions.
var validStateTransitions = map[StateName][]StateName{
	// Transitions from NotReady state.
	NotReady: {
		// Epoch transition occurred and we are not in the committee.
		NotReady,
		// Epoch transition occurred and we are in the committee.
		WaitingForBatch,
	},

	// Transitions from WaitingForBatch state.
	WaitingForBatch: {
		WaitingForBatch,
		// Received batch, need to catch up current block.
		WaitingForBlock,
		// Received batch, current block is up to date.
		ProcessingBatch,
		// Received batch, waiting for disrepancy event.
		WaitingForEvent,
		// Epoch transition occurred and we are no longer in the committee.
		NotReady,
	},

	// Transitions from WaitingForBlock state.
	WaitingForBlock: {
		// Abort: seen newer block while waiting for block.
		WaitingForBatch,
		// Seen block that we were waiting for.
		ProcessingBatch,
		// Seen block that we were waiting for, waiting for disrepancy event.
		WaitingForEvent,
		// Epoch transition occurred and we are no longer in the committee.
		NotReady,
	},

	// Transitions from WaitingForEvent state.
	WaitingForEvent: {
		// Abort: seen newer block while waiting for event.
		WaitingForBatch,
		// Discrepancy event received.
		ProcessingBatch,
		// Epoch transition occurred and we are no longer in the committee.
		NotReady,
	},

	// Transitions from ProcessingBatch state.
	ProcessingBatch: {
		// Batch has been successfully processed or has been aborted.
		WaitingForFinalize,
	},

	// Transitions from WaitingForFinalize state.
	WaitingForFinalize: {
		// Round has been finalized.
		WaitingForBatch,
		// Epoch transition occurred and we are no longer in the committee.
		NotReady,
	},
}

// NodeState is a node's state.
type NodeState interface {
	// Name returns the name of the state.
	Name() StateName
}

// StateNotReady is the not ready state.
type StateNotReady struct {
}

// Name returns the name of the state.
func (s StateNotReady) Name() StateName {
	return NotReady
}

// String returns a string representation of the state.
func (s StateNotReady) String() string {
	return string(s.Name())
}

// StateWaitingForBatch is the waiting for batch state.
type StateWaitingForBatch struct {
	// Pending execute discrepancy detected event in case the node is a
	// backup worker and the event was received before the batch.
	pendingEvent *roothash.ExecutionDiscrepancyDetectedEvent
}

// Name returns the name of the state.
func (s StateWaitingForBatch) Name() StateName {
	return WaitingForBatch
}

// String returns a string representation of the state.
func (s StateWaitingForBatch) String() string {
	return string(s.Name())
}

// StateWaitingForBlock is the waiting for block state.
type StateWaitingForBlock struct {
	// I/O root from the transaction scheduler containing the inputs.
	ioRoot hash.Hash
	// Batch that is waiting to be processed.
	batch transaction.RawBatch
	// Tracing for this batch.
	batchSpanCtx opentracing.SpanContext
	// Header of the block we are waiting for.
	header *block.Header
	// Transaction scheduler's signature.
	txnSchedSig signature.Signature
	// Storage signatures for the I/O root containing the inputs.
	inputStorageSigs []signature.Signature
}

// Name returns the name of the state.
func (s StateWaitingForBlock) Name() StateName {
	return WaitingForBlock
}

// String returns a string representation of the state.
func (s StateWaitingForBlock) String() string {
	return string(s.Name())
}

// StateWaitingForEvent is the waiting for event state.
type StateWaitingForEvent struct {
	// I/O root from the transaction scheduler containing the inputs.
	ioRoot hash.Hash
	// Batch that is being processed.
	batch transaction.RawBatch
	// Tracing for this batch.
	batchSpanCtx opentracing.SpanContext
	// Transaction scheduler's signature.
	txnSchedSig signature.Signature
	// Storage signatures for the I/O root containing the inputs.
	inputStorageSigs []signature.Signature
}

// Name returns the name of the state.
func (s StateWaitingForEvent) Name() StateName {
	return WaitingForEvent
}

// String returns a string representation of the state.
func (s StateWaitingForEvent) String() string {
	return string(s.Name())
}

// StateProcessingBatch is the processing batch state.
type StateProcessingBatch struct {
	// I/O root from the transaction scheduler containing the inputs.
	ioRoot hash.Hash
	// Batch that is being processed.
	batch transaction.RawBatch
	// Tracing for this batch.
	batchSpanCtx opentracing.SpanContext
	// Timing for this batch.
	batchStartTime time.Time
	// Function for cancelling batch processing.
	cancelFn context.CancelFunc
	// Channel which will provide the result.
	done chan *protocol.ComputedBatch
	// Transaction scheduler's signature.
	txnSchedSig signature.Signature
	// Storage signatures for the I/O root containing the inputs.
	inputStorageSigs []signature.Signature
}

// Name returns the name of the state.
func (s StateProcessingBatch) Name() StateName {
	return ProcessingBatch
}

// String returns a string representation of the state.
func (s StateProcessingBatch) String() string {
	return string(s.Name())
}

func (s *StateProcessingBatch) cancel() {
	// Invoke the cancellation function and wait for the processing
	// to actually stop.
	(s.cancelFn)()
	<-s.done
}

// StateWaitingForFinalize is the waiting for finalize state.
type StateWaitingForFinalize struct {
	batchStartTime time.Time
}

// Name returns the name of the state.
func (s StateWaitingForFinalize) Name() StateName {
	return WaitingForFinalize
}

// String returns a string representation of the state.
func (s StateWaitingForFinalize) String() string {
	return string(s.Name())
}