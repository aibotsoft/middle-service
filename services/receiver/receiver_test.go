package receiver

import (
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/surebet-service/pkg/clients"
	"github.com/aibotsoft/surebet-service/pkg/store"
	"github.com/aibotsoft/surebet-service/services/handler"
	"testing"
	"time"
)

var r *Receiver

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.NewStore(cfg, log, db)
	conf := config_client.New(cfg, log)
	cli := clients.NewClients(cfg, log, conf)
	h := handler.NewHandler(cfg, log, sto, cli, conf)
	r = New(cfg, log, h)
	m.Run()
	r.Close()
}

func TestNew(t *testing.T) {
	sb := pb.Surebet{
		CreatedAt:       "1",
		Starts:          "2",
		FortedHome:      "3",
		FortedAway:      "4",
		FortedProfit:    5,
		FortedSport:     "6",
		FortedLeague:    "7",
		FilterName:      "8",
		SkynetId:        9,
		FortedSurebetId: 10,
		SurebetId:       11,
		LogId:           12,
	}
	//var network bytes.Buffer
	//enc := gob.NewEncoder(&network) // Will write to network.
	//dec := gob.NewDecoder(&network) // Will read from network.
	//// Encode (send) the value.
	//err := enc.Encode(sb)
	//if err != nil {
	//	r.log.Fatal("encode error:", err)
	//}
	//r.log.Infow("", "", sb)
	//t.Log(network.Bytes())
	//var q pb.Surebet
	//err = dec.Decode(&q)
	//if err != nil {
	//	r.log.Fatal("decode error:", err)
	//}
	//r.log.Infow("", "", sb)

	r.Subscribe()
	r.Send(&sb)
	//r.Send()
	//r.Send()
	time.Sleep(time.Second)
}
