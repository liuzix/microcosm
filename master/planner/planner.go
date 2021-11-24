package planner

import "github.com/hanfei1991/microcosm/model"

// PhysicalPlan represents an executor-runnable plan.
type PhysicalPlan = model.Job

// Planner converts a logical plan (subjob) to a physical plan.
type Planner interface {
	Plan(dag *model.SubJob) (*PhysicalPlan, error)
}
