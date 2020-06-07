package store

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/cache"
	"github.com/aibotsoft/micro/config"
	"github.com/dgraph-io/ristretto"
	"github.com/jmoiron/sqlx"
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
func (s *Store) LoadConfig(ctx context.Context, sb *pb.Surebet) error {
	for i := range sb.Members {
		conf, err := s.GetConfigByName(ctx, sb.Members[i].ServiceName)
		if err != nil {
			return err
		}
		sb.Members[i].BetConfig = &conf
	}
	return nil
}
