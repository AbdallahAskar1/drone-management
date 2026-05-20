package main

import (
	"log"
	"os"

	"drone-management/internal/config"
	"drone-management/internal/database"
	"drone-management/internal/handler"
	mw "drone-management/internal/middleware"
	"drone-management/internal/repo"
	"drone-management/internal/service"
	"drone-management/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database open: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	clock := utils.RealClock{}
	signer := utils.NewJWTSigner(cfg.JWTSecret, cfg.JWTTTL)

	principalRepo := repo.NewPrincipalRepo(db)
	droneRepo := repo.NewDroneRepo(db)
	orderRepo := repo.NewOrderRepo(db)
	jobRepo := repo.NewJobRepo(db)
	eventRepo := repo.NewEventRepo(db)

	authSvc := service.NewAuthService(principalRepo, droneRepo, signer, clock)
	orderSvc := service.NewOrderService(db, orderRepo, jobRepo, eventRepo, clock)
	droneSvc := service.NewDroneService(db, droneRepo, orderRepo, jobRepo, eventRepo, principalRepo, clock, cfg.AvgSpeedMS)
	adminSvc := service.NewAdminService(db, orderRepo, droneRepo, eventRepo, clock)

	handlers := handler.Handlers{
		Auth:    handler.NewAuthHandler(authSvc),
		Enduser: handler.NewEnduserHandler(orderSvc),
		Drone:   handler.NewDroneHandler(droneSvc),
		Admin:   handler.NewAdminHandler(adminSvc, droneSvc),
	}

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(mw.RequestID())
	e.Use(middleware.RequestLogger())

	handler.Register(e, signer, handlers)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := e.Start(addr); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
