package benchmark

import (
	"context"
	"sync"
	"time"

	"github.com/hanfei1991/microcosm/master/cluster"
	"github.com/hanfei1991/microcosm/model"
	"github.com/hanfei1991/microcosm/pb"
	"github.com/hanfei1991/microcosm/pkg/errors"
	"github.com/pingcap/ticdc/dm/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Master implements the master of benchmark workload.
type Master struct {
	*Config
	job *model.Job

	ctx    context.Context
	cancel func()

	resourceManager cluster.ResourceMgr
	client          cluster.ExecutorClient

	offExecutors chan model.ExecutorID

	mu           sync.Mutex
	execTasks    map[model.ExecutorID][]*model.Task
	runningTasks map[model.TaskID]*Task

	scheduleWaitingTasks chan scheduleGroup
	// rate limit for rescheduling when error happens
	scheduleRateLimit *rate.Limiter
}

// New creates a master instance
func New(
	parentCtx context.Context,
	config *Config,
	job *model.Job,
	resourceMgr cluster.ResourceMgr,
	client cluster.ExecutorClient,
) *Master {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Master{
		ctx:             ctx,
		cancel:          cancel,
		Config:          config,
		job:             job,
		resourceManager: resourceMgr,
		client:          client,

		offExecutors:         make(chan model.ExecutorID, 100),
		scheduleWaitingTasks: make(chan scheduleGroup, 1024),
		scheduleRateLimit:    rate.NewLimiter(rate.Every(time.Second), 1),

		execTasks:    make(map[model.ExecutorID][]*model.Task),
		runningTasks: make(map[model.TaskID]*Task),
	}
}

func (m *Master) Cancel() {
	m.cancel()
}

// scheduleGroup is the min unit of scheduler, and the tasks in the same group have to be scheduled in the same node.
type scheduleGroup []*Task

// TaskStatus represents the current status of the task.
type TaskStatus int32

const (
	// Running means the task is running.
	Running TaskStatus = iota
	// Stopped means the task has been stopped by any means.
	Stopped
	// Finished means the task has finished its job.
	Finished
)

// Task is the container of a dispatched tasks,  and records its status.
type Task struct {
	*model.Task

	exec   model.ExecutorID
	status TaskStatus
}

// ID implements JobMaster interface.
func (m *Master) ID() model.JobID {
	return m.job.ID
}

// master dispatches a set of task.
func (m *Master) dispatch(ctx context.Context, tasks []*Task) error {
	arrangement := make(map[model.ExecutorID][]*model.Task)
	for _, task := range tasks {
		subjob, ok := arrangement[task.exec]
		if !ok {
			arrangement[task.exec] = []*model.Task{task.Task}
		} else {
			subjob = append(subjob, task.Task)
			arrangement[task.exec] = subjob
		}
	}

	for execID, taskList := range arrangement {
		// construct sub job
		job := &model.Job{
			ID:    m.job.ID,
			Tasks: taskList,
		}
		reqPb := job.ToPB()
		log.L().Logger.Info("submit sub job", zap.Int32("exec id", int32(execID)), zap.String("req pb", reqPb.String()))
		request := &cluster.ExecutorRequest{
			Cmd: cluster.CmdSubmitBatchTasks,
			Req: reqPb,
		}
		resp, err := m.client.Send(ctx, execID, request)
		if err != nil {
			log.L().Logger.Info("Send meet error", zap.Error(err))
			return err
		}
		respPb := resp.Resp.(*pb.SubmitBatchTasksResponse)
		if respPb.Err != nil {
			return errors.ErrSubJobFailed.GenWithStackByArgs(execID, m.ID())
		}
	}

	// apply the new arrangement.
	m.mu.Lock()
	for eid, taskList := range arrangement {
		originTasks, ok := m.execTasks[eid]
		if ok {
			originTasks = append(originTasks, taskList...)
			m.execTasks[eid] = originTasks
		} else {
			m.execTasks[eid] = taskList
		}
	}
	for _, t := range tasks {
		m.runningTasks[t.ID] = t
	}
	m.mu.Unlock()
	return nil
}

// TODO: Implement different allocate task logic.
func (m *Master) allocateTasksWithNaiveStrategy(snapshot *cluster.ResourceSnapshot, taskInfos []*model.Task) (bool, []*Task) {
	var idx int = 0
	tasks := make([]*Task, 0, len(taskInfos))
	for _, task := range taskInfos {
		originalIdx := idx
		nTask := &Task{
			Task: task,
		}
		for {
			exec := snapshot.Executors[idx]
			used := exec.Used
			if exec.Reserved > used {
				used = exec.Reserved
			}
			rest := exec.Capacity - used
			if rest >= cluster.ResourceUsage(task.Cost) {
				nTask.exec = exec.ID
				exec.Reserved = exec.Reserved + cluster.ResourceUsage(task.Cost)
				break
			}
			idx = (idx + 1) % len(snapshot.Executors)
			if idx == originalIdx {
				return false, nil
			}
		}
		tasks = append(tasks, nTask)
	}
	return true, tasks
}

func (m *Master) reScheduleTask(group scheduleGroup) error {
	snapshot := m.resourceManager.GetResourceSnapshot()
	if len(snapshot.Executors) == 0 {
		return errors.ErrClusterResourceNotEnough.GenWithStackByArgs()
	}
	taskInfos := make([]*model.Task, 0, len(group))
	for _, t := range group {
		taskInfos = append(taskInfos, t.Task)
	}
	success, tasks := m.allocateTasksWithNaiveStrategy(snapshot, taskInfos)
	if !success {
		return errors.ErrClusterResourceNotEnough.GenWithStackByArgs()
	}
	err := m.dispatch(m.ctx, tasks)
	return err
}

func (m *Master) scheduleJobImpl(ctx context.Context) error {
	snapshot := m.resourceManager.GetResourceSnapshot()
	if len(snapshot.Executors) == 0 {
		return errors.ErrClusterResourceNotEnough.GenWithStackByArgs()
	}
	success, tasks := m.allocateTasksWithNaiveStrategy(snapshot, m.job.Tasks)
	if !success {
		return errors.ErrClusterResourceNotEnough.GenWithStackByArgs()
	}

	m.start() // go
	err := m.dispatch(ctx, tasks)
	return err
}

// DispatchJob implements JobMaster interface.
func (m *Master) DispatchJob(ctx context.Context) error {
	retry := 1
	for i := 1; i <= retry; i++ {
		if err := m.scheduleJobImpl(ctx); err == nil {
			return nil
		} else if i == retry {
			return err
		}
		// sleep for a while to backoff
	}
	return nil
}

// Listen the events from every tasks
func (m *Master) start() {
	// Register Listen Handler to Msg Servers

	// Run watch goroutines
	// TODO: keep the goroutines alive.
	go m.monitorExecutorOffline()
	go m.monitorSchedulingTasks()
}

func (m *Master) monitorSchedulingTasks() {
	for {
		select {
		case group := <-m.scheduleWaitingTasks:
			//for _, t := range group {
			//	curT := m.runningTasks[t.ID]
			//	if curT.exec != t.exec {
			//		// this task has been scheduled away.
			//		log.L().Logger.Info("cur task exec id is not same as reschedule one", zap.Int32("cur id", int32(curT.exec)), zap.Int32("id", int32(t.exec)))
			//		continue
			//	}
			//}

			//if t.status == Running {
			// cancel it
			//}

			log.L().Logger.Info("begin to reschedule task group", zap.Any("group", group))
			if err := m.reScheduleTask(group); err != nil {
				log.L().Logger.Error("cant reschedule task", zap.Error(err))

				// Use a global rate limit for task rescheduling
				delay := m.scheduleRateLimit.Reserve().Delay()
				if delay != 0 {
					log.L().Logger.Warn("reschedule task rate limit", zap.Duration("delay", delay))
					timer := time.NewTimer(delay)
					select {
					case <-m.ctx.Done():
						timer.Stop()
						return
					case <-timer.C:
						timer.Stop()
					}
				}
				// FIXME: this could cause deadlock problem if scheduleWaitingTasks channel is full
				m.scheduleWaitingTasks <- group
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// OfflineExecutor implements JobMaster interface.
func (m *Master) OfflineExecutor(id model.ExecutorID) {
	m.offExecutors <- id
	log.L().Logger.Info("executor is offlined", zap.Int32("eid", int32(id)))
}

func (m *Master) monitorExecutorOffline() {
	for {
		select {
		case execID := <-m.offExecutors:
			log.L().Logger.Info("executor is offlined", zap.Int32("eid", int32(execID)))
			m.mu.Lock()
			taskList, ok := m.execTasks[execID]
			if !ok {
				m.mu.Unlock()
				log.L().Logger.Info("executor has been removed, nothing todo", zap.Int32("id", int32(execID)))
				continue
			}
			delete(m.execTasks, execID)
			m.mu.Unlock()

			var group scheduleGroup
			for _, task := range taskList {
				t, ok := m.runningTasks[task.ID]
				if !ok || t.exec != execID {
					log.L().Logger.Error("running task is not consistant with executor-task map")
					continue
				}
				t.status = Finished
				group = append(group, t)
			}
			m.scheduleWaitingTasks <- group
		case <-m.ctx.Done():
			return
		}
	}
}
