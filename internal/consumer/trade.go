package consumer

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/mantis-exchange/mantis-market-data/internal/model"
	"github.com/mantis-exchange/mantis-market-data/internal/service"
)

const tradeTopic = "mantis.trades"

type tradeMessage struct {
	ID           string `json:"id"`
	Symbol       string `json:"symbol"`
	Price        string `json:"price"`
	Quantity     string `json:"quantity"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	MakerSide    string `json:"maker_side"`
	CreatedAt    int64  `json:"created_at"`
}

type TradeConsumer struct {
	repo       *model.KlineRepo
	aggregator *service.KlineAggregator
	depth      *service.DepthService
	brokers    string
}

func New(repo *model.KlineRepo, agg *service.KlineAggregator, depth *service.DepthService, brokers string) *TradeConsumer {
	return &TradeConsumer{repo: repo, aggregator: agg, depth: depth, brokers: brokers}
}

func (c *TradeConsumer) Start() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(c.brokers, ","),
		Topic:       tradeTopic,
		GroupID:     "mantis-market-data",
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	log.Printf("market-data trade consumer started (brokers: %s, topic: %s)", c.brokers, tradeTopic)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("market-data consumer read error: %v", err)
			continue
		}

		var trade tradeMessage
		if err := json.Unmarshal(msg.Value, &trade); err != nil {
			log.Printf("failed to unmarshal trade: %v", err)
			continue
		}

		c.processTrade(context.Background(), trade)
	}
}

func (c *TradeConsumer) processTrade(ctx context.Context, trade tradeMessage) {
	tradeTime := time.UnixMilli(trade.CreatedAt)

	// Insert trade record into DB
	record := &model.TradeRecord{
		ID:        trade.ID,
		Symbol:    trade.Symbol,
		Price:     trade.Price,
		Quantity:  trade.Quantity,
		MakerSide: trade.MakerSide,
		CreatedAt: tradeTime,
	}
	if err := c.repo.InsertTrade(ctx, record); err != nil {
		log.Printf("failed to insert trade %s: %v", trade.ID, err)
	}

	// Feed into KlineAggregator
	c.aggregator.ProcessTrade(ctx, trade.Symbol, trade.Price, trade.Quantity, tradeTime)
}
