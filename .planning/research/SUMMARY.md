# Project Research Summary

**Project:** Prophet Trader Real-Time Web Dashboard
**Domain:** Real-time trading bot dashboard (single-user command center added to existing Go/Gin backend)
**Researched:** 2026-02-11
**Confidence:** HIGH

## Executive Summary

Prophet Trader is an existing Go-based algorithmic trading bot with a full REST API, SQLite persistence, Alpaca brokerage integration, and AI-powered analysis via Gemini. The dashboard milestone adds a real-time web interface for monitoring positions, viewing bot activity, and executing emergency controls. Experts build this type of single-user dashboard using server-rendered HTML with lightweight client-side reactivity -- not a SPA framework. The recommended stack is Go templates + Alpine.js for UI, gorilla/websocket for real-time push, and TradingView Lightweight Charts for financial charting. No Node.js build step. No React/Vue/Svelte. The entire dashboard embeds into the existing Go binary.

The architecture follows a well-established pattern: a WebSocket Hub goroutine broadcasts events to connected browser clients, fed by a Dashboard Ticker that polls existing services at configurable intervals (5-30 seconds depending on data type). The browser receives JSON snapshots over WebSocket and Alpine.js reactively renders the UI. User actions (emergency controls, order cancellation) flow through existing REST endpoints, not WebSocket -- this is a deliberate architectural decision that keeps emergency controls reliable even when the WebSocket connection is degraded. The existing codebase requires zero modifications to controllers, services, or models; all new code lives in a dedicated `dashboard/` package and `web/` static directory.

The top risks are: SQLite lock contention when WebSocket reads overlap with bot writes (mitigated by enabling WAL mode), WebSocket goroutine leaks from improper connection cleanup (mitigated by implementing the full Gorilla hub pattern with deadlines and ping/pong), and Alpaca's single-connection streaming limit conflicting between bot and dashboard (mitigated by designing a single backend connection with fan-out). Emergency control reliability under stress is an architectural concern that must be addressed from Phase 1 by routing all actions through REST, not WebSocket. Data staleness is a UX risk that requires timestamps on every data point and visual degradation indicators in the frontend.

## Key Findings

### Recommended Stack

The stack is zero-build-step, Go-native, and optimized for single-user dashboard deployment. All frontend dependencies are vendored as static files embedded in the Go binary. No npm, no webpack, no separate dev server.

**Core technologies:**
- **Go `html/template` + `embed`:** Server-side templating and single-binary asset embedding -- already in stdlib, no build tools needed
- **Alpine.js 3.15.x:** Client-side reactivity (15KB) for data binding, toggles, modals, filtering -- replaces React/Vue at 1% of the complexity
- **gorilla/websocket 1.5.3:** WebSocket server with the most documentation and examples for the hub/broadcast pattern this dashboard needs
- **TradingView Lightweight Charts 5.1.0:** The standard for web financial charts (35KB gzipped), with `series.update()` for real-time streaming -- no viable alternative
- **htmx 2.0.7:** Dynamic HTML updates via attributes for panel switching and navigation -- complements Alpine.js
- **Tailwind CSS 4.1.x:** Utility-first CSS via standalone CLI -- fast iteration on dashboard components without writing custom CSS
- **Alpaca Go SDK v3.9.1:** Upgrade from current v3.5.0 for latest WebSocket streaming fixes

**Key stack decision:** Alpine.js over htmx for the primary data rendering pattern. While htmx excels at server-rendered HTML fragment swapping, the dashboard's primary data flow is WebSocket-pushed JSON rendered client-side. Alpine.js handles this natively with reactive data binding. htmx is still used for navigation and panel switching where server-rendered HTML is appropriate.

### Expected Features

**Must have (table stakes -- dashboard is useless without these):**
- Portfolio overview panel (balance, equity, buying power, cash ratio, PDT status)
- Open positions list with unrealized P&L (color-coded green/red)
- Emergency controls: close individual position, cancel pending order
- Recent activity feed (orders, fills, state changes)
- Bot health indicator (alive/dead, last activity timestamp)
- Connection status indicator (live/stale/disconnected)

