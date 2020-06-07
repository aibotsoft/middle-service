package receiver

import (
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/middle-service/services/handler"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Receiver struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	handler *handler.Handler
	nats    *nats.EncodedConn
	sub     *nats.Subscription
}

func New(cfg *config.Config, log *zap.SugaredLogger, handler *handler.Handler) *Receiver {
	nc, err := nats.Connect("nats://192.168.1.10:30873")
	if err != nil {
		log.Panic(err)
	}
	c, err := nats.NewEncodedConn(nc, nats.GOB_ENCODER)
	if err != nil {
		log.Panic(err)
	}
	return &Receiver{cfg: cfg, log: log, handler: handler, nats: c}
}
func (r *Receiver) Close() {
	r.handler.Close()
	err := r.sub.Unsubscribe()
	if err != nil {
		r.log.Error(err)
	}
}

func (r *Receiver) Job(sb *pb.Surebet) {
	go r.handler.SurebetLoop(sb)
}

func (r *Receiver) Subscribe() {
	sub, err := r.nats.Subscribe("middle", r.Job)
	if err != nil {
		r.log.Error(err)
	}
	r.sub = sub
}
