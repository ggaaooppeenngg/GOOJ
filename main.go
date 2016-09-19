package main

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"

	"github.com/ggaaooppeenngg/validator"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
)

var engine *xorm.Engine

func init() {
	var err error
	engine, err = xorm.NewEngine("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	if err := engine.Sync2(new(Problem)); err != nil {
		panic(err)
	}

}

func main() {
	r := gin.Default()
	r.POST("/problem", func(c *gin.Context) {
		var problem Problem
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
		c.JSON(http.StatusOK, gin.H{"id": problem.Id})
	})
	r.GET("/problem/:id", func(c *gin.Context) {
		var problem Problem
		_, err := engine.Id(c.Param("id")).Get(&problem)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"problem": problem})
		return

	})
	r.POST("/code/submit", func(c *gin.Context) {
		cmd := exec.Command("sandbox")
		out, err := cmd.CombinedOutput()
		if err != nil {
			c.JSON(400, gin.H{"error": err})
		}
		if string(out) == "hello world" {
			c.JSON(200, gin.H{"output": string(out)})
		} else {
			c.JSON(400, gin.H{"output": string(out)})
		}
	})
	r.Run() // listen and server on 0.0.0.0:8080
}
