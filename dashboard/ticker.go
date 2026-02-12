package dashboard

import (
	"context"
	"encoding/json"
	"prophet-trader/interfaces"
	"prophet-trader/services"
	"time"

	log "github.com/sirupsen/logrus"
)

// Ticker periodically polls existing services and broadcasts JSON snapshots
// to connected WebSocket clients via the Hub.
type Ticker struct {
	hub             *Hub
	tradingService  interfaces.TradingService
	activityLogger  *services.ActivityLogger
	interval        time.Duration
	startTime       time.Time
}

// NewTicker creates a new Ticker that will poll at the given interval.
func NewTicker(hub *Hub, tradingService interfaces.TradingService, activityLogger *services.ActivityLogger, interval time.Duration) *Ticker {
	return &Ticker{
		hub:            hub,
		tradingService: tradingService,
		activityLogger: activityLogger,
		interval:       interval,
		startTime:      time.Now(),
	}
}

// Run starts the ticker loop. It is intended to be run as a goroutine.
// Panic recovery ensures a dashboard failure does not crash the trading bot (WS-05 isolation).
func (t *Ticker) Run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Ticker panic recovered: %v", r)
		}
	}()

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	// Publish an initial snapshot immediately.
	t.publishSnapshot(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info("Dashboard ticker stopped")
			return
		case <-ticker.C:
			t.publishSnapshot(ctx)
		}
	}
}

// publishSnapshot gathers data from services and broadcasts a JSON snapshot.
// Errors are logged but do not crash -- partial snapshots with nil fields are sent.
func (t *Ticker) publishSnapshot(ctx context.Context) {
	snapshot := DashboardSnapshot{
		Type:      "snapshot",
		Timestamp: time.Now(),
	}

	// Fetch account data
	botAlive := false
	if account, err := t.tradingService.GetAccount(ctx); err != nil {
		log.Warnf("Dashboard ticker: failed to get account: %v", err)
	} else {
		botAlive = true
		snapshot.Account = &AccountData{
			Cash:        account.Cash,
			Equity:      account.PortfolioValue,
			BuyingPower: account.BuyingPower,
		}
	}

	// Fetch positions
	if positions, err := t.tradingService.GetPositions(ctx); err != nil {
		log.Warnf("Dashboard ticker: failed to get positions: %v", err)
	} else {
		posData := make([]*PositionData, len(positions))
		for i, p := range positions {
			posData[i] = &PositionData{
				Symbol:          p.Symbol,
				Qty:             p.Qty,
				EntryPrice:      p.AvgEntryPrice,
				CurrentPrice:    p.CurrentPrice,
				UnrealizedPL:    p.UnrealizedPL,
				UnrealizedPLPct: p.UnrealizedPLPC,
				MarketValue:     p.MarketValue,
				Side:            p.Side,
			}
		}
		snapshot.Positions = posData
	}

	// Fetch recent activity
	if actLog, err := t.activityLogger.GetCurrentLog(); err != nil {
		// No active session is normal outside trading hours -- log at debug level.
		log.Debugf("Dashboard ticker: no activity log: %v", err)
	} else {
		activities := actLog.Activities
		// Take last 50 items, reversed for reverse-chronological order.
		maxItems := 50
		if len(activities) < maxItems {
			maxItems = len(activities)
		}
		items := make([]ActivityItem, maxItems)
		for i := 0; i < maxItems; i++ {
			src := activities[len(activities)-1-i]
			items[i] = ActivityItem{
				Timestamp: src.Timestamp,
				Type:      src.Type,
				Action:    src.Action,
				Symbol:    src.Symbol,
				Details:   src.Details,
			}
		}
		snapshot.Activity = items
	}

	// Build bot health data
	snapshot.BotHealth = &BotHealthData{
		Alive:        botAlive,
		LastActivity: time.Now(),
		Uptime:       time.Since(t.startTime).Round(time.Second).String(),
	}

	// Marshal and broadcast
	data, err := json.Marshal(snapshot)
	if err != nil {
		log.Errorf("Dashboard ticker: failed to marshal snapshot: %v", err)
		return
	}

	t.hub.broadcast <- data
}
