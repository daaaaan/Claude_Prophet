# Requirements: Prophet Trader Dashboard

**Defined:** 2026-02-11
**Core Value:** At a glance, know exactly what's happening with your money — with the ability to intervene when needed.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### WebSocket Infrastructure

- [ ] **WS-01**: WebSocket endpoint at `/ws` pushes JSON snapshots to connected browser clients
- [ ] **WS-02**: WebSocket Hub manages client connections with proper goroutine lifecycle (ping/pong, deadlines, cleanup)
- [ ] **WS-03**: Dashboard Ticker polls existing services at configurable intervals and publishes events to Hub
- [ ] **WS-04**: SQLite database runs in WAL mode to prevent lock contention between bot writes and dashboard reads
- [ ] **WS-05**: Dashboard code isolated in `dashboard/` package — panic in WebSocket code cannot crash the trading bot

### Portfolio Overview

- [ ] **PORT-01**: User can see account balance, equity, and buying power updated in real-time
- [ ] **PORT-02**: User can see all open positions with symbol, quantity, entry price, current price, and unrealized P&L color-coded green/red
- [ ] **PORT-03**: User can see total aggregate unrealized P&L across all positions
- [ ] **PORT-04**: User can see cash-to-deployed-capital ratio (trading rules mandate 50-70% cash)
- [ ] **PORT-05**: User can see day trade count and PDT compliance status
- [ ] **PORT-06**: User can see each position's distance to stop-loss and take-profit targets as visual bars
- [ ] **PORT-07**: User can see sector exposure breakdown as a chart showing percentage per sector

### Real-Time Data

- [ ] **RT-01**: Position and account data updates via WebSocket push without manual refresh
- [ ] **RT-02**: Price updates stream for symbols in open positions
- [ ] **RT-03**: Connection status indicator shows live/stale/disconnected state with color degradation (green < 5s, yellow 5-30s, red > 30s)
- [ ] **RT-04**: Bot health indicator shows whether backend is alive and timestamp of last activity

### Activity Feed

- [ ] **ACT-01**: User can see a real-time feed of recent orders and fills in reverse chronological order
- [ ] **ACT-02**: User can see position state changes (PENDING → ACTIVE → STOPPED_OUT etc.) in the feed
- [ ] **ACT-03**: New activity items appear via WebSocket push without refresh

### Emergency Controls

- [ ] **EMR-01**: User can close an individual position with a button click (via REST, not WebSocket)
- [ ] **EMR-02**: User can cancel a pending order with a button click (via REST, not WebSocket)
- [ ] **EMR-03**: User can close ALL positions with a panic button that requires confirmation before executing
- [ ] **EMR-04**: Emergency controls remain functional even when WebSocket connection is degraded

### Intelligence & AI

- [ ] **INT-01**: User can see AI decision feed showing Gemini analysis reasoning from decisive_actions logs
- [ ] **INT-02**: User can see trading rules compliance indicators (green/yellow/red) auto-checked against trading rules
- [ ] **INT-03**: User can see market intelligence panel with AI-curated news from existing endpoints
- [ ] **INT-04**: AI decisions appear as contextual items in the activity feed

### History & Analytics

- [ ] **HIST-01**: User can see a sortable, filterable trade history table with entry/exit prices, P&L, and duration
- [ ] **HIST-02**: User can see daily P&L summary view
- [ ] **HIST-03**: User can see performance analytics: win rate, profit factor, average hold time
- [ ] **HIST-04**: User can see historical equity curve from account snapshots

### Frontend Shell

- [ ] **UI-01**: Dashboard served from Go backend at `/dashboard` route — no separate frontend server
- [ ] **UI-02**: Frontend built with Alpine.js for reactivity, Tailwind CSS for styling, zero build step
- [ ] **UI-03**: Static assets embedded in Go binary for single-binary deployment
- [ ] **UI-04**: Dashboard layout with multiple panels visible simultaneously (command center style)

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Enhanced Features

- **V2-01**: Mobile-responsive view for quick portfolio check on phone
- **V2-02**: Configurable/rearrangeable panel layout
- **V2-03**: Browser push notifications for significant events
- **V2-04**: Options chain visualization
- **V2-05**: Paper vs live mode visual indicator with safeguards

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Full charting with drawing tools | User already has TradingView — dashboard complements, doesn't replace |
| Strategy editor / backtesting | Different product entirely — dashboard is monitoring/control layer |
| Multi-user authentication | Single personal tool on private network |
| Order entry forms | Bot handles trading — dashboard is for monitoring and emergency intervention |
| Mobile-first responsive design | Desktop command center — mobile compromises panel density |
| Copy trading / social features | Single user, zero value |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| WS-01 | — | Pending |
| WS-02 | — | Pending |
| WS-03 | — | Pending |
| WS-04 | — | Pending |
| WS-05 | — | Pending |
| PORT-01 | — | Pending |
| PORT-02 | — | Pending |
| PORT-03 | — | Pending |
| PORT-04 | — | Pending |
| PORT-05 | — | Pending |
| PORT-06 | — | Pending |
| PORT-07 | — | Pending |
| RT-01 | — | Pending |
| RT-02 | — | Pending |
| RT-03 | — | Pending |
| RT-04 | — | Pending |
| ACT-01 | — | Pending |
| ACT-02 | — | Pending |
| ACT-03 | — | Pending |
| EMR-01 | — | Pending |
| EMR-02 | — | Pending |
| EMR-03 | — | Pending |
| EMR-04 | — | Pending |
| INT-01 | — | Pending |
| INT-02 | — | Pending |
| INT-03 | — | Pending |
| INT-04 | — | Pending |
| HIST-01 | — | Pending |
| HIST-02 | — | Pending |
| HIST-03 | — | Pending |
| HIST-04 | — | Pending |
| UI-01 | — | Pending |
| UI-02 | — | Pending |
| UI-03 | — | Pending |
| UI-04 | — | Pending |

**Coverage:**
- v1 requirements: 35 total
- Mapped to phases: 0
- Unmapped: 35 ⚠️

---
*Requirements defined: 2026-02-11*
*Last updated: 2026-02-11 after initial definition*
