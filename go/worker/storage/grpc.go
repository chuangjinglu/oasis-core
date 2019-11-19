package storage

import (
	"context"

	"github.com/pkg/errors"

	"github.com/oasislabs/oasis-core/go/common/cbor"
	"github.com/oasislabs/oasis-core/go/common/crypto/signature"
	"github.com/oasislabs/oasis-core/go/common/grpc"
	pb "github.com/oasislabs/oasis-core/go/grpc/storage"
	"github.com/oasislabs/oasis-core/go/worker/storage/committee"
)

var (
	_ pb.StorageWorkerServer = (*grpcServer)(nil)

	// ErrRuntimeNotFound is the error returned when the called references an unknown runtime.
	ErrRuntimeNotFound = errors.New("worker/storage: runtime not found")
)

type grpcServer struct {
	w *Worker
}

func (s *grpcServer) GetLastSyncedRound(ctx context.Context, req *pb.GetLastSyncedRoundRequest) (*pb.GetLastSyncedRoundResponse, error) {
	var id signature.PublicKey
	if err := id.UnmarshalBinary(req.GetRuntimeId()); err != nil {
		return nil, err
	}

	var node *committee.Node
	node, ok := s.w.runtimes[id]
	if !ok {
		return nil, ErrRuntimeNotFound
	}

	round, ioRoot, stateRoot := node.GetLastSynced()

	resp := &pb.GetLastSyncedRoundResponse{
		Round:     round,
		IoRoot:    cbor.Marshal(ioRoot),
		StateRoot: cbor.Marshal(stateRoot),
	}
	return resp, nil
}

func (s *grpcServer) ForceFinalize(ctx context.Context, req *pb.ForceFinalizeRequest) (*pb.ForceFinalizeResponse, error) {
	var id signature.PublicKey
	if err := id.UnmarshalBinary(req.GetRuntimeId()); err != nil {
		return nil, err
	}

	round := req.GetRound()

	var node *committee.Node
	node, ok := s.w.runtimes[id]
	if !ok {
		return nil, ErrRuntimeNotFound
	}

	if err := node.ForceFinalize(ctx, id, round); err != nil {
		return nil, err
	}

	return &pb.ForceFinalizeResponse{}, nil
}

func newGRPCServer(grpc *grpc.Server, w *Worker) {
	s := &grpcServer{w}
	pb.RegisterStorageWorkerServer(grpc.Server(), s)
}
