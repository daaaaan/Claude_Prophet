# Pitfalls Research

**Domain:** Real-time trading dashboard added to existing Go trading bot
**Researched:** 2026-02-11
**Confidence:** HIGH (based on codebase analysis + verified community patterns + official docs)

## Critical Pitfalls

### Pitfall 1: SQLite "database is locked" Under Concurrent WebSocket + Bot Access

**What goes wrong:**
The existing SQLite database uses GORM with default journal mode (rollback, not WAL). The current bot already has background goroutines writing position snapshots every 5 minutes and running data cleanup. Adding WebSocket handlers that read positions, account snapshots, and activity logs while the bot simultaneously writes creates concurrent read/write contention. SQLite in rollback mode blocks readers when a writer holds a lock, causing "database is locked" errors that propagate as failed dashboard updates or, worse, failed trade recordings.

**Why it happens:**
The current codebase (`database/storage.go`) opens SQLite with `gorm.Open(sqlite.Open(dbPath), ...)` and never sets `PRAGMA journal_mode=WAL` or configures connection pooling. GORM's default SQLite driver (`mattn/go-sqlite3`) creates a connection pool, but SQLite only supports one writer at a time. Without WAL mode, even readers block on writers. WebSocket handlers issuing frequent reads (every 1-5 seconds for live data) dramatically increase contention with the bot's write operations.

**How to avoid:**
1. Enable WAL mode immediately after opening the database: `db.Exec("PRAGMA journal_mode=WAL")`
2. Set `db.SetMaxOpenConns(1)` on the underlying `*sql.DB` to serialize writes through a single connection (prevents "database is locked" from concurrent write attempts)
3. Use `PRAGMA busy_timeout=5000` to have SQLite retry on lock contention rather than immediately failing
4. Keep read transactions short -- no long-lived transactions from WebSocket handlers

**Warning signs:**
- Intermittent "database is locked" errors in logs
- Dashboard showing stale data while bot logs show successful writes
- Errors correlating with position monitor ticker (every 5 minutes)

**Phase to address:**
Phase 1 (WebSocket infrastructure) -- this must be fixed before any WebSocket reads are added, as it affects the existing bot's reliability too.

---

### Pitfall 2: WebSocket Goroutine and Memory Leaks from Unmanaged Connections

**What goes wrong:**
Each WebSocket connection spawns at least 2 goroutines (read pump + write pump, per the Gorilla WebSocket chat pattern). If connections drop without proper cleanup -- browser tab closed, network interruption, laptop sleep -- goroutines and their associated memory (write buffers, channel buffers) leak. The Gorilla WebSocket library holds write buffers for the lifetime of the connection. In a trading dashboard that runs all day, leaked goroutines accumulate and eventually consume significant memory or cause the bot process to degrade.

**Why it happens:**
Developers implement the "happy path" (connect, send messages, client disconnects cleanly) but miss edge cases: browser navigating away without close frame, network drops, WebSocket ping/pong timeout not configured, or the Hub's unregister path having a race condition. Gorilla WebSocket's `ReadMessage` blocks forever if no read deadline is set, pinning the goroutine even after the remote side is gone.

**How to avoid:**
1. Implement the full Hub pattern from Gorilla's chat example: Hub goroutine + per-client read/write goroutines communicating via channels
2. Set read deadlines and use ping/pong: `conn.SetReadDeadline(time.Now().Add(pongWait))` with a pong handler that resets the deadline
3. Set write deadlines on every write: `conn.SetWriteDeadline(time.Now().Add(writeWait))`
4. Use `select` with default case on broadcast to skip slow clients rather than blocking the entire broadcast
5. Configure `WriteBufferPool` on the Upgrader to reduce per-connection memory overhead
6. On unregister, close the send channel AND the WebSocket connection
7. Use `context.WithCancel` tied to the main bot context so all WebSocket goroutines clean up on shutdown

**Warning signs:**
- Goroutine count (visible via `runtime.NumGoroutine()` or `/debug/pprof/goroutine`) steadily increasing over hours
- Memory usage growing without corresponding increase in connections
- Dashboard reconnects seeming to "work" but old goroutines still running

**Phase to address:**
Phase 1 (WebSocket infrastructure) -- get the connection lifecycle right from the start.

---

### Pitfall 3: Alpaca Single-Connection Limit Conflicts Between Bot and Dashboard Streaming

**What goes wrong:**
Alpaca enforces a **1 concurrent WebSocket connection limit per endpoint** for most subscription tiers. The existing bot's `StreamBars` function (currently a TODO stub in `services/alpaca_data.go`) will eventually use this connection. If the dashboard independently opens its own Alpaca streaming connection for live market data, the second connection gets error 406 "connection limit exceeded" and the first connection may be terminated. This means either the bot loses its data stream or the dashboard does.

