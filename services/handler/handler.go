package handler

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/telegram"
	"github.com/aibotsoft/micro/util"
	"github.com/aibotsoft/middle-service/pkg/clients"
	"github.com/aibotsoft/middle-service/pkg/store"
	"go.uber.org/zap"
	"time"
)

const (
	repeatMinTimeout    = 3 * time.Second
	secondBetMaxTry     = 57
	surebetLoopMaxCount = 5
)

type Handler struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	store   *store.Store
	clients clients.Clients
	tel     *telegram.Telegram
	Conf    *config_client.ConfClient
}

func (h *Handler) ProcessSurebet(sb *pb.Surebet) (err *SurebetError) {
	ctx := context.Background()
	sb.SurebetId = util.UnixUsNow()
	if err := h.AllServicesActive(sb); err != nil {
		return err
	}
	if err := h.store.LoadConfig(ctx, sb); err != nil {
		return &SurebetError{Err: err, Msg: "store.LoadConfig error", Permanent: true}
	}
	return
}
func (h *Handler) SurebetLoop(sb *pb.Surebet) {
	h.log.Info("handler_got_sb: ", sb.FortedSurebetId)
	_, ok := h.store.Cache.Get(sb.FortedSurebetId)
	if ok {
		h.log.Infow("loop_already_exists", "id", sb.FortedSurebetId)
		return
	}
	h.store.Cache.Set(sb.FortedSurebetId, true, 1)
	defer h.store.Cache.Del(sb.FortedSurebetId)
	for i := 0; i < surebetLoopMaxCount; i++ {
		ClearSurebet(sb)
		err := h.ProcessSurebet(sb)
		//h.SaveSurebet(sb)
		if err != nil {
			var otherName string
			if sb.Members[0].ServiceName == err.ServiceName {
				otherName = sb.Members[1].ServiceName
			} else {
				otherName = sb.Members[0].ServiceName
			}
			h.log.Infow("result", "err", err, "name", err.ServiceName, "time", ElapsedFromSurebetId(sb.SurebetId), "loop", i, "fid", sb.FortedSurebetId, "other", otherName)
			if err.Permanent {
				//h.log.Info("error permanent, so returning...")
				return
			}
		} else {
			go h.tel.Sendf("middle v=%v, l=%v, p=%v, f=%v, s=%v, t=%v", betAmount(sb), i, sb.Calc.Profit, sb.Members[0].ServiceName, sb.Members[1].ServiceName, ElapsedFromSurebetId(sb.SurebetId))
			h.log.Infow("placed_middle", "profit", sb.GetCalc().GetProfit(), "time", ElapsedFromSurebetId(sb.SurebetId), "loop", i, "fid", sb.FortedSurebetId)
			i = 0
		}
		//if sb.GetCalc().GetProfit() < -9 {
		//	h.log.Infow("profit_too_low, so returning...", "profit", sb.GetCalc().GetProfit(), "fid", sb.FortedSurebetId)
		//}
		time.Sleep(repeatMinTimeout + time.Millisecond*100*time.Duration(i))
	}
}

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, clients clients.Clients, conf *config_client.ConfClient) *Handler {
	return &Handler{cfg: cfg, log: log, store: store, clients: clients, tel: telegram.New(cfg, log), Conf: conf}
}
func (h *Handler) Close() {
	h.store.Close()
	h.Conf.Close()
}
