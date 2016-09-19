package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-xorm/xorm"

	"github.com/ggaaooppeenngg/OJ/model"
)

var (
	engine *xorm.Engine
)

func init() {
	var err error
	engine, err = xorm.NewEngine("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
}

type result struct {
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
func convResult(out []byte) *result {
	results := strings.Split(fmt.Sprintf("%s", out), ":")
	statuss := results[0]
	var (
		status              model.JudgeResult
		wrongAnswer         string
		memory, time, nth64 int64
	)
	if len(results) > 1 {
		memory, _ = strconv.ParseInt(results[1], 0, 64)
	}
	if len(results) > 2 {
		time, _ = strconv.ParseInt(results[2], 0, 64)
	}
	if len(results) > 3 {
		nth64, _ = strconv.ParseInt(results[3], 0, 64)
	}
	switch statuss {
	case "AC":
		status = model.Accept
	case "CE":
		status = model.CompileError
	case "TL":
		status = model.TimeLimitExceeded
	case "ML":
		status = model.MemoryLimitExceeded
	case "RE":
		status = model.RuntimeError
	case "FE":
		status = model.PresentationError
	case "WA":
		status = model.WrongAnswer
		if len(results) > 4 {
			wrongAnswer = results[4]
		}
		nth64 += 1
	}
	return &result{Status: status, Time: time, Memory: memory, Nth: int(nth64), WrongAnswer: wrongAnswer}
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
				fmt.Sprintf("--lang=%s", code.Lang),
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
				log.Error(err)
				_, err = engine.Id(code.Id).Cols("status").Update(&model.Code{Status: model.Unhandled})
				if err != nil {
					log.Error(err)
				}
				return
			}
			var rslt *result
			// panic output
			if !regexp.MustCompile(`\w\w:\d+:\d+:([\s\S]*)`).Match(out) {
				rslt = &result{Status: model.PanicError, PanicOutput: fmt.Sprintf("%s", out)}
				log.WithField("code", code).
					Errorf("exception out %s", out)
				_, err = engine.Id(code.Id).Cols("status").Update(&model.Code{Status: model.Unhandled})
				if err != nil {
					log.Error(err)
				}
				return
			}
			rslt = convResult(out)
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
