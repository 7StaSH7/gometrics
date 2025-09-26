package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/7StaSH7/gometrics/internal/config"
	dbconfig "github.com/7StaSH7/gometrics/internal/config/db"
	databaserepository "github.com/7StaSH7/gometrics/internal/repository/db"

	healthhandler "github.com/7StaSH7/gometrics/internal/handler/health"
	metricshandler "github.com/7StaSH7/gometrics/internal/handler/metrics"
	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/middleware"
	storagerepositsory "github.com/7StaSH7/gometrics/internal/repository/storage"
	metricsservice "github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/7StaSH7/gometrics/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error starting the server:", err)
	}
}

func initDeps(ctx context.Context) (*config.ServerConfig, *gin.Engine, metricsservice.MetricsService) {
	cfg, psqlCfg := config.NewServerConfig()

	router := gin.New()

	router.LoadHTMLGlob("templates/*")

	logger.Initialize(cfg.LogLevel)
	router.Use(middleware.RequestLogger)

	router.Use(middleware.GzipMiddleware)
	router.Use(gin.Recovery())

	stor := storage.NewStorage(cfg)

	psqlPool, err := dbconfig.NewPostgresDriver(ctx, psqlCfg)
	if err != nil {
		logger.Log.Error("psql connection error", zap.Error(err))
	}

	storRep := storagerepositsory.NewMemStorageRepository(stor)
	dbRep := databaserepository.NewDatabaseRepository(psqlPool)

	mSer := metricsservice.New(storRep, dbRep)

	mHan := metricshandler.New(mSer)
	hHan := healthhandler.New(psqlPool)

	mHan.Register(router)
	hHan.Register(router)

	return cfg, router, mSer
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	cfg, router, ser := initDeps(gCtx)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return gCtx
		},
	}

	if cfg.StoreInterval != 0 {
		g.Go(func() error {
			return ser.Store(gCtx, cfg.Restore, cfg.StoreInterval)
		})
	}

	g.Go(func() error {
		logger.Log.Info("server started", zap.String("address", cfg.Address))

		return srv.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()

		return srv.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err.Error())
	}

	return nil
}
