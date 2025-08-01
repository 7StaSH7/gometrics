package main

import (
	"fmt"

	metricshandler "github.com/7StaSH7/gometrics/internal/handler/metrics"
	"github.com/7StaSH7/gometrics/internal/model"
	metricsservice "github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error starting the server:", err)
	}
}

func run() error {
	server := gin.Default()

	model.NewStorage()

	mSer := metricsservice.New()

	mHan := metricshandler.NewHandler(mSer)

	mHan.Register(server)

	return server.Run(":8080")
}