**Should have (add after core is stable -- makes dashboard better than checking Alpaca directly):**
- WebSocket streaming for live push updates (upgrade from polling)
- Panic button: close ALL positions with confirmation step
- AI decision feed (surface Gemini analysis reasoning -- unique differentiator no competitor has)
- Trading rules compliance monitor (automated checking of 20+ rules from TRADING_RULES.md -- unique differentiator)
- Market intelligence panel (AI-curated news already available via existing endpoints)
- Position risk visualization (stop-loss/take-profit distance per position)
- Sector exposure breakdown (pie chart for max-40%-per-sector rule compliance)

**Defer (v2+):**
- Performance analytics (win rate, profit factor, P&L curves) -- needs trade history volume
- Daily P&L calendar heatmap -- needs weeks of data
- Trade history deep-dive with filtering -- weekly review feature
- Historical equity curve -- needs months of snapshots

**Anti-features (deliberately NOT building):**
- Full charting with drawing tools (user already has TradingView)
- Options chain visualization (out of scope per PROJECT.md)
- Strategy editor / backtesting (different product entirely)
- Multi-user auth (single-user, private network)
- Mobile responsive design (desktop command center; mobile compromises density)
- Order entry forms (bot handles trading; dashboard is for monitoring and emergency intervention)

### Architecture Approach

The architecture adds 3 new components to the existing Go process: a WebSocket Hub, an Event Bus, and a Dashboard Ticker. All new code lives in a `dashboard/` package with clean interfaces to existing services. The frontend is plain HTML/JS/CSS served from a `web/` directory. Data flows one direction for streaming (Alpaca/DB -> Ticker -> Hub -> Browser) and uses existing REST endpoints for actions (Browser -> REST Controller -> Alpaca). Full state snapshots are sent on every WebSocket push (no delta merging) because data volumes are trivially small for a single user.

**Major components:**
1. **WebSocket Hub** (`dashboard/hub.go`) -- Single goroutine managing client connections, receiving events from the bus, broadcasting JSON to all connected clients by topic
2. **Dashboard Ticker** (`dashboard/ticker.go`) -- Background goroutine polling existing services at configured intervals (account: 30s, positions: 10s, activity: 5s, news: 60s) and publishing state change events
3. **Event Bus** (`dashboard/events.go`) -- Typed event structs with topic/type/payload/timestamp envelope, in-process Go channels connecting Ticker to Hub
4. **WebSocket Handler** (`dashboard/handler.go`) -- Gin route handler for `/ws` upgrade, per-client read/write pump goroutines with proper lifecycle management
5. **Frontend Shell** (`web/`) -- Alpine.js reactive data store receiving WebSocket JSON, rendering panels, calling REST for actions

### Critical Pitfalls

1. **SQLite "database is locked"** -- The existing DB uses rollback journal mode, not WAL. Adding frequent WebSocket-driven reads alongside bot writes will cause lock contention. Fix: enable WAL mode (`PRAGMA journal_mode=WAL`), set `MaxOpenConns(1)`, add `busy_timeout=5000`. Must be done in Phase 1 before any WebSocket reads.

2. **WebSocket goroutine and memory leaks** -- Each connection spawns 2+ goroutines. Without read/write deadlines, ping/pong handlers, and proper unregister cleanup, goroutines leak on disconnection. Fix: implement full Gorilla hub pattern with deadlines, use `select`/`default` on broadcast to skip slow clients, wrap all goroutines in `recover()`.

3. **Alpaca single-connection streaming limit** -- Alpaca allows only 1 WebSocket connection per endpoint per tier. Bot and dashboard cannot both connect independently. Fix: single Alpaca connection in Go backend with fan-out to dashboard clients. Never let the frontend connect directly to Alpaca streaming.

4. **Emergency controls failing under stress** -- If emergency actions share the WebSocket channel, they can be delayed or lost during market crises. Fix: route ALL emergency controls through dedicated REST endpoints. REST is stateless and does not depend on WebSocket connection health.

