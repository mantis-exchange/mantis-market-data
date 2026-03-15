package consumer

import (
	"log"

	"github.com/mantis-exchange/mantis-market-data/internal/model"
	"github.com/mantis-exchange/mantis-market-data/internal/service"
)

type TradeConsumer struct {
	repo       *model.KlineRepo
	aggregator *service.KlineAggregator
	depth      *service.DepthService
	brokers    string
}

func New(repo *model.KlineRepo, agg *service.KlineAggregator, depth *service.DepthService, brokers string) *TradeConsumer {
	return &TradeConsumer{repo: repo, aggregator: agg, depth: depth, brokers: brokers}
}

// Start begins consuming trade events from Kafka. Placeholder.
func (c *TradeConsumer) Start() {
	log.Printf("market-data trade consumer started (brokers: %s) — placeholder", c.brokers)
	// TODO: Connect to Kafka, consume trade events, and for each:
	// 1. Insert trade record into DB
	// 2. Feed into KlineAggregator
	// 3. Update DepthService
}
