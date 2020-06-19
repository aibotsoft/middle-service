package collector

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/middle-service/pkg/clients"
	"github.com/aibotsoft/middle-service/pkg/store"
	"go.uber.org/zap"
	"time"
)

type Collector struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	store   *store.Store
	clients clients.Clients
}

const CollectJobPeriod = 5 * time.Minute

func (c *Collector) CollectJob() {
	for {
		err := c.CollectResultsRound()
		if err != nil {
			c.log.Error(err)
		}
		time.Sleep(CollectJobPeriod)
	}
}
func (c *Collector) CollectResultsRound() error {
	var res []pb.BetResult
	for _, client := range c.clients {
		results, err := client.GetResults(context.Background(), &pb.GetResultsRequest{})
		if err != nil {
			c.log.Error(err)
			continue
		}
		res = append(res, results.GetResults()...)
	}
	err := c.store.SaveBetList(res)
	return err
}

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, clients clients.Clients) *Collector {
	return &Collector{cfg: cfg, log: log, store: store, clients: clients}
}
