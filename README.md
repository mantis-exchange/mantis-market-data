# mantis-market-data

Market data aggregation service for [Mantis Exchange](https://github.com/mantis-exchange). Consumes trade events from Kafka, builds K-line candlesticks, and serves market data via REST API.

## Features

- **K-line aggregation** — 1m, 5m, 15m, 1h, 4h, 1d intervals
- **Trade history** — persisted to PostgreSQL
- **Depth snapshots** — in-memory order book depth
- **24h tickers** — last price, high, low, volume, change
- **Kafka consumer** — real-time trade event processing

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/klines?symbol=BTC-USDT&interval=1m&limit=100` | K-line data |
| GET | `/api/v1/trades?symbol=BTC-USDT&limit=50` | Recent trades |
| GET | `/api/v1/depth?symbol=BTC-USDT` | Order book depth |
| GET | `/api/v1/tickers` | 24h tickers for all symbols |

## Quick Start

```bash
go build -o mantis-market-data ./cmd/marketdata
./mantis-market-data
```

## Part of [Mantis Exchange](https://github.com/mantis-exchange)

MIT License