**Why it happens:**
The natural instinct is to have the dashboard connect directly to Alpaca's WebSocket for real-time quotes. But Alpaca's free/basic tier only allows 1 streaming connection. Developers discover this only after both components are built, requiring a rearchitecture.

**How to avoid:**
1. Design a single Alpaca streaming connection managed by the Go backend as the sole consumer of Alpaca WebSocket data
2. The Go backend fans out received market data to dashboard clients via its own WebSocket hub
3. Never let the frontend connect directly to Alpaca's streaming endpoints
4. For market data the dashboard needs but the bot doesn't stream, use REST polling from the Go backend (respecting the 200 req/min rate limit)
5. Consider the Alpaca Go SDK's streaming client as the canonical connection, with the dashboard as a downstream consumer

**Warning signs:**
- "Connection limit exceeded" errors from Alpaca
- One of bot/dashboard randomly losing its data stream
- Alpaca auth errors appearing intermittently (the displaced connection)

**Phase to address:**
Phase 2 (real-time data integration) -- must be architecturally decided before implementing any Alpaca streaming.

---

### Pitfall 4: Emergency Controls That Don't Work When You Need Them Most

**What goes wrong:**
The dashboard's emergency controls (kill switch, close-all-positions, cancel-all-orders) are routed through the same WebSocket connection or HTTP path as everything else. During a market crisis -- when you most need emergency controls -- the system is under peak load: market data flooding in, multiple position updates, the bot potentially executing rapid trades. The emergency action gets queued behind other messages, delayed by a saturated WebSocket write buffer, or fails because the dashboard's WebSocket connection dropped under load.

**Why it happens:**
Emergency controls are treated as "just another feature" and share the same communication channel and error handling as routine dashboard updates. No priority system exists. If the WebSocket connection is dead (which happens more under stress), the user has no fallback.

**How to avoid:**
1. Emergency controls must use dedicated REST endpoints, not WebSocket messages -- REST is stateless and doesn't depend on an existing connection being alive
2. The emergency REST endpoints should bypass any middleware that could slow them down (rate limiting, heavy logging, data enrichment)
3. Implement a server-side emergency mode: once triggered, the bot halts all new trades regardless of dashboard connectivity
4. Add a physical confirmation step (double-click, hold button) to prevent accidental triggers, but DO NOT add delays to the actual execution
5. The existing `/api/v1/orders` DELETE endpoint and position close endpoints should be the foundation -- wrap them, don't rebuild them
6. Test emergency controls under load -- simulate high-frequency updates while triggering emergency actions

**Warning signs:**
- Emergency button clicks with no immediate visual feedback
- Emergency actions arriving at the server with >1 second latency
- No fallback when WebSocket disconnects during emergencies
- Emergency endpoints sharing error handling with routine data endpoints

**Phase to address:**
Phase 3 (emergency controls) -- but the architectural decision (REST not WS) must be made in Phase 1.

---

### Pitfall 5: Dashboard Showing Stale Prices That Cause Wrong Trading Decisions