5. **Stale data causing wrong trading decisions** -- Polling-based data has inherent latency. Without staleness indicators, the user cannot distinguish live from stale data. Fix: include source timestamps on every data point, display data age in the UI with color-coded degradation (green < 5s, yellow 5-30s, red > 30s, "DISCONNECTED" > 60s).

6. **Dashboard crash taking down the trading bot** -- Dashboard code shares the Go process with the trading bot. A panic in WebSocket code crashes everything. Fix: `recover()` in all WebSocket goroutines, keep dashboard code in isolated package, extend graceful shutdown to properly drain WebSocket connections.

## Implications for Roadmap

Based on combined research, the dashboard should be built in 5 phases following the dependency chain identified in the architecture research. WebSocket infrastructure is the critical path -- four table-stakes features depend on it.

### Phase 1: WebSocket Infrastructure and Foundation

**Rationale:** The architecture research identifies WebSocket infrastructure as having zero UI dependencies -- it can be tested with `wscat` before any frontend exists. Four of seven critical pitfalls must be addressed in this phase (SQLite WAL, goroutine lifecycle, process isolation, CORS security). Building this first de-risks the entire project.
**Delivers:** Working WebSocket connection at `/ws` that pushes JSON snapshots of positions, account, and activity data. SQLite hardened with WAL mode. Dashboard package with clean service boundaries.
**Addresses:** WebSocket push (table stakes), connection status foundation, bot health endpoint
**Avoids:** SQLite lock contention (Pitfall 1), goroutine leaks (Pitfall 2), bot/dashboard coupling (Pitfall 6), CORS vulnerability
**Stack elements:** gorilla/websocket 1.5.3, Go channels (event bus), existing Gin router
**Architecture components:** Hub, Client, Events, Handler, Ticker -- the entire `dashboard/` package

### Phase 2: Frontend Shell and Core Panels

**Rationale:** With WebSocket pushing data, the frontend can now render it. The portfolio overview panel and emergency controls are the two most operationally critical features -- the portfolio panel is the foundation all other panels build on, and emergency controls are the safety-critical reason a dashboard exists beyond just reading an API.
**Delivers:** Working dashboard at `/dashboard` with portfolio overview (balance, equity, positions with P&L), emergency controls (close position, cancel order), activity feed, and bot health indicator. Full Alpine.js reactive rendering from WebSocket data.
**Addresses:** Portfolio overview (P1), open positions with P&L (P1), close position button (P1), cancel order button (P1), activity feed (P1), bot health indicator (P1), connection status indicator (P1)
**Avoids:** Emergency control reliability (Pitfall 4) by using REST for all actions from day one
**Stack elements:** Alpine.js 3.15.x, htmx 2.0.7, Tailwind CSS 4.1.x, Go html/template
**Architecture components:** Frontend shell (`web/`), Alpine.js data store, WebSocket client with reconnect

### Phase 3: Real-Time Streaming and Live Data

**Rationale:** Phase 2 delivers a functional dashboard with polling-based updates. This phase upgrades to true real-time push and adds the panic button. The Alpaca single-connection pitfall must be resolved here. This is also where data staleness indicators become essential.
**Delivers:** Live P&L updates via WebSocket push, price streaming for watched symbols, panic button (close all positions) with confirmation UX, staleness indicators on all data.
**Addresses:** WebSocket streaming (P2), live P&L updates (P2), panic button (P2), staleness indicators
**Avoids:** Alpaca single-connection conflict (Pitfall 3), stale data display (Pitfall 5)
**Stack elements:** Alpaca Go SDK v3.9.1 streaming, TradingView Lightweight Charts 5.1.0

### Phase 4: Differentiator Panels

