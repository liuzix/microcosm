package model

type (
	NodeID   int32
	SubJobID int32
)

// DAG is the directed acyclic graph for SubJobs.
type DAG struct {
	// Root is the root node of the DAG.
	// We represent DAG in a recursive data structure for easier manipulation
	// when building it.
	Root *Node `json:"root"`
}

// Node is a node in the DAG.
type Node struct {
	ID       NodeID          `json:"id"`
	Operator *OperatorHandle `json:"operator"`
	Outputs  []*Node         `json:"outputs"`
}

// SubJob is a logical subjob that can be submitted to the Planner.
type SubJob struct {
	Job JobID `json:"job"`
	// SubJobName provides the JobMaster an identifier to manage the subjobs.
	SubJobName string `json:"subjob_name"`
	DAG        *DAG   `json:"dag"`
}
