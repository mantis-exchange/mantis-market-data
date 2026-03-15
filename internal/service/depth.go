package service

import (
	"sync"
)

type DepthLevel struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

type Depth struct {
	Symbol string       `json:"symbol"`
	Bids   []DepthLevel `json:"bids"`
	Asks   []DepthLevel `json:"asks"`
}

// DepthService maintains in-memory depth snapshots per symbol.
type DepthService struct {
	mu     sync.RWMutex
	depths map[string]*Depth
}

func NewDepthService() *DepthService {
	return &DepthService{
		depths: make(map[string]*Depth),
	}
}

func (s *DepthService) Update(symbol string, bids, asks []DepthLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.depths[symbol] = &Depth{Symbol: symbol, Bids: bids, Asks: asks}
}

func (s *DepthService) Get(symbol string) *Depth {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.depths[symbol]
	if !ok {
		return &Depth{Symbol: symbol}
	}
	return d
}
