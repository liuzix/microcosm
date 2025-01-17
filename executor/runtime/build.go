package runtime

import (
	"encoding/json"

	"github.com/hanfei1991/microcosm/master/jobmaster/benchmark"

	"github.com/hanfei1991/microcosm/model"
	"github.com/hanfei1991/microcosm/pkg/errors"
	"github.com/pingcap/ticdc/dm/pkg/log"
	"go.uber.org/zap"
)

func newTaskContainer(task *model.Task, ctx *taskContext) *taskContainer {
	t := &taskContainer{
		cfg: task,
		id:  task.ID,
		ctx: ctx,
	}
	return t
}

func newHashOp(cfg *benchmark.HashOp) operator {
	return &opHash{}
}

func newReadTableOp(cfg *benchmark.TableReaderOp) operator {
	return &opReceive{
		addr:     cfg.Addr,
		data:     make(chan *Record, 1024),
		tableCnt: cfg.TableNum,
	}
}

func newSinkOp(cfg *benchmark.TableSinkOp) operator {
	return &opSink{
		writer: fileWriter{
			filePath: cfg.File,
			tid:      cfg.TableID,
		},
	}
}

func (s *Runtime) connectTasks(sender, receiver *taskContainer) {
	ch := &Channel{
		innerChan: make(chan *Record, 1024),
		sendWaker: s.getWaker(sender),
		recvWaker: s.getWaker(receiver),
	}
	sender.output = append(sender.output, ch)
	receiver.inputs = append(receiver.inputs, ch)
}

func (s *Runtime) SubmitTasks(tasks []*model.Task) error {
	taskSet := make(map[model.TaskID]*taskContainer)
	for _, t := range tasks {
		task := newTaskContainer(t, s.ctx)
		log.L().Logger.Info("config", zap.ByteString("op", t.Op))
		switch t.OpTp {
		case benchmark.TableReaderType:
			op := &benchmark.TableReaderOp{}
			err := json.Unmarshal(t.Op, op)
			if err != nil {
				return err
			}
			task.op = newReadTableOp(op)
			task.setRunnable()
		case benchmark.HashType:
			op := &benchmark.HashOp{}
			err := json.Unmarshal(t.Op, op)
			if err != nil {
				return err
			}
			task.op = newHashOp(op)
			task.tryBlock()
		case benchmark.TableSinkType:
			op := &benchmark.TableSinkOp{}
			err := json.Unmarshal(t.Op, op)
			if err != nil {
				return err
			}
			task.op = newSinkOp(op)
			task.tryBlock()
		}
		taskSet[task.id] = task
	}

	log.L().Logger.Info("begin to connect tasks")

	for _, t := range taskSet {
		for _, tid := range t.cfg.Outputs {
			dst, ok := taskSet[tid]
			if ok {
				s.connectTasks(t, dst)
			} else {
				return errors.ErrTaskNotFound.GenWithStackByArgs(tid)
			}
		}
		err := t.prepare()
		if err != nil {
			return err
		}
	}

	log.L().Logger.Info("begin to push")
	// add to queue, begin to run.
	for _, t := range taskSet {
		if t.status == int32(Runnable) {
			s.q.push(t)
		}
	}
	return nil
}
