# Feature Research

**Domain:** Real-time trading bot dashboard (personal single-user command center)
**Researched:** 2026-02-11
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = dashboard feels broken or useless.

#### Portfolio Overview Panel

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Account balance / equity / buying power | First thing any trader looks at. Every broker dashboard shows this. | LOW | Already available via `GET /api/v1/account` -- just display it |
| Open positions list with live P&L | Core purpose of a trading dashboard. Cryptohopper, Fidelity, every platform shows this front-and-center. Positions highlighted green/red by profit/loss. | MEDIUM | `GET /api/v1/orders/positions` returns positions. Need WebSocket for live P&L updates derived from position entry price vs streaming market price. |
| Total unrealized P&L (aggregate) | Trader needs single-glance portfolio health. Sum of all position P&L. | LOW | Computed client-side from position data |
| Day trade count / PDT status | Project context: user is a Pattern Day Trader with $100K+ equity. Must track remaining day trades and PDT compliance. | LOW | Already in Account struct (`DayTradeCount`, `PatternDayTrader` fields) |
| Cash / deployed capital ratio | TRADING_RULES.md mandates "50-70% cash at all times." Dashboard must make this visible. | LOW | Computed from account cash vs portfolio value |

#### Real-Time Data Streaming

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| WebSocket push for position updates | PROJECT.md explicitly requires "WebSocket for live updates, not polling." Every modern trading dashboard streams. Even 1-second polling feels sluggish vs push. | HIGH | `StreamBars()` is currently a stub. Must implement Alpaca WebSocket streaming on backend, then relay to browser via Gin WebSocket upgrade. Two-hop architecture: Alpaca WS -> Go backend -> Browser WS. |
| Price updates for watched symbols | Trader monitors SPY, QQQ, NVDA, AMD, TSLA, MSTR, etc. Must see live prices without manual refresh. | MEDIUM | Piggybacks on same WebSocket infrastructure. Backend subscribes to Alpaca streams for symbols in positions + watchlist, fans out to browser. |
| Connection status indicator | User must know if live data is flowing or stale. Disconnected WebSocket with no indicator = silent data staleness, which is dangerous for a trader. | LOW | Simple heartbeat/reconnection status on frontend. Small but critical UX element. |

#### Activity Feed

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Recent orders and fills | Trader needs to see what the bot did. Cryptohopper shows "Activity" section with last actions and next scheduled checks. | MEDIUM | `GET /api/v1/orders?status=all` returns order history. Display as reverse-chronological feed. Push new fills via WebSocket. |
| Position state changes | When a managed position transitions (PENDING -> ACTIVE -> STOPPED_OUT), the dashboard must surface this immediately. | MEDIUM | Position manager already tracks state. Need to emit events on state change, relay via WebSocket. |
| Bot status / health | Is the bot running? When was last activity? Cryptohopper prominently shows "bot enabled/disabled" and "last activity" timestamp. | LOW | Health endpoint + last-seen timestamp. Simple but reassuring. |

#### Emergency Controls

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Close individual position | PROJECT.md lists "Emergency controls (close position, cancel order)" as active requirement. Every trading platform has a close button per position. | LOW | Already available via `DELETE /api/v1/positions/managed/:id`. Just wire a button. |
| Cancel pending order | Same as above. Must be able to cancel any open order. | LOW | Already available via `DELETE /api/v1/orders/:id`. Wire a button. |
| Close ALL positions (panic button) | Cryptohopper's panic button is their most visible emergency control. When things go wrong fast, trader needs one-click liquidation. Research shows this is standard for bot dashboards. | MEDIUM | Need a new backend endpoint that iterates all open positions and submits market sell orders. Must have confirmation step to prevent accidental activation (research recommends "time delays or multiple-step authentication"). |

#### Historical Data Views

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Trade history table | PROJECT.md wants "Historical drill-down." Every trading journal (TradesViz, Tradervue, WealthBee) centers on a sortable, filterable trade log. | MEDIUM | `DBTrade` model stores completed trades with entry/exit prices, P&L, duration. Need a paginated table with sorting and date-range filtering. |
| Daily P&L summary | TradesViz and Jigsaw Trading both use calendar-based P&L views. Trader needs to see "how did today go?" and "how did this week go?" at a glance. | MEDIUM | `DBAccountSnapshot` stores periodic portfolio values. Compute daily delta. Display as calendar heatmap or simple day-by-day list. |

### Differentiators (Competitive Advantage)

