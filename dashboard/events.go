package dashboard

import "time"

// DashboardSnapshot is the top-level JSON payload sent to WebSocket clients.
type DashboardSnapshot struct {
	Type             string                 `json:"type"`
	Timestamp        time.Time              `json:"timestamp"`
	Account          *AccountData           `json:"account,omitempty"`
	Positions        []*PositionData        `json:"positions,omitempty"`
	Activity         []ActivityItem         `json:"activity,omitempty"`
	BotHealth        *BotHealthData         `json:"bot_health,omitempty"`
	ManagedPositions []*ManagedPositionData `json:"managed_positions,omitempty"`
}

// ManagedPositionData represents a managed position with risk management state.
type ManagedPositionData struct {
	PositionID   string    `json:"position_id"`
	Symbol       string    `json:"symbol"`
	Status       string    `json:"status"`
	Side         string    `json:"side"`
	EntryPrice   float64   `json:"entry_price"`
	StopLoss     float64   `json:"stop_loss"`
	TakeProfit   float64   `json:"take_profit"`
	UnrealizedPL float64   `json:"unrealized_pl"`
	CreatedAt    time.Time `json:"created_at"`
}

// AccountData represents the trading account state.
type AccountData struct {
	Cash        float64 `json:"cash"`
	Equity      float64 `json:"equity"`
	BuyingPower float64 `json:"buying_power"`
}

// PositionData represents a single open position.
type PositionData struct {
	Symbol          string  `json:"symbol"`
	Qty             float64 `json:"qty"`
	EntryPrice      float64 `json:"entry_price"`
	CurrentPrice    float64 `json:"current_price"`
	UnrealizedPL    float64 `json:"unrealized_pl"`
	UnrealizedPLPct float64 `json:"unrealized_pl_pct"`
	MarketValue     float64 `json:"market_value"`
	Side            string  `json:"side"`
}

// ActivityItem represents a single activity entry for the dashboard feed.
type ActivityItem struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Symbol    string                 `json:"symbol,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Reasoning string                 `json:"reasoning,omitempty"`
}

// BotHealthData represents the bot's operational health status.
type BotHealthData struct {
	Alive        bool      `json:"alive"`
	LastActivity time.Time `json:"last_activity"`
	Uptime       string    `json:"uptime"`
}
