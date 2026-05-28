package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"skybase/internal/config"
	httpHandler "skybase/internal/handler/http"
	"skybase/internal/service"
)

type App struct {
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	metaService := service.NewMetaService(cfg)
	db, err := service.OpenMySQL(cfg)
	if err != nil {
		return nil, err
	}
	agentService := service.NewAgentService(cfg, db)
	opsService := service.NewOpsService(db)
	loginLogService := service.NewLoginLogService(db)
	departmentService := service.NewDepartmentService(db)
	userService := service.NewUserService(db, departmentService)
	roleService := service.NewRoleService(db)
	authService := service.NewAuthService(cfg, userService, departmentService, roleService)
	log.Printf("mysql: pinging host=%s port=%d db=%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	if err := withStartupTimeout(func(ctx context.Context) error {
		return db.PingContext(ctx)
	}); err != nil {
		log.Printf("mysql: ping failed host=%s port=%d db=%s err=%v", cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, err)
		return nil, err
	}
	log.Printf("mysql: ping ok host=%s port=%d db=%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	log.Printf("redis: pinging addr=%s db=%d", cfg.Redis.Addr, cfg.Redis.DB)
	if err := withStartupTimeout(func(ctx context.Context) error {
		return authService.Ping(ctx)
	}); err != nil {
		log.Printf("redis: ping failed addr=%s db=%d err=%v", cfg.Redis.Addr, cfg.Redis.DB, err)
		return nil, err
	}
	log.Printf("redis: ping ok addr=%s db=%d", cfg.Redis.Addr, cfg.Redis.DB)
	log.Printf("bootstrap: ensuring auth bootstrap")
	if err := withStartupTimeout(func(ctx context.Context) error {
		return authService.EnsureBootstrap(ctx)
	}); err != nil {
		log.Printf("bootstrap: auth bootstrap failed err=%v", err)
		return nil, err
	}
	log.Printf("bootstrap: auth bootstrap ok")
	log.Printf("bootstrap: ensuring role bootstrap")
	if err := withStartupTimeout(func(ctx context.Context) error {
		return roleService.EnsureBootstrap(ctx, cfg.Auth.AdminUsername)
	}); err != nil {
		log.Printf("bootstrap: role bootstrap failed err=%v", err)
		return nil, err
	}
	log.Printf("bootstrap: role bootstrap ok")
	log.Printf("bootstrap: ensuring agent schema")
	if err := withStartupTimeout(func(ctx context.Context) error {
		return agentService.EnsureSchema(ctx)
	}); err != nil {
		log.Printf("bootstrap: agent schema ensure failed err=%v", err)
		return nil, err
	}
	log.Printf("bootstrap: agent schema ensure ok")
	log.Printf("bootstrap: syncing storage targets")
	if err := withStartupTimeout(func(ctx context.Context) error {
		return agentService.SyncStorageTargetsFromDB(ctx)
	}); err != nil {
		log.Printf("bootstrap: storage sync failed err=%v", err)
		return nil, err
	}
	log.Printf("bootstrap: storage sync ok")
	log.Printf("bootstrap: syncing groups")
	if err := withStartupTimeout(func(ctx context.Context) error {
		return agentService.SyncGroupsFromDB(ctx)
	}); err != nil {
		log.Printf("bootstrap: group sync failed err=%v", err)
		return nil, err
	}
	log.Printf("bootstrap: group sync ok")
	log.Printf("bootstrap: syncing agents")
	if err := withStartupTimeout(func(ctx context.Context) error {
		return agentService.SyncAgentsFromDB(ctx)
	}); err != nil {
		log.Printf("bootstrap: agent sync failed err=%v", err)
		return nil, err
	}
	log.Printf("bootstrap: agent sync ok")
	log.Printf("bootstrap: syncing versions")
	if err := withStartupTimeout(func(ctx context.Context) error {
		return agentService.SyncVersionsFromDB(ctx)
	}); err != nil {
		log.Printf("bootstrap: version sync failed err=%v", err)
		return nil, err
	}
	log.Printf("bootstrap: version sync ok")
	handler := httpHandler.NewServer(metaService, agentService, opsService, authService, roleService, loginLogService, departmentService, userService)

	return &App{
		server: &http.Server{
			Addr:              cfg.HTTP.Addr,
			Handler:           handler.Routes(),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func withStartupTimeout(fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return fn(ctx)
}
