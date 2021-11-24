package planner

import "github.com/hanfei1991/microcosm/model"

type TrivialPlanner struct{}

func (p *TrivialPlanner) Plan(dag *model.SubJob) (*PhysicalPlan, error) {
	return nil, nil
}
