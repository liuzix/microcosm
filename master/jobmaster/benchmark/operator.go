package benchmark

import "github.com/hanfei1991/microcosm/model"

const (
	TableReaderType model.OperatorType = iota
	HashType
	TableSinkType
)

// benchmark operators
type TableReaderOp struct {
	Addr     string `json:"address"`
	TableNum int32  `json:"table-num"`
}

type HashOp struct {
	TableID int32 `json:"id"`
}

type TableSinkOp struct {
	TableID int32  `json:"id"`
	File    string `json:"file"`
}
