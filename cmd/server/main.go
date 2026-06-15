package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/config"
	"github.com/dosu-logi/logistics-erp/internal/db"
	"github.com/dosu-logi/logistics-erp/internal/integration/mailer"
	"github.com/dosu-logi/logistics-erp/internal/integration/tracking3p"
	"github.com/dosu-logi/logistics-erp/internal/module/marketing"
	"github.com/dosu-logi/logistics-erp/internal/module/tracking"
	"github.com/dosu-logi/logistics-erp/internal/router"
	"github.com/dosu-logi/logistics-erp/internal/util"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	if err := db.EnsureDatabase(ctx, cfg); err != nil {
		log.Printf("warn: could not ensure database (postgres may be offline): %v", err)
	}

	pool, err := db.NewPostgres(ctx, cfg)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	redisClient := db.NewRedis(cfg)
	if err := db.PingRedis(ctx, redisClient); err != nil {
		log.Printf("warn: redis unavailable: %v", err)
	} else {
		log.Println("redis connected")
	}
	defer redisClient.Close()

	_ = os.MkdirAll(cfg.UploadDir, 0755)
	_ = os.MkdirAll(cfg.UploadDir+"/contracts", 0755)
	_ = os.MkdirAll(cfg.UploadDir+"/invoices", 0755)

	jwtMgr := util.NewJWTManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	trackRepo := tracking.NewRepository(pool)
	trackClient := tracking3p.NewClient(cfg.TrackingAPIBaseURL, cfg.TrackingAPIKey)
	trackSvc := tracking.NewService(trackRepo, trackClient)
	poller := tracking.NewPoller(trackSvc, cfg.TrackingPollInterval)

	mktRepo := marketing.NewRepository(pool)
	mailer := mailer.New(cfg.FromEmail)
	mktSvc := marketing.NewService(mktRepo, mailer)
	scheduler := marketing.NewScheduler(mktSvc)

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	poller.Start(appCtx)
	scheduler.Start(appCtx)

	r := router.Setup(router.Deps{
		Config: cfg,
		DB:     pool,
		JWT:    jwtMgr,
		Poller: poller,
		Sched:  scheduler,
	})

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on :%s (env=%s, db=%s)", cfg.AppPort, cfg.AppEnv, cfg.DBName)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	poller.Stop()
	scheduler.Stop()
	_ = srv.Shutdown(shutdownCtx)
}
