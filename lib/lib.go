package lib

import (
	"context"
	runtime "github.com/hanfei1991/microcosm/executor/worker"
	"github.com/hanfei1991/microcosm/lib/common"
)

type MasterID = common.MasterID
type WorkerID = common.WorkerID
type WorkerType = common.WorkerType

type Master interface {
	Init(ctx context.Context) error
	Poll(ctx context.Context) error
	MasterID() MasterID

	runtime.Closer
}

type MasterImpl interface {
	// InitImpl provides customized logic for the business logic to initialize.
	InitImpl(ctx context.Context) error

	// Tick is called on a fixed interval.
	Tick(ctx context.Context) error

	// OnMasterRecovered is called when the master has recovered from an error.
	OnMasterRecovered(ctx context.Context) error

	// OnWorkerDispatched is called when a request to launch a worker is finished.
	OnWorkerDispatched(worker WorkerHandle, result error) error

	// OnWorkerOnline is called when the first heartbeat for a worker is received.
	OnWorkerOnline(worker WorkerHandle) error

	// OnWorkerOffline is called when a worker exits or has timed out.
	OnWorkerOffline(worker WorkerHandle, reason error) error

	// OnWorkerMessage is called when a customized message is received.
	OnWorkerMessage(worker WorkerHandle, topic p2p.Topic, message interface{}) error

	// CloseImpl is called when the master is being closed
	CloseImpl(ctx context.Context) error
}