**What goes wrong:**
The dashboard displays a price that was accurate 30 seconds ago, the user makes a trade decision based on that price, and the fill is significantly different. This is especially dangerous with the existing system because the MCP server and bot make trades programmatically -- if the dashboard shows stale data, the human operator may not intervene when they should (or may intervene when they shouldn't).

**Why it happens:**
Multiple staleness sources compound: (a) Alpaca REST polling has inherent delay vs. streaming, (b) the Go backend may cache data for efficiency, (c) the WebSocket broadcast to the dashboard adds latency, (d) the frontend may not re-render immediately, (e) there is no visual indicator distinguishing "live" data from "last known" data. During market hours with fast-moving stocks, even 5-10 seconds of staleness matters.

**How to avoid:**
1. Every data point sent to the dashboard must include a timestamp from the data source (not the server's current time)
2. The frontend must display data age: show "Last update: 2s ago" and visually degrade (yellow > orange > red) as data ages
3. Define staleness thresholds: <5s = fresh (green), 5-30s = stale (yellow), >30s = very stale (red), >60s = show "DISCONNECTED"
4. Never cache market data for more than a few seconds in the Go backend
5. On WebSocket reconnection, immediately request a full state refresh rather than waiting for the next push
6. Disable or visually gate trade-related actions when data is stale beyond a threshold

**Warning signs:**
- No timestamps visible on price displays
- Prices that don't change during active market hours
- Users making decisions without noticing the dashboard is disconnected
- No visual difference between "live streaming" and "last polled 30 seconds ago"

**Phase to address:**
Phase 2 (real-time data) for the data pipeline, Phase 4 (UI polish) for the visual staleness indicators.

---

### Pitfall 6: Coupling Dashboard State to Bot Process Lifecycle

**What goes wrong:**
The existing bot is a single `main.go` process that runs the Gin HTTP server, background position monitors, activity logging, and data cleanup -- all in one process. Adding WebSocket infrastructure, a streaming hub, market data fan-out, and dashboard state management to this same process means: (a) any panic in a WebSocket handler can crash the entire trading bot, (b) restarting the bot to fix a dashboard bug interrupts active trading, (c) the bot's graceful shutdown (currently `time.Sleep(2*time.Second)`) may not be enough time to close all WebSocket connections and flush pending data.

**Why it happens:**
It is the natural path of least resistance to add the dashboard directly into the existing process. The existing `setupRouter` function in `main.go` already has `router.Static("/dashboard", "./web")`, suggesting this was always the plan. For a single-user local tool, a separate process feels like over-engineering.

**How to avoid:**
1. Accept single-process for now (it is correct for a single-user local tool) BUT isolate via careful goroutine management
2. Use `recover()` in all WebSocket handler goroutines so a panic in a dashboard connection does not crash the bot
3. Extend the graceful shutdown to properly close the WebSocket hub, drain client connections, and wait for all WebSocket goroutines to exit
4. Keep WebSocket code in its own package (e.g., `ws/` or `dashboard/`) with clean interfaces to the existing services -- do not scatter WebSocket logic across existing controllers
5. Do not modify existing controller signatures to accommodate WebSocket needs; instead, create a dashboard-specific service layer that calls the existing controllers/services
6. Consider: if the dashboard needs a restart, can you do it without restarting the trading bot? With the single-process model, the answer is no. Document this as an accepted tradeoff.

**Warning signs:**
- WebSocket code importing directly from `controllers/` and modifying existing handlers
- No `recover()` in WebSocket goroutines
- Bot crashes with panic stack traces originating from WebSocket code
- Unable to restart dashboard functionality without interrupting active trading

**Phase to address:**
Phase 1 (WebSocket infrastructure) -- establish the package boundary and isolation patterns from the start.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Polling REST instead of WebSocket streaming for market data | Simpler to implement, no connection management | Higher Alpaca API usage (200 req/min limit), higher latency, more server load | Phase 1 MVP only, must migrate to streaming in Phase 2 |
| Sending full state on every WebSocket message instead of diffs | Simpler client code, no state synchronization bugs | Bandwidth waste, higher CPU for JSON serialization, slow on mobile/weak connections | Acceptable for single-user local deployment, revisit if ever multi-user |
| Storing dashboard preferences in localStorage only | No backend changes needed | Lost on browser clear, no sync across devices | Always acceptable for single-user tool |
| No authentication on WebSocket endpoint | Faster development, simpler code | Anyone on the network can view portfolio and trigger trades | Only acceptable when bound to localhost. If ever exposed to network, CRITICAL security issue |
| Embedding frontend assets with `go:embed` from the start | Single binary deployment | Requires Go rebuild for any frontend change, painful during frontend development | Use filesystem serving during development, switch to `go:embed` for production builds |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Alpaca Market Data WS | Opening a second streaming connection from the dashboard (triggers 406 error) | Single connection in Go backend, fan out to dashboard clients via internal WebSocket hub |
| Alpaca Trade Updates WS | Not handling binary frames from paper trading endpoint (docs confirm paper uses binary, not text frames) | Check frame type and decode accordingly; test with both paper and live endpoints |
| Alpaca REST API | Polling too aggressively for dashboard freshness, hitting 200 req/min rate limit | Batch requests (multi-symbol endpoints), cache with short TTL, prefer streaming for frequently-changing data |
| Alpaca Auth | Hardcoded 10-second auth timeout on WS connect -- slow startup or network can fail auth | Implement retry with backoff on auth failure; don't treat first auth failure as fatal |
| SQLite via GORM | Using GORM's default connection pool settings with SQLite (SQLite doesn't support true connection pooling) | Set `SetMaxOpenConns(1)` and enable WAL mode; keep queries fast and transactions short |
| Frontend charting library | Loading full price history into memory for chart rendering | Use data decimation/downsampling for long time ranges; only load full resolution for visible viewport |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Broadcasting full portfolio state every second via WebSocket | UI feels responsive initially | Send diffs or use event-driven updates (only push when data changes) | With 20+ positions and multiple data fields, JSON payloads become large; serialization CPU adds up |
| Querying Alpaca REST API per-symbol in a loop for multi-position dashboard | Works with 2-3 positions | Use batch/multi-symbol endpoints (`GetLatestBars([]string{...})` already in codebase) | At 10+ positions, sequential REST calls exceed latency budget and approach rate limit |
| Storing every WebSocket message in SQLite for history | Audit trail works | Store aggregated snapshots on intervals, not every tick; use in-memory ring buffer for recent data | At 1 msg/sec for 8 hours = 28,800 rows/day per data type; SQLite write throughput becomes bottleneck |
| Canvas-based chart re-rendering entire dataset on every update | Smooth with 100 data points | Append-only rendering; only redraw the new candle/point, not the full history | At 10,000+ data points, full re-render causes visible frame drops |
| No backpressure on WebSocket write queue | Fast clients get instant updates | Use buffered channel with `select`/`default` to drop messages for slow clients; configurable buffer size | When browser tab is backgrounded, it stops processing WS messages; buffer fills, goroutine blocks, hub broadcast stalls for ALL clients |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| WebSocket endpoint bound to 0.0.0.0 instead of 127.0.0.1 | Anyone on the local network can view portfolio, trigger trades, access emergency controls | Bind to localhost only; the existing Gin server should verify this in config. Add `AllowedOrigins` check on WebSocket upgrade |
| No CORS restriction on WebSocket upgrade (current CORS is `Allow-Origin: *`) | Any website open in the browser can connect to the local WebSocket and read/control the trading bot via cross-origin requests | Set `CheckOrigin` on the WebSocket upgrader to only allow requests from the dashboard's origin (localhost:4534) |
| API keys visible in WebSocket messages or frontend code | Alpaca credentials exposed in browser dev tools | Never send API keys to the frontend; all Alpaca communication happens server-side only |
| Emergency kill switch with no confirmation | Accidental click liquidates entire portfolio | Require double-action (click + confirm) but DO NOT add artificial delays to the execution itself |
| No rate limiting on trade-execution endpoints from dashboard | A bug or accidental rapid clicking could submit dozens of orders | Server-side rate limit on order submission endpoints (e.g., max 1 order per symbol per 2 seconds from dashboard) |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Showing raw P&L numbers without context | User sees "-$150" but doesn't know if that's -0.1% or -15% of the position | Always show both absolute and percentage P&L; color-code (green/red) consistently |
| No visual distinction between paper and live trading | User thinks they're paper trading but are live (or vice versa) | Prominent, persistent banner: "PAPER TRADING" in yellow or "LIVE TRADING" in red; different background accent color |
| Chart time axis without market hours context | User sees flat lines during market close and thinks data is broken | Mark pre-market, market hours, and after-hours on charts; optionally collapse non-trading hours |
| Activity feed showing every internal bot event | Feed becomes noise; user can't find important events | Categorize events (trade, alert, system, debug); default to showing trades + alerts only; allow filtering |
| Price updates that flash too rapidly to read | User can't process information; causes anxiety | Throttle UI updates to max 1-2 per second even if data arrives faster; use smooth transitions, not jarring number jumps |
| Emergency controls buried in a menu | Under stress, user can't find the kill switch | Emergency controls always visible, fixed position, distinct color (red), never scrolled off screen |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **WebSocket connection:** Often missing reconnection with exponential backoff on the client side -- verify the dashboard reconnects after network blips and server restarts
- [ ] **Real-time prices:** Often missing staleness indicators -- verify the UI shows data age and degrades visually when data is old
- [ ] **Portfolio display:** Often missing handling of zero positions -- verify the dashboard shows a meaningful empty state, not a blank table or error
- [ ] **Emergency controls:** Often missing server-side persistence of emergency state -- verify that killing the dashboard browser doesn't cancel an in-progress emergency liquidation
- [ ] **Activity feed:** Often missing pagination/virtualization -- verify the feed doesn't load 10,000 entries into the DOM and freeze the browser
- [ ] **Chart rendering:** Often missing cleanup on component unmount -- verify switching between chart views doesn't leak chart instances and memory
- [ ] **WebSocket hub:** Often missing graceful handling of hub shutdown -- verify that stopping the bot properly closes all client connections with a close frame
- [ ] **Paper/Live indicator:** Often missing on every page -- verify the indicator is visible from ANY dashboard view, not just the home page
- [ ] **Market data:** Often missing market-closed handling -- verify the dashboard shows "Market Closed" status and doesn't keep polling during off-hours
- [ ] **Error handling:** Often missing user-facing error messages -- verify that backend errors show a toast/notification, not a silent failure

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| SQLite "database is locked" errors | LOW | Enable WAL mode (single PRAGMA statement); set busy timeout; no data loss, just add the pragma |
| Goroutine leak from WebSocket | MEDIUM | Add `/debug/pprof/goroutine` endpoint; identify leaked goroutines; add proper cleanup; requires process restart to reclaim leaked goroutines |
| Alpaca connection limit exceeded | LOW | Kill the duplicate connection; restructure to single-connection fan-out pattern; no data loss |
| Emergency controls unresponsive | HIGH | Fall back to direct Alpaca API calls via curl/CLI; document emergency Alpaca API commands in a runbook; keep Alpaca dashboard URL bookmarked as last resort |
| Stale data displayed to user | MEDIUM | Add timestamp to all data payloads; implement staleness detection in frontend; retroactively audit any trades made during stale-data period |
| Bot crash from WebSocket panic | MEDIUM | Add `recover()` wrappers; use process supervisor (systemd) for auto-restart; review `autonomous_trading.sh` for restart logic; some in-flight data may be lost |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| SQLite database locked | Phase 1: WebSocket infrastructure | Run concurrent read/write test; check PRAGMA journal_mode returns "wal" |
| WebSocket goroutine leaks | Phase 1: WebSocket infrastructure | Monitor `runtime.NumGoroutine()` over 1 hour of connect/disconnect cycles; count should be stable |
| Alpaca single-connection limit | Phase 2: Real-time data integration | Verify only 1 Alpaca WS connection exists while dashboard is active; check no 406 errors in logs |
| Emergency controls reliability | Phase 3: Emergency controls | Test emergency actions while WebSocket is disconnected; verify server-side execution completes independently |
| Stale data display | Phase 2: Data pipeline + Phase 4: UI polish | Disconnect network briefly; verify dashboard shows staleness indicator within 5 seconds |
| Bot/dashboard coupling | Phase 1: Package structure | Verify `ws/` package only imports interfaces, not concrete controller types; verify `recover()` in all WS goroutines |
| CORS/security on WebSocket | Phase 1: WebSocket infrastructure | Attempt WebSocket connection from a different origin; verify rejection |
| Chart memory leaks | Phase 4: Chart implementation | Open/close chart views 50 times; verify browser memory (via DevTools) is stable |
| Activity feed performance | Phase 4: Activity feed view | Load 10,000 activity entries; verify scroll performance stays smooth (virtualized list) |

## Sources

- [Gorilla WebSocket documentation and issues](https://pkg.go.dev/github.com/gorilla/websocket) -- HIGH confidence (official docs)
- [Gorilla WebSocket chat example (Hub pattern)](https://github.com/gorilla/websocket/blob/main/examples/chat/README.md) -- HIGH confidence
- [Gorilla WebSocket memory leak issues #141, #236, #296, #462, #995](https://github.com/gorilla/websocket/issues/236) -- HIGH confidence (official issue tracker)
- [GORM SQLite "database is locked" issue #3709](https://github.com/go-gorm/gorm/issues/3709) -- HIGH confidence (official issue tracker)
- [SQLite WAL mode documentation](https://sqlite.org/wal.html) -- HIGH confidence (official docs)
- [Alpaca WebSocket streaming docs](https://docs.alpaca.markets/docs/streaming-market-data) -- HIGH confidence (official docs, verified via WebFetch)
- [Alpaca trade updates streaming docs](https://docs.alpaca.markets/docs/websocket-streaming) -- HIGH confidence (official docs, verified binary frame gotcha)
- [Alpaca API rate limits](https://alpaca.markets/support/usage-limit-api-calls) -- HIGH confidence (official support page)
- [Alpaca connection limit exceeded community reports](https://forum.alpaca.markets/t/t-error-code-406-msg-connection-limit-exceeded/10896) -- MEDIUM confidence (community forum, consistent with official docs)
- [Go embed for SPA assets](https://blog.jetbrains.com/go/2021/06/09/how-to-use-go-embed-in-go-1-16/) -- MEDIUM confidence (reputable source)
- [Chart.js performance / decimation docs](https://www.chartjs.org/docs/latest/general/performance.html) -- HIGH confidence (official docs)
- Codebase analysis of `database/storage.go`, `cmd/bot/main.go`, `services/alpaca_data.go`, `config/config.go` -- HIGH confidence (direct source examination)

---
*Pitfalls research for: Prophet Trader real-time dashboard milestone*
*Researched: 2026-02-11*
