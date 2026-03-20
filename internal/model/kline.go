package model

import (
	"context"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Ticker struct {
	Symbol    string `json:"symbol"`
	LastPrice string `json:"last_price"`
	High24h   string `json:"high_24h"`
	Low24h    string `json:"low_24h"`
	Volume24h string `json:"volume_24h"`
	Change24h string `json:"change_24h"`
}

type Kline struct {
	Symbol    string    `json:"symbol"`
	Interval  string    `json:"interval"`
	OpenTime  time.Time `json:"open_time"`
	Open      string    `json:"open"`
	High      string    `json:"high"`
	Low       string    `json:"low"`
	Close     string    `json:"close"`
	Volume    string    `json:"volume"`
	CloseTime time.Time `json:"close_time"`
}

type TradeRecord struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Price     string    `json:"price"`
	Quantity  string    `json:"quantity"`
	MakerSide string   `json:"maker_side"`
	CreatedAt time.Time `json:"created_at"`
}

type KlineRepo struct {
	pool *pgxpool.Pool
}

func NewKlineRepo(pool *pgxpool.Pool) *KlineRepo {
	return &KlineRepo{pool: pool}
}

func (r *KlineRepo) Upsert(ctx context.Context, k *Kline) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO klines (symbol, interval, open_time, open, high, low, close, volume, close_time)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (symbol, interval, open_time)
		 DO UPDATE SET high = $5, low = $6, close = $7, volume = $8, close_time = $9`,
		k.Symbol, k.Interval, k.OpenTime, k.Open, k.High, k.Low, k.Close, k.Volume, k.CloseTime,
	)
	return err
}

func (r *KlineRepo) List(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT symbol, interval, open_time, open, high, low, close, volume, close_time
		 FROM klines WHERE symbol = $1 AND interval = $2
		 ORDER BY open_time DESC LIMIT $3`, symbol, interval, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var klines []Kline
	for rows.Next() {
		var k Kline
		if err := rows.Scan(&k.Symbol, &k.Interval, &k.OpenTime, &k.Open, &k.High, &k.Low, &k.Close, &k.Volume, &k.CloseTime); err != nil {
			return nil, err
		}
		klines = append(klines, k)
	}
	// Reverse to ascending order
	for i, j := 0, len(klines)-1; i < j; i, j = i+1, j-1 {
		klines[i], klines[j] = klines[j], klines[i]
	}
	return klines, nil
}

func (r *KlineRepo) InsertTrade(ctx context.Context, t *TradeRecord) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO trades (id, symbol, price, quantity, maker_side, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO NOTHING`,
		t.ID, t.Symbol, t.Price, t.Quantity, t.MakerSide, t.CreatedAt,
	)
	return err
}

func (r *KlineRepo) ListTrades(ctx context.Context, symbol string, limit int) ([]TradeRecord, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, symbol, price, quantity, maker_side, created_at
		 FROM trades WHERE symbol = $1
		 ORDER BY created_at DESC LIMIT $2`, symbol, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []TradeRecord
	for rows.Next() {
		var t TradeRecord
		if err := rows.Scan(&t.ID, &t.Symbol, &t.Price, &t.Quantity, &t.MakerSide, &t.CreatedAt); err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, nil
}

func (r *KlineRepo) GetTickers(ctx context.Context) ([]Ticker, error) {
	rows, err := r.pool.Query(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (symbol) symbol, close as last_price
			FROM klines WHERE interval = '1m'
			ORDER BY symbol, open_time DESC
		),
		stats AS (
			SELECT symbol,
				MAX(high) as high_24h,
				MIN(low) as low_24h,
				SUM(volume::numeric) as volume_24h,
				(SELECT open FROM klines k2
				 WHERE k2.symbol = k1.symbol AND k2.interval = '1h'
				   AND k2.open_time >= NOW() - INTERVAL '24 hours'
				 ORDER BY k2.open_time LIMIT 1) as open_24h
			FROM klines k1
			WHERE interval = '1h' AND open_time >= NOW() - INTERVAL '24 hours'
			GROUP BY symbol
		)
		SELECT l.symbol, l.last_price,
			COALESCE(s.high_24h, l.last_price) as high_24h,
			COALESCE(s.low_24h, l.last_price) as low_24h,
			COALESCE(s.volume_24h::text, '0') as volume_24h,
			COALESCE(s.open_24h, l.last_price) as open_24h
		FROM latest l
		LEFT JOIN stats s ON l.symbol = s.symbol
		ORDER BY l.symbol
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickers []Ticker
	for rows.Next() {
		var t Ticker
		var open24h string
		if err := rows.Scan(&t.Symbol, &t.LastPrice, &t.High24h, &t.Low24h, &t.Volume24h, &open24h); err != nil {
			return nil, err
		}
		lastF, _, _ := new(big.Float).SetPrec(64).Parse(t.LastPrice, 10)
		openF, _, _ := new(big.Float).SetPrec(64).Parse(open24h, 10)
		if lastF != nil && openF != nil && openF.Sign() > 0 {
			change := new(big.Float).Sub(lastF, openF)
			change.Quo(change, openF)
			change.Mul(change, new(big.Float).SetFloat64(100))
			t.Change24h = change.Text('f', 2)
		} else {
			t.Change24h = "0.00"
		}
		tickers = append(tickers, t)
	}
	return tickers, nil
}
