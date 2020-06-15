package handler

import (
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/status"
)

const (
	maxWinDiffPercent = 7
	ProfitTooLow      = "Profit_lower_MinPercent"
	ProfitTooHigh     = "Profit_higher_MaxPercent"
	CountLineLimit    = "CountLine has reached MaxCountLine"
	CountEventLimit   = "CountEvent has reached MaxCountEvent"
	AmountEventLimit  = "AmountEvent has reached MaxAmountEvent"
	MaxStakeTooLow    = "MaxStake_lower_MinStake"
	MinStakeTooHigh   = "MinStake_higher_MaxStake"
	WinDiffTooHigh    = "WinDiffRel_too_high"
)

func (h *Handler) Calc(sb *pb.Surebet) *SurebetError {
	sb.Calc.Profit = Profit(sb)
	var err *SurebetError
	for i := range sb.Members {
		m := sb.Members[i]
		m.CheckCalc = &pb.CheckCalc{}
		if m.Check.Price == 0 {
			return &SurebetError{Msg: "Check.Price is 0", ServiceName: m.ServiceName}
		}
		CalcMaxStake(m)
		CalcMinStake(m)
		CalcMaxWin(m)
		m.CheckCalc.Stake = m.CheckCalc.MaxStake
		CalcWin(m)
		m.CheckCalc.Status = status.StatusOk
		if m.Check.CountLine >= m.BetConfig.MaxCountLine {
			m.CheckCalc.Status = CountLineLimit
			err = &SurebetError{Msg: fmt.Sprintf("%s, CountLine: %v, MaxCountLine: %v", m.CheckCalc.Status, m.Check.CountLine, m.BetConfig.MaxCountLine), Permanent: true, ServiceName: m.ServiceName}
		} else if m.Check.CountEvent >= m.BetConfig.MaxCountEvent {
			m.CheckCalc.Status = CountEventLimit
			err = &SurebetError{Msg: fmt.Sprintf("%s, CountEvent: %v, MaxCountEvent: %v", m.CheckCalc.Status, m.Check.CountEvent, m.BetConfig.MaxCountEvent), Permanent: true, ServiceName: m.ServiceName}
		} else if m.Check.AmountEvent >= m.BetConfig.MaxAmountEvent {
			m.CheckCalc.Status = AmountEventLimit
			err = &SurebetError{Msg: fmt.Sprintf("%s, AmountEvent: %v, MaxAmountEvent: %v", m.CheckCalc.Status, m.Check.AmountEvent, m.BetConfig.MaxAmountEvent), Permanent: true, ServiceName: m.ServiceName}
		} else if sb.Calc.Profit < m.BetConfig.MinPercent {
			m.CheckCalc.Status = ProfitTooLow
			err = &SurebetError{Msg: fmt.Sprintf("%s, Profit: %v, minPercent: %v", m.CheckCalc.Status, sb.Calc.Profit, m.BetConfig.MinPercent), Permanent: false, ServiceName: m.ServiceName}
		} else if sb.Calc.Profit > float64(m.BetConfig.MaxPercent) {
			m.CheckCalc.Status = ProfitTooHigh
			err = &SurebetError{Msg: fmt.Sprintf("%s, Profit: %v, MaxPercent: %v", m.CheckCalc.Status, sb.Calc.Profit, m.BetConfig.MaxPercent), Permanent: false, ServiceName: m.ServiceName}
		} else if m.CheckCalc.MaxStake < m.CheckCalc.MinStake {
			m.CheckCalc.Status = MaxStakeTooLow
			err = &SurebetError{Msg: fmt.Sprintf("%s, MaxStake: %v, MinStake: %v", m.CheckCalc.Status, m.CheckCalc.MaxStake, m.CheckCalc.MinStake), Permanent: false, ServiceName: m.ServiceName}
		} else if m.CheckCalc.MinStake > m.CheckCalc.MaxStake {
			m.CheckCalc.Status = MinStakeTooHigh
			err = &SurebetError{Msg: fmt.Sprintf("%s, MinStake: %v, MinStake: %v", m.CheckCalc.Status, m.CheckCalc.MinStake, m.CheckCalc.MaxStake), Permanent: false, ServiceName: m.ServiceName}
		}
		if err != nil && err.Permanent {
			return err
		}
	}
	if err != nil {
		return err
	}
	FirstSecond(sb)
	a := sb.Members[sb.Calc.LowerWinIndex]
	b := sb.Members[sb.Calc.HigherWinIndex]
	b.CheckCalc.Stake = CalcStake(a.CheckCalc.Win, b.Check.Price)
	if b.CheckCalc.Stake < b.CheckCalc.MinStake {
		b.CheckCalc.Stake = b.CheckCalc.MinStake
	}
	CalcWin(b)
	sb.Calc.WinDiff = CalcWinDiff(a.CheckCalc.Win, b.CheckCalc.Win)
	sb.Calc.WinDiffRel = CalcWinDiffRel(a.CheckCalc.Win, b.CheckCalc.Win)
	if sb.Calc.WinDiffRel > maxWinDiffPercent {
		b.CheckCalc.Status = WinDiffTooHigh
		err = &SurebetError{Msg: fmt.Sprintf("%s, WinDiffRel: %v, WinDiff: %v", b.CheckCalc.Status, sb.Calc.WinDiffRel, sb.Calc.WinDiff), Permanent: false, ServiceName: b.ServiceName}
	}
	return err
}

func (h *Handler) CalcSecond(sb *pb.Surebet) {
	a := sb.Members[sb.Calc.FirstIndex]
	b := sb.Members[sb.Calc.SecondIndex]
	CalcMaxStake(b)
	CalcMinStake(b)
	CalcMaxWin(b)
	b.CheckCalc.Stake = a.Bet.Stake * a.Bet.Price / b.Check.Price
	if b.CheckCalc.Stake < b.CheckCalc.MinStake {
		b.CheckCalc.Stake = b.CheckCalc.MinStake
	}
	CalcWin(b)
	b.CheckCalc.Status = status.StatusOk
}
