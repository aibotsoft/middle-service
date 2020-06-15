module github.com/aibotsoft/middle-service

go 1.14

require (
	github.com/aibotsoft/gen v0.0.0-20200531091936-c4d5d714bf82
	github.com/aibotsoft/micro v0.0.0-20200531091141-36c4ab85b13e
	github.com/dgraph-io/ristretto v0.0.2
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/jmoiron/sqlx v1.2.0
	github.com/nats-io/nats-server/v2 v2.1.7 // indirect
	github.com/nats-io/nats.go v1.10.0
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.2.0
	github.com/stretchr/testify v1.6.0
	go.uber.org/zap v1.15.0
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0 // indirect
)

replace github.com/aibotsoft/micro => ../micro

replace github.com/aibotsoft/gen => ../gen
