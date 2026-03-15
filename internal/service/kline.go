package service

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/mantis-exchange/mantis-market-data/internal/model"
)

var intervals = []string{"1m", "5m", "15m", "1h", "4h", "1d"}

func intervalDuration(interval string) time.Duration {
	switch interval {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return time.Minute
	}
}

type candleKey struct {
	Symbol   string
	Interval string
}

type candle struct {
	Open     string
	High     string
	Low      string
	Close    string
	Volume   *big.Float
	OpenTime time.Time
}

type KlineAggregator struct {
	mu      sync.Mutex
	candles map[candleKey]*candle
	repo    *model.KlineRepo
}

func NewKlineAggregator(repo *model.KlineRepo) *KlineAggregator {
	return &KlineAggregator{
		candles: make(map[candleKey]*candle),
		repo:    repo,
	}
}

// ProcessTrade updates all interval candles with a new trade.
func (a *KlineAggregator) ProcessTrade(ctx context.Context, symbol, price, quantity string, tradeTime time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	qty, _, _ := new(big.Float).Parse(quantity, 10)
	if qty == nil {
		qty = new(big.Float)
	}

	for _, iv := range intervals {
		dur := intervalDuration(iv)
		openTime := tradeTime.Truncate(dur)
		key := candleKey{Symbol: symbol, Interval: iv}

		c, exists := a.candles[key]
		if !exists || c.OpenTime != openTime {
			// Flush previous candle if it exists
			if exists && c.OpenTime != openTime {
				a.flush(ctx, symbol, iv, c)
			}
			// Start new candle
			a.candles[key] = &candle{
				Open:     price,
				High:     price,
				Low:      price,
				Close:    price,
				Volume:   new(big.Float).Set(qty),
				OpenTime: openTime,
			}
		} else {
			c.Close = price
			if comparePrices(price, c.High) > 0 {
				c.High = price
			}
			if comparePrices(price, c.Low) < 0 {
				c.Low = price
			}
			c.Volume.Add(c.Volume, qty)
		}
	}
}

func (a *KlineAggregator) flush(ctx context.Context, symbol, interval string, c *candle) {
	dur := intervalDuration(interval)
	k := &model.Kline{
		Symbol:    symbol,
		Interval:  interval,
		OpenTime:  c.OpenTime,
		Open:      c.Open,
		High:      c.High,
		Low:       c.Low,
		Close:     c.Close,
		Volume:    c.Volume.Text('f', 8),
		CloseTime: c.OpenTime.Add(dur),
	}
	_ = a.repo.Upsert(ctx, k)
}

// FlushAll persists all current candles to the database.
func (a *KlineAggregator) FlushAll(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for key, c := range a.candles {
		a.flush(ctx, key.Symbol, key.Interval, c)
	}
}

func comparePrices(a, b string) int {
	fa, _, _ := new(big.Float).Parse(a, 10)
	fb, _, _ := new(big.Float).Parse(b, 10)
	if fa == nil || fb == nil {
		return 0
	}
	return fa.Cmp(fb)
}
