package main

// Problem is a model of problem.
type Problem struct {
	Id             int64  ``                                                                 // primary key
	Title          string `validate:"nonzero"       json:"title"`                            // problem title
	Solved         int64  ``                                                                 // times of accepted submit
	TimeLimit      int64  `validate:"nonzero,min=1" json:"timeLimit"`                        // time limit in ms
	MemoryLimit    int64  `validate:"nonzero,min=1" json:"memoryLimit"`                      // memory limit in byte
	Description    string `validate:"nonzero"       json:"description"  xorm:"TEXT"`         // problem description
	InputSample    string `                         json:"outputSample" xorm:"varchar(512)"` // input sample
	OutputSample   string `                         json:"intputSample" xorm:"varchar(512)"` // output sample
	InputTestPath  string ``                                                                 // input test path
	OutputTestPath string ``                                                                 // output test path
	PosterId       int64  ``                                                                 // Post id TODO
}