**Rationale:** With the core dashboard stable and streaming live data, add the features that make this dashboard genuinely better than checking Alpaca directly. The AI decision feed and trading rules compliance monitor are unique differentiators no competitor offers, and both require zero new backend work (data already exists).
**Delivers:** AI decision feed (Gemini reasoning visible), trading rules compliance monitor (20+ rules auto-checked), market intelligence panel (AI-curated news), sector exposure breakdown, position risk visualization.
**Addresses:** AI decision feed (P2), trading rules compliance (P2), market intelligence (P2), sector exposure (P2), position risk visualization (P2)
**Avoids:** Activity feed performance (Pitfall -- virtualization needed for large feeds), chart memory leaks (Pitfall -- proper cleanup on panel switch)
**Stack elements:** TradingView Lightweight Charts for risk visualization, Alpine.js components

### Phase 5: Analytics and History

**Rationale:** Analytics features compound in value over time -- they need trade history volume to be meaningful. Deferring to the final phase ensures the dashboard has been collecting data for weeks/months before analytics are built. This phase is isolated from all others with no downstream dependencies.
**Delivers:** Trade history table with sorting/filtering, daily P&L summary, performance analytics (win rate, profit factor, average hold time), historical equity curve.
**Addresses:** Trade history (P3), daily P&L summary (P3), performance analytics (P3), historical equity curve (P3)
**Stack elements:** TradingView Lightweight Charts for equity curves, Alpine.js data tables

### Phase Ordering Rationale

- **Infrastructure before UI:** WebSocket hub can be tested without a frontend (via `wscat` or browser console). Building and verifying the data pipeline first means frontend development has real data to render from day one.
- **Safety before features:** Emergency controls in Phase 2 (not deferred) because the dashboard is operationally useless -- and potentially dangerous -- if you can see positions but cannot act on them.
- **Polling before streaming:** Phase 2 works with the Ticker's polling intervals. Phase 3 upgrades to Alpaca streaming. This is deliberate technical debt: polling is simpler, and the visual difference between 10-second polling and sub-second streaming is negligible for a portfolio dashboard. It lets Phase 2 ship faster.
- **Differentiators after stability:** AI decision feed and rules compliance are unique and high-value but not operationally critical. Adding them after the core is stable ensures they do not block the user from getting a working dashboard.
- **Analytics last:** Requires accumulated data. Building it first means showing empty charts for weeks.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 1:** WebSocket hub pattern needs careful implementation -- review Gorilla chat example thoroughly. SQLite WAL mode interaction with GORM's connection management needs testing.
- **Phase 3:** Alpaca streaming SDK integration is the most uncertain area. The `StreamBars()` function is currently a stub. Paper trading uses binary WebSocket frames (not text). Connection limit behavior needs live verification.

