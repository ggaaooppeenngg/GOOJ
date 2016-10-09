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

var engine *xorm.Engine

func main() {
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

	// Add a new problem in problem set
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

		c.JSON(http.StatusOK, gin.H{"id": problem.Id})
	})

	// Get problem description
	r.GET("/problem/:id", func(c *gin.Context) {
		var problem model.Problem
		if _, err := engine.Id(c.Param("id")).Get(&problem); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"problem": problem})
		return

	})

	// Submit code to test
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
		if _, err := engine.InsertOne(&code); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := SaveFile(code.SourcePath(), code.Source); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": code.Id})
	})

	// Get code description
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

	r.Run() // listen and server on 0.0.0.0:8080
}
