package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"

	"github.com/ggaaooppeenngg/OJ/loghook"
	"github.com/ggaaooppeenngg/OJ/model"
)

var (
	engine *xorm.Engine
)

type M log.Fields

func init() {
	var err error
	engine, err = xorm.NewEngine("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	log.AddHook(loghook.NewCallerHook())
}

type Result struct {
	StatusLit   string `json:"status"`
	Status      model.JudgeResult
	Time        int64
	Memory      int64
	Nth         int
	WrongAnswer string
	PanicOutput string // exception output
}

func getUnhandledCode() <-chan model.Code {
	unHandledCodeChan := make(chan model.Code)
	go func() {
		defer close(unHandledCodeChan)
		for {
			var codes []model.Code
			err := engine.Where("status = ?", model.Unhandled).Find(&codes)

			if err != nil {
				log.Error(err)
			}
			for _, code := range codes {
				_, err = engine.Id(code.Id).Cols("status").Update(&model.Code{Status: model.Handling})
				if err != nil {
					log.Error(err)
					continue
				}
				unHandledCodeChan <- code
			}
			time.Sleep(time.Second)
		}
	}()
	return unHandledCodeChan
}

func (r *Result) Init() {
	switch r.StatusLit {
	case "AC":
		r.Status = model.Accept
	case "CE":
		r.Status = model.CompileError
	case "TL":
		r.Status = model.TimeLimitExceeded
	case "ML":
		r.Status = model.MemoryLimitExceeded
	case "RE":
		r.Status = model.RuntimeError
	case "FE":
		r.Status = model.PresentationError
	case "WA":
		r.Status = model.WrongAnswer
	}
}

func judgeCode(codeChan <-chan model.Code) {
	for codec := range codeChan {
		// TODO: taskpool
		code := codec
		go func() {
			var problem model.Problem
			if _, err := engine.Id(code.ProblemId).Get(&problem); err != nil {
				log.Error(err)
				_, err = engine.Id(code.Id).Cols("status").Update(&model.Code{Status: model.Unhandled})
				if err != nil {
					log.Error(err)
				}
				return
			}
			cmd := exec.Command("sandbox",
				fmt.Sprintf("--lang=%s", strings.ToLower(code.Language.String())),
				fmt.Sprintf("--time=%d", problem.TimeLimit),
				fmt.Sprintf("--memory=%d", problem.MemoryLimit),
				"--compile",
				"--source", code.SourcePath(),
				"--binary", code.BinaryPath(),
				"--input", problem.InputTestPath(),
				"--output", problem.OutputTestPath(),
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.WithFields(log.Fields{
					"command": strings.Join(cmd.Args, " "),
					"output":  string(out),
				}).Error(err)
				_, err = engine.Id(code.Id).Cols("status").Update(&model.Code{Status: model.Unhandled})
				if err != nil {
					log.Error(err)
				}
				return
			}
			var rslt Result
			// panic output
			if err := json.Unmarshal(out, &rslt); err != nil {
				rslt = Result{Status: model.PanicError, PanicOutput: fmt.Sprintf("%s", out)}
				log.WithFields(log.Fields{"code": code}).
					Errorf("exception out %s", out)
				_, err = engine.Id(code.Id).Cols("status").Update(&model.Code{Status: model.Unhandled})
				if err != nil {
					log.Error(err)
				}
				return
			}
			rslt.Init()
			transaction := engine.NewSession()
			defer transaction.Close()
			err = transaction.Begin()
			if err != nil {
				log.Error(err)
				return
			}
			if _, err := transaction.Id(code.Id).Cols("status", "time", "memory", "nth", "wrong_answer").Update(model.Code{
				Status:      rslt.Status,
				Time:        rslt.Time,
				Memory:      rslt.Memory,
				Nth:         rslt.Nth,
				WrongAnswer: rslt.WrongAnswer,
			}); err != nil {
				log.Error(err)
				err = transaction.Rollback()
				if err != nil {
					log.Error(err)
				}
				return
			}
			if _, err := engine.Id(code.ProblemId).Incr("solved", 1).Update(model.Problem{}); err != nil {
				log.Error(err)
				err = transaction.Rollback()
				if err != nil {
					log.Error(err)
				}
				return
			}
			err = transaction.Commit()
			if err != nil {
				log.Error(err)
				return
			}

		}()
	}
}

func main() {
	codeChan := getUnhandledCode()
	judgeCode(codeChan)
}
