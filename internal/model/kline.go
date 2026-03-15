package model

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