Phases with standard patterns (skip research-phase):
- **Phase 2:** Standard Alpine.js + Tailwind dashboard panels. Extremely well-documented patterns. Existing REST API endpoints provide all needed data.
- **Phase 4:** Rendering existing data in new panels. No new integrations, no new backend work. Pure frontend feature development.
- **Phase 5:** Standard data table and chart patterns. Well-documented TradingView Lightweight Charts API for equity curves.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All technologies are mature, well-documented, and version-compatible. gorilla/websocket (24.5K stars), Alpine.js, Lightweight Charts are production-proven. No experimental or bleeding-edge choices. |
| Features | HIGH | Feature landscape derived from competitor analysis (Cryptohopper, Fidelity, TradesViz, TradingView) and direct codebase analysis. Existing API endpoints cover 90%+ of data needs. Clear MVP definition with honest anti-features list. |
| Architecture | HIGH | Hub/broadcast pattern is the canonical Go WebSocket architecture (Gorilla's own example). Ticker pattern is standard polling-to-push bridge. Existing codebase already has the static file serving route stubbed. No novel architecture. |
| Pitfalls | HIGH | All pitfalls verified against official documentation (SQLite WAL docs, Gorilla issues tracker, Alpaca streaming docs). SQLite contention and goroutine leaks are the two most commonly reported issues in Go WebSocket projects. Alpaca connection limit confirmed via official docs and community forum reports. |

**Overall confidence:** HIGH

### Gaps to Address

- **Alpaca streaming SDK behavior:** The `StreamBars()` stub has never been implemented. Actual behavior of the Alpaca Go SDK's `marketdata/stream` package under reconnection, auth failure, and binary frame conditions needs live testing during Phase 3. Research identified the risks but could not verify the SDK's resilience firsthand.
- **Tailwind CSS standalone CLI with Go templates:** Tailwind v4's content scanning needs to be configured to find class names inside `.html` Go template files. This is documented but untested for this specific project. May need a `tailwind.config.js` or `@source` directive adjustment.
- **go:embed vs filesystem serving during development:** Research recommends filesystem serving during development and `go:embed` for production. The exact toggle mechanism (build tags vs. config flag) needs to be decided during Phase 1 implementation.
- **Paper vs. live trading indicator:** Pitfalls research flags this as a critical UX element, but the mechanism for detecting paper vs. live mode from the Alpaca SDK configuration needs verification. The config has `APCA_API_BASE_URL` which differs between paper and live, but surfacing this to the frontend cleanly needs design.
- **Options position display:** The codebase has known tech debt around OCC symbol parsing for options. Displaying options positions with Greeks inline (as recommended in features research) may be blocked by this unresolved parsing issue.

## Sources

### Primary (HIGH confidence)
- [gorilla/websocket](https://github.com/gorilla/websocket) -- WebSocket server implementation, hub pattern, v1.5.3
- [gorilla/websocket chat example](https://github.com/gorilla/websocket/blob/main/examples/chat/README.md) -- canonical hub/broadcast pattern
- [SQLite WAL mode documentation](https://sqlite.org/wal.html) -- concurrent read/write solution
- [Alpaca WebSocket streaming docs](https://docs.alpaca.markets/docs/streaming-market-data) -- market data streaming, connection limits
- [Alpaca trade updates streaming](https://docs.alpaca.markets/docs/websocket-streaming) -- binary frame gotcha for paper trading
- [TradingView Lightweight Charts](https://tradingview.github.io/lightweight-charts/) -- v5.1.0 API, real-time updates
- [Alpine.js](https://alpinejs.dev/) -- v3.15.8 reactive framework
- [htmx.org](https://htmx.org/) -- v2.0.7, WebSocket extension
- [Tailwind CSS v4](https://tailwindcss.com/blog/tailwindcss-v4) -- standalone CLI, zero-config
- [Alpaca Go SDK](https://pkg.go.dev/github.com/alpacahq/alpaca-trade-api-go/v3) -- v3.9.1
- [GORM SQLite issues](https://github.com/go-gorm/gorm/issues/3709) -- "database is locked" solutions
- Direct codebase analysis of `cmd/bot/main.go`, `database/storage.go`, `services/alpaca_data.go`, `config/config.go`

### Secondary (MEDIUM confidence)
- [Ably WebSocket architecture best practices](https://ably.com/topic/websocket-architecture-best-practices) -- architecture patterns
- [Cryptohopper dashboard docs](https://docs.cryptohopper.com/docs/trading-bot/what-is-on-the-dashboard-for-the-trading-bot/) -- competitor feature analysis
- [TradesViz trading journal](https://www.tradesviz.com/trading-journal/) -- analytics feature benchmarking
- [NYIF Trading System Kill Switch](https://www.nyif.com/articles/trading-system-kill-switch-panacea-or-pandoras-box/) -- emergency control design
- [Alpaca connection limit forum reports](https://forum.alpaca.markets/t/t-error-code-406-msg-connection-limit-exceeded/10896) -- connection limit behavior
- [htmx + Alpine.js comparison](https://www.infoworld.com/article/3856520/htmx-and-alpine-js-how-to-combine-two-great-lean-front-ends.html) -- framework selection rationale

### Tertiary (LOW confidence)
- [TailAdmin stock market dashboard templates](https://tailadmin.com/blog/stock-market-dashboard-templates) -- general dashboard UI expectations
- [Pocket Option trading dashboard guide](https://pocketoption.com/blog/en/interesting/trading-platforms/trading-dashboard/) -- dashboard component best practices

---
*Research completed: 2026-02-11*
*Ready for roadmap: yes*