Features that set this dashboard apart from generic trading dashboards. Not required, but high-value for a personal AI-driven trading bot.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| AI decision feed | No competitor dashboard shows WHY the bot made decisions. Prophet Trader has Gemini-powered analysis and decisive_actions logs. Surfacing AI reasoning ("Bought NVDA calls because technical confluence + bullish news sentiment") transforms the dashboard from a monitor into a learning tool. | MEDIUM | `decisive_actions/` JSON files + `DBSignal` model already capture this. Parse and display as a feed alongside activity. User explicitly wants AI decisions visible. |
| Trading rules compliance monitor | TRADING_RULES.md defines 20+ rules (max 15% per position, max 10 positions, max 5 scalps/day, etc.). No competitor auto-checks rule compliance. Dashboard can show green/yellow/red indicators for each rule. | MEDIUM | Computed from current positions + account data + today's trade count. Rules are well-defined and quantifiable. Unique to this system. |
| Market intelligence panel | Built-in news aggregation from Google News + MarketWatch, cleaned by Gemini AI. Most dashboards link OUT to news. This one has AI-curated, trading-relevant intelligence inline. | MEDIUM | `GET /api/v1/intelligence/quick-market` and `/api/v1/intelligence/cleaned-news` already exist. Just render the output. |
| Position risk visualization | Show each position's distance to stop-loss, take-profit targets, and trailing stop levels. Managed positions already track all of this. Visual risk map is rare in personal dashboards. | MEDIUM | `DBManagedPosition` has `StopLossPrice`, `TakeProfitPrice`, `TrailingPercent`. Render as visual bars or gauge per position showing current price relative to stop/target. |
| Performance analytics | Win rate, average winner vs average loser, profit factor, best/worst trades, P&L by symbol, P&L by strategy. TradesViz tracks 600+ stats. We do not need 600, but core analytics (win rate, profit factor, avg hold time) would be genuinely valuable and not available in any off-the-shelf bot dashboard. | HIGH | Requires querying `DBTrade` history and computing aggregates. More of a reporting/analytics feature than real-time. Good candidate for a later phase. |
| Sector exposure breakdown | TRADING_RULES.md mandates max 40% in any single sector. Dashboard could show a pie chart of exposure by sector (Tech, Crypto, Broad Market). Visual compliance check. | LOW | Map symbols to sectors (hardcoded or config-driven), compute percentages from positions. Simple but unique. |

### Anti-Features (Deliberately NOT Building)

Features that seem good but create problems for this specific project.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Full charting with drawing tools | TradingView-style charting seems essential. Traders love charts. | Massive complexity (TradingView took years to build). User already HAS TradingView for charting. Duplicating it is waste. Dashboard should complement, not replace, dedicated charting tools. | Show simple sparkline price charts per position. Link to TradingView for deep analysis. |
| Options chain visualization | PROJECT.md explicitly lists this as "Out of Scope - future enhancement." Options chains are complex UI (strike matrix, Greeks columns, expiration tabs). | High complexity, incomplete backend (OCC symbol parsing is a known tech debt item). Options chain viewer is a product unto itself. | Show options positions with their Greeks inline. Defer chain browsing to broker platform. |
| Strategy editor / backtesting | Common in algo trading platforms (QuantConnect, Zipline). Seems like natural extension. | Completely different product. Dashboard is a monitoring/control layer, not a development environment. Backtesting requires historical data infrastructure, strategy DSL, simulation engine. | Keep strategy logic in the Go backend and MCP server. Dashboard shows results, not code. |
| Multi-user / auth system | "What if someone else needs access?" | PROJECT.md and constraints explicitly state single-user, private network. Auth adds complexity with zero value. Adding auth means session management, login UI, token refresh, CORS headaches. | No auth. If sharing is ever needed, put it behind a VPN or SSH tunnel. |
| Mobile responsive design | "I want to check on my phone." | Optimizing for mobile is a separate design effort. A trading command center with 6-8 panels does not compress well to mobile. Trying to be responsive early means compromising desktop density. | Build desktop-first. Mobile can be a future milestone if needed. A simple "portfolio value + P&L" mobile view could be added later without redesigning the whole dashboard. |
| Notification / alerting system | Price alerts, position alerts, news alerts. Every platform has them. | PROJECT.md lists this as "Out of Scope - future enhancement." Alerts require a notification delivery system (email, SMS, push, browser notifications), preference management, rate limiting. Significant infrastructure for a personal tool. | The dashboard itself IS the alert -- it is the always-open command center. The user checks it 2-3x per day per their trading rules. If alerts are needed later, start with simple browser notifications (low complexity). |
| Copy trading / social features | Popular in retail platforms (eToro, Cryptohopper marketplace). | Single user. No one to copy. Zero value. | N/A |
| Order entry forms | Full order entry with all parameters (limit price, stop price, time-in-force, etc.). | The bot handles order entry. Dashboard is for monitoring and emergency intervention, not primary trading. Complex order forms duplicate what Alpaca's own UI and the MCP tools already do. | Emergency controls only: close position, cancel order, panic button. If the user needs to place a nuanced order, use the MCP tools or Alpaca directly. |

