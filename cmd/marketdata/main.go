package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mantis-exchange/mantis-market-data/internal/config"
	"github.com/mantis-exchange/mantis-market-data/internal/consumer"
	"github.com/mantis-exchange/mantis-market-data/internal/handler"
	"github.com/mantis-exchange/mantis-market-data/internal/model"
	"github.com/mantis-exchange/mantis-market-data/internal/service"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	klineRepo := model.NewKlineRepo(pool)
	aggregator := service.NewKlineAggregator(klineRepo)
	depthService := service.NewDepthService()

	// Start trade consumer
	tc := consumer.New(klineRepo, aggregator, depthService, cfg.KafkaBrokers)
	go tc.Start()

	// Periodically flush candles
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			aggregator.FlushAll(context.Background())
		}
	}()

	h := handler.New(klineRepo, depthService)

	r := gin.Default()
	api := r.Group("/api/v1")
	{
		api.GET("/klines", h.GetKlines)
		api.GET("/trades", h.GetTrades)
		api.GET("/depth", h.GetDepth)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("mantis-market-data starting on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
