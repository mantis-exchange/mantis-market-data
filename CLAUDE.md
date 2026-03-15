# mantis-market-data

Mantis Exchange market data aggregation service — K-lines, trades, depth snapshots.

## Architecture

- `internal/model/kline.go` — Kline + TradeRecord models, PostgreSQL CRUD
- `internal/service/kline.go` — KlineAggregator: builds candles from trades (1m/5m/15m/1h/4h/1d)
- `internal/service/depth.go` — In-memory depth snapshot per symbol
- `internal/handler/handler.go` — REST API handlers
- `internal/consumer/trade.go` — Kafka consumer placeholder

## API Endpoints

- `GET /api/v1/klines?symbol=BTC-USDT&interval=1m&limit=100`
- `GET /api/v1/trades?symbol=BTC-USDT&limit=50`
- `GET /api/v1/depth?symbol=BTC-USDT`

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP server port |
| `DB_URL` | `postgres://mantis:mantis@localhost:5432/mantis_market?sslmode=disable` | PostgreSQL |
| `KAFKA_BROKERS` | `localhost:9092` | Kafka brokers |
| `MATCHING_ENGINE_ADDR` | `localhost:50051` | Matching engine gRPC |
