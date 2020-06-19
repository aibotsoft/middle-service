package store

import (
	"context"
	"database/sql"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/cache"
	"github.com/aibotsoft/micro/config"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/dgraph-io/ristretto"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type Store struct {
	cfg   *config.Config
	log   *zap.SugaredLogger
	db    *sqlx.DB
	Cache *ristretto.Cache
}

func New(cfg *config.Config, log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{log: log, db: db, Cache: cache.NewCache(cfg)}
}
func (s *Store) Close() {
	err := s.db.Close()
	if err != nil {
		s.log.Error(err)
	}
	s.Cache.Close()
}

func (s *Store) GetConfigByName(ctx context.Context, name string) (conf pb.BetConfig, err error) {
	got, b := s.Cache.Get("config:" + name)
	if b {
		return got.(pb.BetConfig), nil
	}
	err = s.db.GetContext(ctx, &conf, "SELECT * FROM dbo.BetConfig WHERE ServiceName = @p1", name)
	if err != nil {
		return
	}
	s.Cache.SetWithTTL("config:"+name, conf, 1, time.Minute)
	return
}
func (s *Store) SaveFortedSurebet(sb *pb.Surebet) error {
	_, err := s.db.Exec("dbo.uspSaveFortedSurebet",
		sql.Named("CreatedAt", sb.CreatedAt),
		sql.Named("Starts", sb.Starts),
		sql.Named("FortedHome", sb.FortedHome),
		sql.Named("FortedAway", sb.FortedAway),
		sql.Named("FortedProfit", sb.FortedProfit),
		sql.Named("FortedSport", sb.FortedSport),
		sql.Named("FortedLeague", sb.FortedLeague),
		sql.Named("FilterName", sb.FilterName),
		sql.Named("FortedSurebetId", sb.FortedSurebetId))
	if err != nil {
		return errors.Wrap(err, "uspSaveFortedSurebet error")
	}
	return nil
}
func (s *Store) SaveCalc(sb *pb.Surebet) error {
	_, err := s.db.Exec("dbo.uspSaveCalc",
		sql.Named("SurebetId", sb.SurebetId),
		sql.Named("FortedSurebetId", sb.FortedSurebetId),
		sql.Named("Profit", sb.Calc.Profit),
		sql.Named("EffectiveProfit", sb.Calc.EffectiveProfit),
		sql.Named("MiddleDiff", sb.Calc.MiddleDiff),
		sql.Named("MiddleMargin", sb.Calc.MiddleMargin),
		sql.Named("HoursBeforeEvent", sb.Calc.HoursBeforeEvent),
		sql.Named("Gross", sb.Calc.Gross),
		sql.Named("SurebetType", sb.Calc.SurebetType),
		sql.Named("FirstName", sb.Calc.FirstName),
		sql.Named("SecondName", sb.Calc.SecondName),
		sql.Named("LowerWinIndex", sb.Calc.LowerWinIndex),
		sql.Named("HigherWinIndex", sb.Calc.HigherWinIndex),
		sql.Named("FirstIndex", sb.Calc.FirstIndex),
		sql.Named("SecondIndex", sb.Calc.SecondIndex),
		sql.Named("WinDiff", sb.Calc.WinDiff),
		sql.Named("WinDiffRel", sb.Calc.WinDiffRel),
	)
	if err != nil {
		return errors.Wrap(err, "uspSaveCalc error")
	}
	return nil
}

func (s *Store) SaveSide(sb *pb.Surebet) error {
	for i, side := range sb.Members {
		_, err := s.db.Exec("dbo.uspSaveSide",
			sql.Named("SurebetId", sb.SurebetId),
			sql.Named("SideIndex", i),
			sql.Named("ServiceName", side.ServiceName),
			sql.Named("SportName", side.SportName),
			sql.Named("LeagueName", side.LeagueName),
			sql.Named("Home", side.Home),
			sql.Named("Away", side.Away),
			sql.Named("MarketName", side.MarketName),
			sql.Named("Price", side.Price),
			sql.Named("Initiator", side.Initiator),
			sql.Named("Starts", side.Starts),
			sql.Named("EventId", side.EventId),

			sql.Named("CheckId", side.Check.Id),
			sql.Named("AccountId", side.Check.AccountId),
			sql.Named("AccountLogin", side.Check.AccountLogin),
			sql.Named("CheckStatus", side.Check.Status),
			sql.Named("StatusInfo", side.Check.StatusInfo),
			sql.Named("CountLine", side.Check.CountLine),
			sql.Named("CountEvent", side.Check.CountEvent),
			sql.Named("AmountEvent", side.Check.AmountEvent),
			sql.Named("MinBet", side.Check.MinBet),
			sql.Named("MaxBet", side.Check.MaxBet),
			sql.Named("Balance", side.Check.Balance),
			sql.Named("CheckPrice", side.Check.Price),
			sql.Named("Currency", side.Check.Currency),
			sql.Named("CheckDone", side.Check.Done),
			sql.Named("MiddleMargin", side.Check.MiddleMargin),

			sql.Named("CalcStatus", side.GetCheckCalc().GetStatus()),
			sql.Named("MaxStake", side.GetCheckCalc().GetMaxStake()),
			sql.Named("MinStake", side.GetCheckCalc().GetMinStake()),
			sql.Named("MaxWin", side.GetCheckCalc().GetMaxWin()),
			sql.Named("Stake", side.GetCheckCalc().GetStake()),
			sql.Named("Win", side.GetCheckCalc().GetWin()),
			sql.Named("IsFirst", side.GetCheckCalc().GetIsFirst()),

			sql.Named("ToBetId", side.GetToBet().GetId()),
			sql.Named("TryCount", side.GetToBet().GetTryCount()),

			sql.Named("BetStatus", side.GetBet().GetStatus()),
			sql.Named("BetStatusInfo", side.GetBet().GetStatusInfo()),
			sql.Named("Start", side.GetBet().GetStart()),
			sql.Named("Done", side.GetBet().GetDone()),
			sql.Named("BetPrice", side.GetBet().GetPrice()),
			sql.Named("BetStake", side.GetBet().GetStake()),
			sql.Named("ApiBetId", side.GetBet().GetApiBetId()),
		)
		if err != nil {
			return errors.Wrap(err, "uspSaveCalc error")
		}
	}
	return nil
}
func (s *Store) SaveBetList(results []pb.BetResult) error {
	if len(results) == 0 {
		return nil
	}
	return nil
	tvp := mssql.TVP{TypeName: "BetListType", Value: results}

	_, err := s.db.Exec("dbo.uspSaveBetList", tvp)
	if err != nil {
		return errors.Wrap(err, "uspSaveResults error")
	}
	return nil
}
