package handler

import (
	"context"
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"time"
)

var d100 = decimal.NewFromInt(100)

func ClearSurebet(sb *pb.Surebet) {
	sb.Calc = pb.Calc{}
	for i := range sb.Members {
		sb.Members[i].Check = nil
		sb.Members[i].BetConfig = nil
		sb.Members[i].ToBet = nil
		sb.Members[i].Bet = nil
	}
}
func ElapsedFromSurebetId(surebetId int64) int64 {
	return (util.UnixUsNow() - surebetId) / 1000
}
func betAmount(sb *pb.Surebet) int {
	var amount float64
	for i := range sb.Members {
		amount = amount + sb.Members[i].GetBet().GetStake()
	}
	return int(amount)
}
func (h *Handler) AllServicesActive(sb *pb.Surebet) *SurebetError {
	for i := range sb.Members {
		if h.clients[sb.Members[i].ServiceName] == nil {
			return &SurebetError{Msg: "service_not_active", Permanent: true, ServiceName: sb.Members[i].ServiceName}
		}
	}
	return nil
}
func (h *Handler) AnyDisabled(sb *pb.Surebet) *SurebetError {
	for i := range sb.Members {
		if sb.Members[i].BetConfig == nil {
			return &SurebetError{Msg: "service_has_no_config", Permanent: true, ServiceName: sb.Members[i].ServiceName}
		}
		if sb.Members[i].BetConfig.GetRegime() == status.StatusDisabled {
			return &SurebetError{Msg: "service_disabled", Permanent: true, ServiceName: sb.Members[i].ServiceName}
		}
	}
	return nil
}
func (h *Handler) GetCurrency(ctx context.Context, sb *pb.Surebet) *SurebetError {
	get, b := h.store.Cache.Get("currency_list")
	if b {
		sb.Currency = get.([]pb.Currency)
		return nil
	}
	currency, err := h.Conf.GetCurrency(ctx)
	if err != nil {
		return &SurebetError{Err: err, Msg: "get_currency_error", Permanent: true}
	}
	err = copier.Copy(&sb.Currency, &currency)
	if err != nil {
		return &SurebetError{Err: err, Msg: "copy_currency_slice_error", Permanent: true}
	}
	h.store.Cache.SetWithTTL("currency_list", sb.Currency, 1, time.Hour)
	return nil
}
func (h *Handler) LoadConfig(ctx context.Context, sb *pb.Surebet) *SurebetError {
	for i := range sb.Members {
		conf, err := h.store.GetConfigByName(ctx, sb.Members[i].ServiceName)
		if err != nil {
			return &SurebetError{Err: err, Msg: "GetConfigByName_error", Permanent: true}
		}
		sb.Members[i].BetConfig = &conf
	}
	return nil
}
func SurebetWithOneMember(sb *pb.Surebet, i int64) *pb.Surebet {
	copySb := *sb
	copySb.Members = sb.Members[i : i+1]
	return &copySb
}

func AllCheckStatus(sb *pb.Surebet) *SurebetError {
	var err *SurebetError
	for i := 0; i < len(sb.Members); i++ {
		switch sb.Members[i].Check.Status {
		case status.StatusOk:
			continue
		case status.StatusNotFound:
			err = &SurebetError{Msg: "check_status: NotFound", Permanent: false, ServiceName: sb.Members[i].ServiceName}
		case status.StatusError:
			err = &SurebetError{Msg: "check_status: Error", Permanent: false, ServiceName: sb.Members[i].ServiceName}
		case status.BadBettingStatus:
			err = &SurebetError{Msg: status.BadBettingStatus, Permanent: true, ServiceName: sb.Members[i].ServiceName}
		case status.ServiceSportDisabled:
			err = &SurebetError{Msg: status.ServiceSportDisabled, Permanent: true, ServiceName: sb.Members[i].ServiceName}
		case status.PitchersRequired:
			err = &SurebetError{Msg: status.PitchersRequired, Permanent: true, ServiceName: sb.Members[i].ServiceName}
		case status.ServiceError:
			err = &SurebetError{Msg: status.ServiceError, Permanent: true, ServiceName: sb.Members[i].ServiceName}
		case status.MarketClosed:
			err = &SurebetError{Msg: status.MarketClosed, Permanent: true, ServiceName: sb.Members[i].ServiceName}
		case status.DeadlineExceeded:
			err = &SurebetError{Msg: status.DeadlineExceeded, Permanent: false, ServiceName: sb.Members[i].ServiceName}
		default:
			err = &SurebetError{Msg: "check_status not Ok", Permanent: false, ServiceName: sb.Members[i].ServiceName}
		}
		if err.Permanent {
			return err
		}
	}
	if sb.Members[0].Check.Status != status.StatusOk && sb.Members[1].Check.Status != status.StatusOk {
		return &SurebetError{Msg: "both_check_status_not_ok", Permanent: true, ServiceName: fmt.Sprintf("f:%s, s:%s", sb.Members[0].ServiceName, sb.Members[1].ServiceName)}
	}
	return err
}

