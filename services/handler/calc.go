package handler

import (
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"time"
)

const (
	maxWinDiffPercent = 7
	profitCoefficient = 0.02
	minGross          = 0.12
	baseMiddleMargin  = 2

	ProfitTooLow     = "Profit_too_low"
	ProfitTooHigh    = "Profit_too_high"
	CountLineLimit   = "CountLine_reached_MaxCountLine"
	CountEventLimit  = "CountEvent_reached_MaxCountEvent"
	AmountEventLimit = "AmountEvent_reached_MaxAmountEvent"
	MaxStakeTooLow   = "MaxStake_lower_MinStake"
	MinStakeTooHigh  = "MinStake_higher_MaxStake"
	WinDiffTooHigh   = "WinDiffRel_too_high"
	GrossTooLow      = "gross_too_Low"
)

func (h *Handler) Calc(sb *pb.Surebet) *SurebetError {
	sb.Calc.Profit = Profit(sb)
	starts, err2 := time.Parse(util.ISOFormat, sb.Starts)
	if err2 == nil {
		sb.Calc.HoursBeforeEvent = util.TruncateFloat(starts.Sub(time.Now()).Hours(), 2)
	}
	sb.Calc.EffectiveProfit = util.TruncateFloat(sb.Calc.Profit-sb.Calc.HoursBeforeEvent*profitCoefficient, 2)
	//sb.FortedSport
	sb.Calc.MiddleMargin = baseMiddleMargin
	for i := range sb.Members {
		if sb.Members[i].Check.MiddleMargin > sb.Calc.MiddleMargin {
			sb.Calc.MiddleMargin = sb.Members[i].Check.MiddleMargin
		}
	}
	middleCorrection := sb.Calc.MiddleDiff * sb.Calc.MiddleMargin
	sb.Calc.EffectiveProfit = util.TruncateFloat(sb.Calc.EffectiveProfit+middleCorrection, 2)
	h.log.Debugw("middle_calc", "middleMargin", sb.Calc.MiddleMargin, "middleCorrection", middleCorrection, "MiddleDiff", sb.Calc.MiddleDiff, "EP", sb.Calc.EffectiveProfit, "sport", sb.FortedSport)

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

		} else if sb.Calc.EffectiveProfit < m.BetConfig.MinPercent {
			m.CheckCalc.Status = ProfitTooLow
			err = &SurebetError{Msg: fmt.Sprintf("%s, Profit:%v, EP:%v, min:%v", m.CheckCalc.Status, sb.Calc.Profit, sb.Calc.EffectiveProfit, m.BetConfig.MinPercent),
				Permanent:   false,
				ServiceName: m.ServiceName}

		} else if sb.Calc.EffectiveProfit > float64(m.BetConfig.MaxPercent) {
			m.CheckCalc.Status = ProfitTooHigh
			err = &SurebetError{Msg: fmt.Sprintf("%s, Profit:%v, EP:%v, max:%v", m.CheckCalc.Status, sb.Calc.Profit, sb.Calc.EffectiveProfit, m.BetConfig.MaxPercent),
				Permanent: false, ServiceName: m.ServiceName}

		} else if m.CheckCalc.MaxStake < m.CheckCalc.MinStake {
			m.CheckCalc.Status = MaxStakeTooLow
			err = &SurebetError{Msg: fmt.Sprintf("%s, MaxStake: %v, MinStake: %v", m.CheckCalc.Status, m.CheckCalc.MaxStake, m.CheckCalc.MinStake), Permanent: false, ServiceName: m.ServiceName}
		} else if m.CheckCalc.MinStake > m.CheckCalc.MaxStake {
			m.CheckCalc.Status = MinStakeTooHigh
			err = &SurebetError{Msg: fmt.Sprintf("%s, MinStake: %v, MinStake: %v", m.CheckCalc.Status, m.CheckCalc.MinStake, m.CheckCalc.MaxStake), Permanent: false, ServiceName: m.ServiceName}
		}
		//h.log.Debug(m.Starts)
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
	sb.Calc.Gross = util.TruncateFloat(a.CheckCalc.Win*sb.Calc.EffectiveProfit/100, 2)
	//if sb.Calc.Gross < minGross {
	//	a.CheckCalc.Status = GrossTooLow
	//	return &SurebetError{Msg: fmt.Sprintf("GrossTooLow:%v, minGross:%v", sb.Calc.Gross, minGross), Permanent: false, ServiceName: a.ServiceName}
	//}

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
