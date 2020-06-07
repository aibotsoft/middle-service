package handler

import (
	"github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/util"
)

func ClearSurebet(sb *fortedpb.Surebet) {
	sb.Calc = nil
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
func betAmount(sb *fortedpb.Surebet) int {
	var amount float64
	for i := range sb.Members {
		amount = amount + sb.Members[i].GetBet().GetStake()
	}
	return int(amount)
}
func (h *Handler) AllServicesActive(sb *fortedpb.Surebet) *SurebetError {
	for i := range sb.Members {
		if h.clients[sb.Members[i].ServiceName] == nil {
			return &SurebetError{Msg: "service_not_active", Permanent: true, ServiceName: sb.Members[i].ServiceName}
		}
	}
	return nil
}