func (h *Handler) AllSurebet(sb *pb.Surebet) *SurebetError {
	for i := range sb.Members {
		if sb.Members[i].BetConfig.Regime != status.RegimeSurebet {
			return &SurebetError{Msg: "regime_not_Surebet", Permanent: true, ServiceName: sb.Members[i].ServiceName}
		}
	}
	return nil
}
func Profit(sb *pb.Surebet) (prob float64) {
	for i := range sb.Members {
		prob += 1 / sb.Members[i].Check.Price
	}
	profit := 1/prob*100 - 100
	return util.TruncateFloat(profit, 3)
}
func CalcMaxStake(m *pb.SurebetSide) {
	m.CheckCalc.MaxStake, _ = decimal.Min(
		decimal.New(m.Check.Balance, 0),
		decimal.NewFromFloat(m.Check.MaxBet),
		decimal.New(m.BetConfig.MaxStake, 0),
		decimal.New(m.BetConfig.MaxWin, 0).DivRound(decimal.NewFromFloat(m.Check.Price), 3)).Float64()
}
func CalcMinStake(m *pb.SurebetSide) {
	m.CheckCalc.MinStake, _ = decimal.Max(decimal.NewFromFloat(m.Check.MinBet), decimal.NewFromInt(m.BetConfig.MinStake)).Float64()
}
func CalcMaxWin(m *pb.SurebetSide) {
	m.CheckCalc.MaxWin, _ = decimal.NewFromFloat(m.CheckCalc.MaxStake).Mul(decimal.NewFromFloat(m.Check.Price)).Float64()
}
func CalcWin(m *pb.SurebetSide) {
	m.CheckCalc.Win, _ = decimal.NewFromFloat(m.CheckCalc.Stake).Mul(decimal.NewFromFloat(m.Check.Price)).Round(5).Float64()
}
func FirstSecond(sb *pb.Surebet) {
	if sb.Members[0].BetConfig.Priority >= sb.Members[1].BetConfig.Priority {
		sb.Calc.FirstIndex = 0
		sb.Calc.SecondIndex = 1
	} else {
		sb.Calc.FirstIndex = 1
		sb.Calc.SecondIndex = 0
	}
	sb.Calc.FirstName = sb.Members[sb.Calc.FirstIndex].ServiceName
	sb.Calc.SecondName = sb.Members[sb.Calc.SecondIndex].ServiceName
	sb.Members[sb.Calc.FirstIndex].CheckCalc.IsFirst = true
	if sb.Members[0].CheckCalc.MaxWin <= sb.Members[1].CheckCalc.MaxWin {
		sb.Calc.LowerWinIndex = 0
		sb.Calc.HigherWinIndex = 1
	} else {
		sb.Calc.LowerWinIndex = 1
		sb.Calc.HigherWinIndex = 0
	}
}
func CalcStake(aWin float64, bPrice float64) float64 {
	f, _ := decimal.NewFromFloat(aWin).DivRound(decimal.NewFromFloat(bPrice), 5).Float64()
	return f
}
func CalcWinDiff(aWin float64, bWin float64) float64 {
	f, _ := decimal.NewFromFloat(aWin).Sub(decimal.NewFromFloat(bWin)).Abs().Float64()
	return f
}

func CalcWinDiffRel(aWin float64, bWin float64) float64 {
	aWinD := decimal.NewFromFloat(aWin)
	bWinD := decimal.NewFromFloat(bWin)
	sumWinD := aWinD.Add(bWinD)
	res, _ := aWinD.Sub(bWinD).Abs().Mul(d100).DivRound(sumWinD, 2).Float64()
	return res
}