## Feature Dependencies

```
[WebSocket Infrastructure]
    |--enables--> [Live Position P&L Updates]
    |--enables--> [Live Price Updates for Watchlist]
    |--enables--> [Real-time Activity Feed]
    |--enables--> [Connection Status Indicator]

[Account/Position REST Endpoints (existing)]
    |--enables--> [Portfolio Overview Panel]
    |--enables--> [Emergency Controls]
    |--enables--> [Trading Rules Compliance Monitor]
    |--enables--> [Sector Exposure Breakdown]

[Portfolio Overview Panel]
    |--enhances--> [Position Risk Visualization]
                       |--requires--> [Managed Position Data]

[Trade History REST Endpoint (existing)]
    |--enables--> [Trade History Table]
    |--enables--> [Daily P&L Summary]
    |--enables--> [Performance Analytics]

[Activity Logging (existing)]
    |--enables--> [AI Decision Feed]
    |--enables--> [Activity Feed]

[News/Intelligence Endpoints (existing)]
    |--enables--> [Market Intelligence Panel]

[Emergency Controls]
    |--requires--> [Portfolio Overview Panel] (need to see what you are closing)
    |--requires--> [Confirmation Dialog] (prevent accidental panic)

[Performance Analytics]
    |--requires--> [Trade History Table] (build table first, analytics on top)
```

### Dependency Notes

- **WebSocket Infrastructure is the critical path**: Four table-stakes features depend on it. Without WebSocket, the dashboard is just a REST API viewer with manual refresh -- which fails the core requirement of "real-time."
- **Portfolio Overview Panel is the foundation**: Emergency controls and compliance monitoring both read from position/account data. Build the data layer first, controls second.
- **AI Decision Feed requires no new backend work**: `decisive_actions/` files and `DBSignal` already exist. This is purely a frontend parsing and display task, making it a quick win differentiator.
- **Performance Analytics is isolated**: It depends only on trade history data and can be built independently as a later phase without blocking anything else.

## MVP Definition

### Launch With (v1)

Minimum viable dashboard -- what is needed to replace "check Alpaca UI + read log files manually."

- [ ] **Portfolio overview panel** -- account balance, equity, buying power, cash ratio, PDT status. This is the single most important screen.
- [ ] **Open positions list with P&L** -- even without real-time streaming, polling every 5-10 seconds is acceptable for v1. Positions with unrealized P&L, color-coded.
- [ ] **Emergency controls** -- close position button, cancel order button. The reason a dashboard exists beyond just reading an API.
- [ ] **Recent activity feed** -- last N orders/fills, displayed in reverse chronological order. Polling-based is fine for v1.
- [ ] **Bot health indicator** -- is the backend alive, when was last activity.

### Add After Validation (v1.x)

Features to add once the core dashboard is working and the user is actually using it daily.

- [ ] **WebSocket streaming** -- upgrade from polling to push. Add when the polling refresh feels too slow (it will). This is the biggest single improvement.
- [ ] **Panic button (close all positions)** -- add when the user has experienced a moment of wanting it. Requires new backend endpoint + confirmation UX.
- [ ] **AI decision feed** -- surface decisive_actions and signals. Add when user wants to understand bot behavior without reading JSON files.
- [ ] **Trading rules compliance indicators** -- add when user wants automated rule-checking instead of mental math.
- [ ] **Market intelligence panel** -- render cleaned news and analysis. Add when user wants news in the dashboard instead of asking the MCP tools.
- [ ] **Sector exposure breakdown** -- simple pie chart. Add when compliance monitoring is in place.
- [ ] **Position risk visualization** -- stop/target distance bars. Add when managed positions are the primary workflow.

### Future Consideration (v2+)

Features to defer until the dashboard is a daily driver and the user wants more depth.

