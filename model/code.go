package model

import (
	"fmt"
	"strings"
	"time"
)

type JudgeResult int

//go:generate stringer -type JudgeResult

// test error types
const (
	Unhandled JudgeResult = iota
	Accept
	CompileError
	WrongAnswer
	TimeLimitExceeded
	MemoryLimitExceeded
	Handling
	RuntimeError
	PresentationError
	PanicError
)

type Language int

//go:generate stringer -type Language

// language
const (
	Go Language = iota
	C
	CPP
)

// delimiter
const (
	DELIM = "!-_-\n" //delimiter of tests
)

// code is a resolution to some problem
type Code struct {
	Id          int64       `json:"-"`
	ProblemId   int64       `json:"problemId" validate:"nonzero"`
	CreatedAt   time.Time   `json:"-"`
	Status      JudgeResult `json:"-"`
	Lang        string      `json:"language"  validate:"nonzero" xorm:"-"` // source code literal language
	Language    Language    `json:"-"`                                     // source code language
	Time        int64       `json:"-"`                                     // time used in ms
	Memory      int64       `json:"-"`                                     // memory used in KB
	Nth         int         `json:"-"`                                     // the number of the test not passed
	WrongAnswer string      `json:"-"`                                     // the last wrong answer
	PanicError  string      `json:"-"`                                     // panic ouput
	Version     int         `json:"-"         xorm:"version"`              // happy lock
	Source      string      `json:"source"    validate:"nonzero" xorm:"-"` // source code
}

func (c *Code) Init() error {
	switch c.Lang {
	case "c":
		c.Language = C
	case "cpp":
		c.Language = CPP
	case "go":
		c.Language = Go
	default:
		return fmt.Errorf("unknown or unspported language %s", c.Language)
	}
	c.CreatedAt = time.Now()
	return nil
}

func (c Code) SourcePath() string {
	return fmt.Sprintf("codes/%d.%s", c.Id, strings.ToLower(c.Language.String()))
}

func (c Code) BinaryPath() string {
	return fmt.Sprintf("codes/%d", c.Id)
}
