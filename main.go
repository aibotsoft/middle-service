package main

import (
	"fmt"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/mig"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/middle-service/pkg/clients"
	"github.com/aibotsoft/middle-service/pkg/store"
	"github.com/aibotsoft/middle-service/services/collector"
	"github.com/aibotsoft/middle-service/services/handler"
	"github.com/aibotsoft/middle-service/services/receiver"
	"github.com/aibotsoft/middle-service/services/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()
	log := logger.New()
	log.Infow("Begin service", "conf", cfg.Service)
	conf := config_client.New(cfg, log)
	db := sqlserver.MustConnectX(cfg)
	err := mig.MigrateUp(cfg, log, db)
	if err != nil {
		log.Fatal(err)
	}
	sto := store.New(cfg, log, db)
	cli := clients.NewClients(cfg, log, conf)
	h := handler.New(cfg, log, sto, cli, conf)
	s := server.New(cfg, log, h)
	// Инициализируем Close
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	c := collector.New(cfg, log, sto, cli)
	go c.CollectJob()
	r := receiver.New(cfg, log, h)
	r.Subscribe()

	go func() { errc <- s.Serve() }()
	defer func() { s.Close() }()
	log.Info("exit: ", <-errc)
}