- [ ] **Performance analytics** -- win rate, profit factor, P&L curves. Requires meaningful trade history to be useful. Defer until enough trades are logged.
- [ ] **Daily P&L calendar** -- calendar heatmap view. Nice visualization but not actionable until weeks of data exist.
- [ ] **Trade history deep-dive** -- sortable, filterable trade log with drill-down to individual trade details. Useful for weekly review sessions.
- [ ] **Historical equity curve** -- portfolio value over time from account snapshots. Requires consistent snapshot collection over weeks/months.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Portfolio overview (balance, equity, positions) | HIGH | LOW | P1 |
| Open positions with P&L | HIGH | MEDIUM | P1 |
| Close position button | HIGH | LOW | P1 |
| Cancel order button | HIGH | LOW | P1 |
| Recent activity feed | HIGH | LOW | P1 |
| Bot health indicator | MEDIUM | LOW | P1 |
| Connection status indicator | MEDIUM | LOW | P1 |
| WebSocket streaming infrastructure | HIGH | HIGH | P2 |
| Live P&L updates via WebSocket | HIGH | MEDIUM | P2 |
| Panic button (close all) | HIGH | MEDIUM | P2 |
| AI decision feed | HIGH | LOW | P2 |
| Trading rules compliance | MEDIUM | MEDIUM | P2 |
| Market intelligence panel | MEDIUM | LOW | P2 |
| Sector exposure chart | MEDIUM | LOW | P2 |
| Position risk visualization | MEDIUM | MEDIUM | P2 |
| Trade history table | MEDIUM | MEDIUM | P3 |
| Daily P&L summary | MEDIUM | MEDIUM | P3 |
| Performance analytics | MEDIUM | HIGH | P3 |
| Historical equity curve | LOW | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch -- the dashboard is useless without these
- P2: Should have, add when core is stable -- these make the dashboard genuinely better than checking Alpaca directly
- P3: Nice to have, future consideration -- analytics and history that compound in value over time

## Competitor Feature Analysis

| Feature | Cryptohopper | Fidelity Trader+ | TradesViz | TradingView | Prophet Dashboard (Our Approach) |
|---------|--------------|------------------|-----------|-------------|----------------------------------|
| Live positions with P&L | Yes, green/red highlighting | Yes | Post-trade only | No (charting focus) | Yes, real-time via WebSocket |
| Panic button / kill switch | Yes, prominent | No | No | No | Yes, with confirmation step |
| Bot activity log | Yes, "Output" section | N/A | N/A | N/A | Yes, plus AI reasoning |
| AI decision transparency | No | No | No | No | **Yes -- unique differentiator** |
| Trading rules compliance | No | No | No | No | **Yes -- unique differentiator** |
| Performance analytics | Basic (profit, total assets) | Advanced | 600+ metrics | N/A | Core metrics (win rate, profit factor) |
| News integration | No | Yes (external) | No | Yes (community) | Yes, AI-curated and inline |
| Charting | No | Advanced | Basic | Best-in-class | Minimal (sparklines). Link to TradingView. |
| Options Greeks display | No | Yes | Yes | Basic | Inline per position, no chain browser |
| Customizable layout | Widget-based | Multi-panel | Dashboard tabs | Fully customizable | Fixed layout initially. Simplicity over customization for single user. |

## Sources

- [TailAdmin: Stock Market Dashboard Templates for 2026](https://tailadmin.com/blog/stock-market-dashboard-templates) -- dashboard template features and expectations
- [Pocket Option: Trading Dashboard Essential Tools](https://pocketoption.com/blog/en/interesting/trading-platforms/trading-dashboard/) -- effective dashboard components and best practices
- [OptionStranglers: Creating a Custom Options Trading Dashboard](https://optionstranglers.com.sg/blogs/news/creating-a-custom-options-trading-dashboard-build-the-ultimate-workspace-for-financial-freedom) -- 8 core panel types for options dashboards
- [Cryptohopper: Dashboard Documentation](https://docs.cryptohopper.com/docs/trading-bot/what-is-on-the-dashboard-for-the-trading-bot/) -- bot dashboard widgets, controls, and panic button
- [Cryptohopper: Panic Button](https://support.cryptohopper.com/en/articles/9010202-what-does-the-panic-button-do) -- emergency liquidation feature design
- [NYIF: Trading System Kill Switch](https://www.nyif.com/articles/trading-system-kill-switch-panacea-or-pandoras-box/) -- kill switch design considerations and regulatory perspective
- [TradesViz: Trading Journal](https://www.tradesviz.com/trading-journal/) -- 600+ analytics metrics, calendar views, P&L visualization
- [Jigsaw Trading: Journal Analytics](https://www.jigsawtrading.com/trade-journal-analytics/) -- calendar-based performance views
- [TradingView: Features](https://www.tradingview.com/features/) -- charting and dashboard capabilities
- [Fidelity: Trading Dashboard](https://www.fidelity.com/trading/trading-dashboard) -- institutional dashboard feature set
- [SSE vs WebSockets for Stock Market Apps](https://www.marketcalls.in/python/sse-vs-websockets-choosing-the-right-tool-for-stock-market-applications.html) -- WebSocket latency and update frequency guidance

---
*Feature research for: Prophet Trader Dashboard*
*Researched: 2026-02-11*
