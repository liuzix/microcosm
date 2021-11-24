package model

// OperatorType represents the type of the operator.
type OperatorType int32

// OperatorConfig is the configuration for an operator.
// It can be modified dynamically.
type OperatorConfig interface{}

// OperatorParams specifies the operator's behavior.
// It cannot be changed once the operator is scheduled.
type OperatorParams string

// OperatorHandle is the logical representation of the operator.
type OperatorHandle struct {
	Type   OperatorType   `json:"type"`
	Params OperatorParams `json:"params"`
	Config OperatorConfig `json:"config"`
}
