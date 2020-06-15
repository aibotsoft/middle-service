package store

import (
	"context"
	"database/sql"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/cache"
	"github.com/aibotsoft/micro/config"
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
		sql.Named("Profit", sb.Calc.Profit),
		sql.Named("FirstName", sb.Calc.FirstName),
		sql.Named("SecondName", sb.Calc.SecondName),
		sql.Named("LowerWinIndex", sb.Calc.LowerWinIndex),
		sql.Named("HigherWinIndex", sb.Calc.HigherWinIndex),
		sql.Named("FirstIndex", sb.Calc.FirstIndex),
		sql.Named("SecondIndex", sb.Calc.SecondIndex),
		sql.Named("WinDiff", sb.Calc.WinDiff),
		sql.Named("WinDiffRel", sb.Calc.WinDiffRel),
		sql.Named("FortedSurebetId", sb.FortedSurebetId),
		sql.Named("SurebetId", sb.SurebetId),
	)
	if err != nil {
		return errors.Wrap(err, "uspSaveCalc error")
	}
	return nil
}
