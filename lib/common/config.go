package common

import "time"

type TimeoutConfig struct {
	WorkerTimeoutDuration            time.Duration
	WorkerTimeoutGracefulDuration    time.Duration
	WorkerHeartbeatInterval          time.Duration
	WorkerReportStatusInterval       time.Duration
	MasterHeartbeatCheckLoopInterval time.Duration
}

var DefaultTimeoutConfig TimeoutConfig = TimeoutConfig{
	WorkerTimeoutDuration:            time.Second * 15,
	WorkerTimeoutGracefulDuration:    time.Second * 5,
	WorkerHeartbeatInterval:          time.Second * 3,
	WorkerReportStatusInterval:       time.Second * 3,
	MasterHeartbeatCheckLoopInterval: time.Second * 1,
}
