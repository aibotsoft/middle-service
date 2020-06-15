package handler

import (
	"context"
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/telegram"
	"github.com/aibotsoft/micro/util"
	"github.com/aibotsoft/middle-service/pkg/clients"
	"github.com/aibotsoft/middle-service/pkg/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
	"sync"
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

func (h *Handler) SendCheckLines(ctx context.Context, sb *pb.Surebet) {
	var wg sync.WaitGroup
	wg.Add(2)
	go h.CheckLine(ctx, sb, 0, &wg)
	go h.CheckLine(ctx, sb, 1, &wg)
	wg.Wait()
}

func (h *Handler) ReleaseChecks(sb *pb.Surebet) {
	var wg sync.WaitGroup
	for i := range sb.Members {
		wg.Add(1)
		go h.ReleaseCheck(sb.Members[i].ServiceName, sb.SurebetId, &wg)
	}
	wg.Wait()
}
func (h *Handler) ReleaseCheck(serviceName string, surebetId int64, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	_, err := h.clients[serviceName].ReleaseCheck(context.Background(), &pb.ReleaseCheckRequest{SurebetId: surebetId})
	if err != nil {
		h.log.Error(err)
	}
}
func (h *Handler) CheckLine(ctx context.Context, sb *pb.Surebet, i int64, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	side := sb.Members[i]
	side.Check = &pb.Check{Status: status.StatusError, Id: util.UnixMsNow()}
	response, err := h.clients[side.ServiceName].CheckLine(ctx, &pb.CheckLineRequest{Surebet: SurebetWithOneMember(sb, i)})
	if err != nil {
		side.Check.Done = util.UnixMsNow()
		st := gstatus.Convert(err)
		if st.Code() == codes.DeadlineExceeded {
			side.Check.Status = status.DeadlineExceeded
		} else {
			h.log.Errorw("check_line_error", "err", err, "name", side.ServiceName)
			side.Check.Status = status.ServiceError
			side.Check.StatusInfo = fmt.Sprintf("code:%q, msg:%q", st.Code(), st.Message())
		}
	} else if response.GetSide() != nil {
		sb.Members[i] = response.GetSide()
		sb.Members[i].Check.Done = util.UnixMsNow()
		if sb.Members[i].GetCheck().GetStatus() != status.StatusOk {
			h.log.Infow("check_line_resp", "name", side.ServiceName, "check", response.Side.Check, "marketName", side.MarketName, "sport", side.SportName, "league", side.LeagueName,
				"home", side.Home, "away", side.Away)
		}
	}
}

func (h *Handler) PlaceBet(ctx context.Context, sb *pb.Surebet, i int64, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	side := sb.Members[i]
	if side.ToBet == nil {
		side.ToBet = &pb.ToBet{Id: util.UnixMsNow()}
	} else {
		h.log.Info("add try count")
		side.ToBet.TryCount = side.ToBet.TryCount + 1
	}
	side.Bet = &pb.Bet{Status: status.StatusError, Start: util.UnixMsNow()}
	response, err := h.clients[side.ServiceName].PlaceBet(ctx, &pb.PlaceBetRequest{Surebet: SurebetWithOneMember(sb, i)})
	if err != nil {
		h.log.Errorw("place_bet_error", "err", err, "name", side.ServiceName, "fid", sb.FortedSurebetId)
		side.Bet.StatusInfo = "service error"
		side.Bet.Done = util.UnixMsNow()
		return
	} else if response.GetSide() != nil {
		sb.Members[i] = response.GetSide()
		sb.Members[i].Bet.Done = util.UnixMsNow()
		if sb.Members[i].GetBet().GetStatus() != status.StatusOk {
			h.log.Infow("place_bet_not_ok", "name", side.ServiceName, "bet", sb.Members[i].GetBet(), "fid", sb.FortedSurebetId)
		}
	}
}
func (h *Handler) SendPlaceFirst(ctx context.Context, sb *pb.Surebet) {
	var wg sync.WaitGroup
	go h.PlaceBet(ctx, sb, sb.Calc.FirstIndex, &wg)
	go h.CheckLine(ctx, sb, sb.Calc.SecondIndex, &wg)
	wg.Add(2)
	wg.Wait()
}
func (h *Handler) ProcessSurebet(sb *pb.Surebet) *SurebetError {
	ctx := context.Background()
	sb.SurebetId = util.UnixUsNow()
	if err := h.AllServicesActive(sb); err != nil {
		return err
	}
	if err := h.LoadConfig(ctx, sb); err != nil {
		return err
	}
	if err := h.AnyDisabled(sb); err != nil {
		return err
	}
	if err := h.GetCurrency(ctx, sb); err != nil {
		return err
	}
	checkCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
	h.SendCheckLines(checkCtx, sb)
	cancel()
	defer h.ReleaseChecks(sb)
	if err := AllCheckStatus(sb); err != nil {
		return err
	}
	if err := h.Calc(sb); err != nil {
		return err
	}
	if err := h.AllSurebet(sb); err != nil {
		return err
	}
	h.log.Infow("begin_bet_first", "profit", sb.Calc.Profit, "stake", sb.Members[sb.Calc.FirstIndex].CheckCalc.Stake, "time", ElapsedFromSurebetId(sb.SurebetId),
		"fid", sb.FortedSurebetId, "price", sb.Members[sb.Calc.FirstIndex].Check.Price, "name", sb.Members[sb.Calc.FirstIndex].ServiceName)
	h.SendPlaceFirst(ctx, sb)
	switch sb.Members[sb.Calc.FirstIndex].Bet.Status {
	case status.StatusOk:
	case status.AboveEventMax:
		return &SurebetError{Msg: "first_bet_not_ok", Permanent: true, ServiceName: sb.Members[sb.Calc.FirstIndex].ServiceName}
	case status.MarketClosed:
		return &SurebetError{Msg: "MarketClosed", Permanent: true, ServiceName: sb.Members[sb.Calc.FirstIndex].ServiceName}
	default:
		return &SurebetError{Msg: "first_bet_not_ok", Permanent: false, ServiceName: sb.Members[sb.Calc.FirstIndex].ServiceName}
	}
	h.ReleaseCheck(sb.Members[sb.Calc.FirstIndex].ServiceName, sb.SurebetId, nil)
	for i := 0; i < secondBetMaxTry; i++ {
		if sb.Members[sb.Calc.SecondIndex].Check.GetStatus() != status.StatusOk {
			time.Sleep(time.Millisecond * 100 * time.Duration(i))
			h.CheckLine(ctx, sb, sb.Calc.SecondIndex, nil)
			continue
		}
		h.CalcSecond(sb)
		h.PlaceBet(ctx, sb, sb.Calc.SecondIndex, nil)
		if sb.Members[sb.Calc.SecondIndex].GetBet().GetStatus() == status.StatusOk {
			break
		} else if i < secondBetMaxTry {
			sb.Members[sb.Calc.SecondIndex].Check.Status = status.StatusError
		}
	}
	if sb.Members[sb.Calc.SecondIndex].GetBet().GetStatus() != status.StatusOk {
		h.log.Infow("second bet not ok", "name", sb.Members[sb.Calc.SecondIndex].ServiceName)
		go h.tel.Sendf("oblom v=%v, p=%v, f=%v, s=%v, t=%v", betAmount(sb), sb.Calc.Profit, sb.Members[0].ServiceName, sb.Members[1].ServiceName, ElapsedFromSurebetId(sb.SurebetId))
		return &SurebetError{Msg: "second_bet_not_ok", Permanent: true, ServiceName: sb.Members[sb.Calc.SecondIndex].ServiceName}
	}
	return nil
}

func (h *Handler) SaveSurebet(sb *pb.Surebet) {
	if !h.HasAnyBet(sb) {
		return
	}
	if err := h.store.SaveFortedSurebet(sb); err != nil {
		h.log.Error(err)
	}
	//h.log.Infow("save surebet", "profit", sb.Calc.Profit, "time", ElapsedFromSurebetId(sb.SurebetId))
	if err := h.store.SaveCalc(sb); err != nil {
		h.log.Error(err)
	}
	//h.log.Infow("save sides", "0", sb.Members[0], "1", sb.Members[1])
	if err := h.store.SaveSide(sb); err != nil {
		h.log.Error(err)
	}
}
func (h *Handler) SurebetLoop(sb *pb.Surebet) {
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
		h.SaveSurebet(sb)
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
			h.log.Infow("placed_middle", "profit", sb.Calc.Profit, "time", ElapsedFromSurebetId(sb.SurebetId), "loop", i, "fid", sb.FortedSurebetId)
			i = 0
		}
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
