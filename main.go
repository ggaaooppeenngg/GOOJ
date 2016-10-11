package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"

	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"

	"github.com/ggaaooppeenngg/OJ/model"
	"github.com/ggaaooppeenngg/validator"
)

var (
	engine *xorm.Engine
	inTest bool
)

func NewEngine() *gin.Engine {
	var err error
	engine, err = xorm.NewEngine("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	if err := engine.Sync2(new(model.Problem), new(model.Code)); err != nil {
		panic(err)
	}
	r := gin.New()
	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type, X-Requested-With",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	// POST /problem adds a new problem in problem set
	r.POST("/problem", func(c *gin.Context) {
		var problem model.Problem
		if err := c.BindJSON(&problem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errs := validator.Validate(problem); errs != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errs})
			return
		}

		transaction := engine.NewSession()
		defer transaction.Close()
		if err := transaction.Begin(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if _, err := engine.InsertOne(&problem); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := SaveFile(problem.InputTestPath(), problem.Input); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			engine.Delete(problem)
			return
		}
		if err := SaveFile(problem.OutputTestPath(), problem.Output); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			engine.Delete(problem)
			return
		}
		if err := transaction.Commit(); err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": problem.Id})
	})

	// POST /problems gets problems by limit and start
	r.POST("/problems", func(c *gin.Context) {
		var problems []model.Problem
		var req struct {
			Limit int `json:"limit"`
			Start int `json:"start"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := engine.Limit(req.Limit, req.Start).Find(&problems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"problems": problems})
	})

	// GET /problem/:id gets problem description
	r.GET("/problem/:id", func(c *gin.Context) {
		var problem model.Problem
		if _, err := engine.Id(c.Param("id")).Get(&problem); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"problem": problem})
		return

	})

	// POST /code submits code to test
	r.POST("/code", func(c *gin.Context) {
		var code model.Code
		if err := c.BindJSON(&code); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := code.Init(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if errs := validator.Validate(code); errs != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errs})
			return
		}
		transaction := engine.NewSession()
		defer transaction.Close()
		if err := transaction.Begin(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if _, err := engine.InsertOne(&code); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := SaveFile(code.SourcePath(), code.Source); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := transaction.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": code.Id})
	})

	// POST /codes gets codes by limit and start
	r.POST("/codes", func(c *gin.Context) {
		var codes []model.Code
		var req struct {
			Limit int `json:"limit"`
			Start int `json:"start"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := engine.Limit(req.Limit, req.Start).Find(&codes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"codes": codes})
	})

	// GET /code/:id gets code description
	r.GET("/code/:id", func(c *gin.Context) {
		var (
			code    model.Code
			problem model.Problem
		)
		if _, err := engine.Id(c.Param("id")).Get(&code); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if _, err := engine.Id(code.ProblemId).Get(&problem); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "problem": problem})
		return

	})

	return r
}

func main() {
	r := NewEngine()
	r.Run() // listen and server on 0.0.0.0:8080
}
