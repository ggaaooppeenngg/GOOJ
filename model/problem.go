package model

import (
	"fmt"
)

// Problem is a model of problem.
type Problem struct {
	Id           int64  `json:"id"`                                                        // primary key
	Title        string `validate:"nonzero"       json:"title"`                            // problem title
	Solved       int64  ``                                                                 // times of accepted submit
	TimeLimit    int64  `validate:"nonzero,min=1" json:"timeLimit"`                        // time limit in ms
	MemoryLimit  int64  `validate:"nonzero,min=1" json:"memoryLimit"`                      // memory limit in byte
	Description  string `validate:"nonzero"       json:"description"  xorm:"TEXT"`         // problem description
	InputSample  string `                         json:"outputSample" xorm:"varchar(512)"` // input sample
	OutputSample string `                         json:"intputSample" xorm:"varchar(512)"` // output sample
	Input        string `validate:"nonzero"       json:"input"        xorm:"-"`            // input test
	Output       string `validate:"nonzero"	      json:"output"       xorm:"-"`            // output test
	PosterId     int64  ``                                                                 // Post id TODO
}

func (p Problem) InputTestPath() string {
	return fmt.Sprintf("problems/%d-input.txt", p.Id)
}

func (p Problem) OutputTestPath() string {
	return fmt.Sprintf("problems/%d-output.txt", p.Id)
}
