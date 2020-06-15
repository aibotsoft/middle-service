package handler

import (
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/middle-service/pkg/clients"
	"github.com/aibotsoft/middle-service/pkg/store"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var h *Handler

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.New(cfg, log, db)
	conf := config_client.New(cfg, log)
	cli := clients.NewClients(cfg, log, conf)
	h = New(cfg, log, sto, cli, conf)
	m.Run()
	h.Close()
}
func surebetHelper(t *testing.T) *pb.Surebet {
	t.Helper()
	return &pb.Surebet{
		CreatedAt: time.Time{}.Format(time.RFC3339Nano),
		//Starts:       time.Time{}.Format(time.RFC3339Nano),
		Starts:       "2020-05-02T10:28:01.568",
		FortedHome:   "FortedHome",
		FortedAway:   "FortedAway",
		FortedProfit: 6.66,
		Members: []*pb.SurebetSide{{
			Num:         1,
			ServiceName: "Sbobet",
			SportName:   "E Sports",
			LeagueName:  "FIFA 20 - Russia Liga Pro (12 Mins)",
			Home:        "EZ1D 11 (EZ1)",
			Away:        "Gambit Esports (GMB)",
			MarketName:  "ТМ(3,25)",
			Price:       2.26,
			Url:         "https://www.sbobet.com/ru-ru/euro/e-sports/fifa-20---russia-liga-pro-(12-mins)/2985196/ez1d-11-(ez1)-vs-gambit-esports-(gmb)",
			Initiator:   true,
		}, {Num: 2,
			ServiceName: "Pinnacle",
			SportName:   "Soccer",
			LeagueName:  "eSoccer - Liga Pro (12 mins)",
			Home:        "Ez1d (EZ1) Esports",
			Away:        "Gambit (GMB) Esports",
			MarketName:  "ТБ(3,25)",
			Price:       1.862,
			Url:         "https://members.pinnacle.com/Sportsbook/Mobile/ru-RU/Enhanced/Regular/SportsBookAll/35/Curacao/Odds/Soccer-29/Market/2/208851/1122224338",
			Initiator:   false,
		}}}
}
func TestHandler_ProcessSurebet(t *testing.T) {
	sb := surebetHelper(t)
	err := h.ProcessSurebet(sb)
	assert.NoError(t, err)
	//t.Log(sb.Members)
}
