package main

import (
	"fmt"

	"github.com/7StaSH7/gometrics/internal/config"
	metricshandler "github.com/7StaSH7/gometrics/internal/handler/metrics"
	"github.com/7StaSH7/gometrics/internal/repository"
	metricsservice "github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/7StaSH7/gometrics/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error starting the server:", err)
	}
}

func run() error {
	sCfg := config.NewServerConfig()

	server := gin.Default()
	server.LoadHTMLGlob("templates/*")

	stor := storage.NewStorage()

	storRep := repository.NewMemStorageRepository(stor)

	mSer := metricsservice.New(storRep)

	mHan := metricshandler.NewHandler(mSer)

	mHan.Register(server)

	return server.Run(sCfg.Address)
}
