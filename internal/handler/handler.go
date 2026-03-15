package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/mantis-exchange/mantis-market-data/internal/model"
	"github.com/mantis-exchange/mantis-market-data/internal/service"
)

type Handler struct {
	klineRepo *model.KlineRepo
	depth     *service.DepthService
}

func New(klineRepo *model.KlineRepo, depth *service.DepthService) *Handler {
	return &Handler{klineRepo: klineRepo, depth: depth}
}

func (h *Handler) GetKlines(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.DefaultQuery("interval", "1m")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	klines, err := h.klineRepo.List(c.Request.Context(), symbol, interval, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"klines": klines})
}

func (h *Handler) GetTrades(c *gin.Context) {
	symbol := c.Query("symbol")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	trades, err := h.klineRepo.ListTrades(c.Request.Context(), symbol, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trades": trades})
}

func (h *Handler) GetDepth(c *gin.Context) {
	symbol := c.Query("symbol")
	depth := h.depth.Get(symbol)
	c.JSON(http.StatusOK, depth)
}
