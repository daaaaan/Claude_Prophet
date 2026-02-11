# Prophet Trader Dashboard

## What This Is

A real-time web dashboard for the Prophet Trader system — a personal command center that surfaces portfolio state, market intelligence, trading activity, and historical performance from the existing Go backend and MCP server. Single-user tool for monitoring live trading with emergency manual override controls.

## Core Value

At a glance, know exactly what's happening with your money — open positions, P&L, recent activity, and market context — with the ability to intervene when needed.

## Requirements

### Validated

<!-- Existing capabilities from the Go backend and MCP server. -->

- ✓ REST API for trading operations (orders, positions, account) — existing
- ✓ Alpaca integration for order execution and market data — existing
- ✓ Managed position lifecycle with stop-loss/take-profit — existing
- ✓ Market news aggregation (Google News, MarketWatch RSS) — existing
- ✓ AI-powered news analysis via Gemini — existing
- ✓ Technical analysis indicators — existing
- ✓ MCP server for AI agent integration — existing
- ✓ Vector DB for trade history similarity search — existing
- ✓ Activity logging to JSON files — existing

### Active

<!-- Dashboard features to build. -->

- [ ] Real-time portfolio overview (balance, buying power, equity, open positions with live P&L)
- [ ] Live market data display for watched symbols
- [ ] Real-time activity feed (orders, fills, position changes)
- [ ] Market intelligence panel (news headlines, AI analysis summaries)
- [ ] Historical trade and performance views (drill into past activity)
- [ ] Emergency controls (close position, cancel order)
- [ ] WebSocket streaming from Go backend for real-time push
- [ ] Web frontend served from the existing Go backend

### Out of Scope

- Multi-user support / authentication — single personal tool
- Mobile app — web dashboard only
- Modifying AI agent logic or MCP server behavior — dashboard is read/control layer
- Backtesting or strategy editor — separate concern
- Options chain visualization — future enhancement
- Automated alerting/notifications — future enhancement

## Context

- Go backend already runs on port 4534 with Gin, exposing REST endpoints for all trading operations
- MCP server (Node.js) handles AI agent integration separately — dashboard reads from Go API, not MCP
- Activity logs stored as JSON files in `activity_logs/` directory
- AI decisive actions stored as JSON in `decisive_actions/` directory
- SQLite database at `./data/prophet_trader.db` holds orders, positions, bars, snapshots
- Alpaca API provides real-time market data and account state
- No existing web frontend — this is greenfield UI on a brownfield backend

## Constraints

- **Backend**: Must integrate with existing Go/Gin server — add routes, don't replace
- **No auth**: Single user on private network, no authentication layer needed
- **Data source**: Dashboard reads from Go REST API — does not talk to Alpaca directly
- **Real-time**: WebSocket for live updates, not polling
- **Deployment**: Runs locally alongside existing bot process

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Serve dashboard from Go backend | Single process, no separate frontend server, simpler deployment | — Pending |
| WebSocket for real-time updates | User wants live streaming, not refresh-based | — Pending |
| No authentication | Personal single-user tool on private network | — Pending |
| AI decisions as feed context | User wants them visible but not prominent | — Pending |

---
*Last updated: 2026-02-11 after initialization*
