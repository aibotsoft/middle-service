package main

import (
	"fmt"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/surebet-service/pkg/clients"
	"github.com/aibotsoft/surebet-service/pkg/store"
	"github.com/aibotsoft/surebet-service/services/collector"
	"github.com/aibotsoft/surebet-service/services/handler"
	"github.com/aibotsoft/surebet-service/services/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()
	log := logger.New()
	log.Infow("Begin service", "name", cfg.Service.Name, "config", cfg.Service.GrpcPort)
	conf := config_client.New(cfg, log)
	db := sqlserver.MustConnectX(cfg)
	sto := store.NewStore(cfg, log, db)
	cli := clients.NewClients(cfg, log, conf)
	h := handler.NewHandler(cfg, log, sto, cli, conf)
	s := server.NewServer(cfg, log, h)
	// Инициализируем Close
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	c := collector.New(cfg, log, sto, cli)
	go c.CollectJob()

	go func() { errc <- s.Serve() }()
	defer func() { s.Close() }()
	log.Info("exit: ", <-errc)
}
