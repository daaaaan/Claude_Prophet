# Roadmap: Prophet Trader Dashboard

## Overview

Build a real-time web dashboard on the existing Go/Gin backend that surfaces portfolio state, market intelligence, and trading activity with emergency intervention controls. The build progresses from data pipeline and core UI, through safety controls and live streaming, to AI-powered intelligence panels, and finally historical analytics. Each phase delivers a complete, verifiable capability layer.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation & Core Dashboard** - WebSocket pipeline, frontend shell, portfolio overview, activity feed, and real-time status indicators
- [ ] **Phase 2: Emergency Controls & Live Streaming** - Safety intervention controls and upgrade from polling to true real-time price streaming
- [ ] **Phase 3: Intelligence Panels** - AI decision feed, trading rules compliance, market intelligence, and position risk visualization
- [ ] **Phase 4: History & Analytics** - Trade history table, daily P&L summaries, performance metrics, and equity curve

## Phase Details

### Phase 1: Foundation & Core Dashboard
**Goal**: User opens `/dashboard` in a browser and sees live portfolio state -- account balances, open positions with P&L, and recent activity -- all updating without manual refresh
**Depends on**: Nothing (first phase)
**Requirements**: WS-01, WS-02, WS-03, WS-04, WS-05, PORT-01, PORT-02, PORT-03, RT-01, RT-03, RT-04, ACT-01, ACT-02, ACT-03, UI-01, UI-02, UI-03, UI-04
**Success Criteria** (what must be TRUE):
  1. User navigates to `/dashboard` and sees a multi-panel command center layout with portfolio, activity, and status panels rendered from WebSocket-pushed data
  2. User can see account balance, equity, and buying power updating in real-time without refreshing the page
  3. User can see all open positions listed with symbol, quantity, entry price, current price, and unrealized P&L color-coded green/red, plus aggregate P&L
  4. User can see a reverse-chronological feed of recent orders, fills, and position state changes that updates live via WebSocket
  5. User can see connection status (green/yellow/red based on data age) and bot health indicator (alive/dead with last activity timestamp)
**Plans:** 3 plans

Plans:
- [ ] 01-01-PLAN.md -- WebSocket infrastructure (Hub, Client, Ticker, Events, Handler, WAL mode, dashboard package isolation)
- [ ] 01-02-PLAN.md -- Frontend shell and portfolio panels (Alpine.js + Tailwind, templates, portfolio overview, positions list)
- [ ] 01-03-PLAN.md -- Activity feed and status indicators (activity feed panel, connection status, bot health, visual verification)

### Phase 2: Emergency Controls & Live Streaming
**Goal**: User can intervene in emergencies (close positions, cancel orders, panic-close everything) and sees true real-time price updates streaming for open positions
**Depends on**: Phase 1
**Requirements**: EMR-01, EMR-02, EMR-03, EMR-04, RT-02, PORT-04, PORT-05, PORT-06
**Success Criteria** (what must be TRUE):
  1. User can close an individual position or cancel a pending order with a single button click, and these controls work even when WebSocket connection is degraded
  2. User can close ALL positions via a panic button that requires explicit confirmation before executing
  3. User can see live price updates streaming for symbols in open positions (not just polling-interval snapshots)
  4. User can see cash-to-deployed ratio, day trade count with PDT compliance status, and stop-loss/take-profit distance bars per position
**Plans**: TBD

Plans:
- [ ] 02-01: Emergency controls (close position, cancel order, panic button with confirmation, REST-based resilience)
- [ ] 02-02: Live price streaming and portfolio detail panels (Alpaca streaming fan-out, cash ratio, PDT status, SL/TP distance bars)

### Phase 3: Intelligence Panels
**Goal**: User sees AI-powered insights -- Gemini analysis reasoning, automated trading rules compliance checks, curated market news, and sector exposure -- that make this dashboard better than checking Alpaca directly
**Depends on**: Phase 2
**Requirements**: INT-01, INT-02, INT-03, INT-04, PORT-07
**Success Criteria** (what must be TRUE):
  1. User can see AI decision feed showing Gemini analysis reasoning from decisive_actions logs, with AI decisions also appearing as contextual items in the activity feed
  2. User can see trading rules compliance indicators (green/yellow/red) automatically checked against the trading rules
  3. User can see a market intelligence panel with AI-curated news from existing endpoints
  4. User can see sector exposure breakdown as a chart showing percentage allocation per sector
**Plans**: TBD

Plans:
- [ ] 03-01: AI decision feed and trading rules compliance (decisive_actions rendering, rules engine, compliance indicators)
- [ ] 03-02: Market intelligence and sector exposure (news panel, sector breakdown chart)

### Phase 4: History & Analytics
**Goal**: User can review historical trading performance -- drill into past trades, see daily summaries, track win rate and profit factor, and visualize the equity curve over time
**Depends on**: Phase 3
**Requirements**: HIST-01, HIST-02, HIST-03, HIST-04
**Success Criteria** (what must be TRUE):
  1. User can see a sortable, filterable trade history table with entry/exit prices, P&L, and duration for completed trades
  2. User can see daily P&L summary view aggregating performance by day
  3. User can see performance analytics including win rate, profit factor, and average hold time
  4. User can see a historical equity curve charted from account snapshots
**Plans**: TBD

Plans:
- [ ] 04-01: Trade history and daily P&L (sortable/filterable table, daily summaries, REST endpoints for historical data)
- [ ] 04-02: Performance analytics and equity curve (win rate, profit factor, hold time metrics, TradingView Lightweight Charts equity curve)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 --> 2 --> 3 --> 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Core Dashboard | 0/3 | Planned | - |
| 2. Emergency Controls & Live Streaming | 0/2 | Not started | - |
| 3. Intelligence Panels | 0/2 | Not started | - |
| 4. History & Analytics | 0/2 | Not started | - |
