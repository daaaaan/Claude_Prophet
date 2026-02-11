# Architecture Research

**Domain:** Real-time trading dashboard added to existing Go/Gin backend
**Researched:** 2026-02-11
**Confidence:** HIGH

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Browser (Single User)                        │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────────┐    │
│  │ Portfolio  │  │ Activity  │  │  Market   │  │   Emergency   │    │
│  │  Panel    │  │   Feed    │  │   Data    │  │   Controls    │    │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └──────┬────────┘    │
│        │              │              │               │              │
│        └──────────────┴──────┬───────┴───────────────┘              │
│                              │                                      │
│                    ┌─────────┴─────────┐                            │
│                    │  WebSocket Client │  (+ REST for actions)      │
│                    └─────────┬─────────┘                            │
├──────────────────────────────┼──────────────────────────────────────┤
│                     Network  │  :4534                               │
├──────────────────────────────┼──────────────────────────────────────┤
│                        Go/Gin Process                               │
│                              │                                      │
│  ┌───────────────────────────┴────────────────────────────────┐     │
│  │                    Gin Router                               │     │
│  │  /api/v1/*  (existing REST)   /ws  (new WebSocket)         │     │
│  │  /dashboard/* (new static)                                  │     │
│  └────────┬──────────────────────────────┬────────────────────┘     │
│           │                              │                          │
│  ┌────────┴────────┐          ┌──────────┴──────────┐               │
│  │  REST Controllers│          │   WebSocket Hub     │               │
│  │  (existing)     │          │   (new)             │               │
│  │  - Orders       │          │   - Client registry │               │
│  │  - Positions    │          │   - Event broadcast │               │
│  │  - Account      │          │   - Topic channels  │               │
│  │  - News         │          └──────────┬──────────┘               │
│  │  - Intelligence │                     │                          │
│  └────────┬────────┘          ┌──────────┴──────────┐               │
│           │                   │   Event Bus         │               │
│           │                   │   (new, in-process) │               │
│           └───────┬───────────┴──────────┬──────────┘               │
│                   │                      │                          │
│  ┌────────────────┴──────────────────────┴───────────────────┐      │
│  │                     Services Layer (existing)              │      │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐   │      │
│  │  │ Trading      │ │ Data         │ │ Position         │   │      │
│  │  │ Service      │ │ Service      │ │ Manager          │   │      │
│  │  └──────────────┘ └──────────────┘ └──────────────────┘   │      │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐   │      │
│  │  │ News         │ │ Gemini AI    │ │ Activity         │   │      │
│  │  │ Service      │ │ Service      │ │ Logger           │   │      │
│  │  └──────────────┘ └──────────────┘ └──────────────────┘   │      │
│  └───────────────────────────┬───────────────────────────────┘      │
│                              │                                      │
│  ┌───────────────────────────┴───────────────────────────────┐      │
│  │                     Data Layer                             │      │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐   │      │
│  │  │ SQLite/GORM  │ │ JSON Files   │ │ Alpaca API       │   │      │
│  │  │ (local DB)   │ │ (activity)   │ │ (external)       │   │      │
│  │  └──────────────┘ └──────────────┘ └──────────────────┘   │      │
│  └───────────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Gin Router** (existing) | Route HTTP requests, serve static files, upgrade WebSocket connections | `gin.Default()` with route groups, `router.Static("/dashboard", ...)` |
| **REST Controllers** (existing) | Handle API requests for trading, data, news, intelligence | `controllers/*.go` -- unchanged, dashboard consumes these |
| **WebSocket Hub** (new) | Manage connected clients, broadcast events by topic, handle reconnection | Single goroutine hub with register/unregister channels, topic-based filtering |
| **Event Bus** (new) | Decouple services from WebSocket delivery; services publish events, hub consumes them | Go channels with typed event structs, in-process pub/sub |
| **Dashboard Ticker** (new) | Periodically poll services and publish state change events to the event bus | Goroutine with configurable intervals per data type |
| **Static File Server** (existing stub) | Serve the HTML/JS/CSS dashboard from `./web` directory | `router.Static("/dashboard", "./web")` already wired in `main.go` |
| **Services Layer** (existing) | Business logic for trading, market data, news, AI analysis | `services/*.go` -- unchanged, event bus wraps calls to these |
| **Storage Layer** (existing) | SQLite persistence via GORM, JSON activity logs | `database/storage.go` -- unchanged |

## Recommended Project Structure

```
prophet-trader/
├── cmd/bot/main.go              # Entry point (modify to add hub + ticker)
├── config/config.go             # Config (add dashboard settings)
├── controllers/                 # Existing REST controllers (unchanged)
│   ├── order_controller.go
│   ├── position_controller.go
│   ├── activity_controller.go
│   ├── news_controller.go
│   └── intelligence_controller.go
├── dashboard/                   # NEW: all dashboard backend code
│   ├── hub.go                   # WebSocket hub (client registry, broadcast)
│   ├── client.go                # WebSocket client (read/write pumps)
│   ├── events.go                # Event types and bus
│   ├── ticker.go                # Periodic data polling and event emission
│   └── handler.go               # Gin handler for /ws upgrade
├── services/                    # Existing services (unchanged)
├── database/                    # Existing storage (unchanged)
├── interfaces/                  # Existing interfaces (unchanged)
├── models/                      # Existing models (unchanged)
└── web/                         # NEW: frontend static files
    ├── index.html               # Single HTML entry point
    ├── css/
    │   └── dashboard.css        # Dashboard styles
    └── js/
        ├── app.js               # Main application logic
        ├── websocket.js          # WebSocket client with reconnect
        ├── components/           # UI component modules
        │   ├── portfolio.js      # Portfolio panel
        │   ├── activity.js       # Activity feed panel
        │   ├── market.js         # Market data panel
        │   ├── news.js           # News panel
        │   ├── history.js        # Trade history panel
        │   └── controls.js       # Emergency controls panel
        └── lib/
            └── alpine.min.js    # Alpine.js (vendored, no build step)
```

### Structure Rationale

- **`dashboard/` package:** All WebSocket/real-time code lives in one new Go package, separate from existing controllers. This avoids modifying any existing code except `main.go` (to wire the hub in) and keeps the real-time concern isolated.
- **`web/` directory:** Already stubbed out in `main.go` line 229 (`router.Static("/dashboard", "./web")`). No build step required -- plain HTML/JS/CSS served as static files. Alpine.js vendored locally so no CDN dependency.
- **Existing code untouched:** The REST controllers, services, models, interfaces, and database packages require zero modifications. The dashboard reads from existing services through the event bus and ticker.

## Architectural Patterns

### Pattern 1: WebSocket Hub (Central Broadcast)

**What:** A single goroutine manages all connected WebSocket clients. It receives events from the event bus and broadcasts them to registered clients. Clients can subscribe to specific event topics.

**When to use:** Always -- this is the standard Go pattern for WebSocket broadcasting. Gorilla/websocket's own chat example uses this exact pattern.

**Trade-offs:** Simple and battle-tested. Single hub goroutine means no lock contention. For a single-user dashboard this is massive overkill, but the pattern is so simple that simplifying it further would actually add complexity.

**Example:**
```go
package dashboard

import (
    "sync"
    "github.com/gorilla/websocket"
)

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan *Event
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

type Client struct {
    hub    *Hub
    conn   *websocket.Conn
    send   chan []byte
    topics map[string]bool // subscribed topics
}

type Event struct {
    Topic   string      `json:"topic"`
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
    Time    int64       `json:"time"`
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan *Event, 256),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
        case event := <-h.broadcast:
            data, _ := json.Marshal(event)
            for client := range h.clients {
                // Send to all clients (single user, no topic filtering needed initially)
                select {
                case client.send <- data:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}
```

### Pattern 2: Dashboard Ticker (Polling-to-Push Bridge)

**What:** A background goroutine that periodically calls existing services (account, positions, activity) and publishes state changes as events on the hub's broadcast channel. Different data types poll at different intervals.

**When to use:** When the data sources are pull-based (REST API, database queries) but the consumer needs push-based updates. This is the case here: Alpaca's API is polled, SQLite is queried, and the dashboard needs live updates.

**Trade-offs:** Adds polling load on the Alpaca API. Mitigated by using sensible intervals (account: 30s, positions: 10s, activity: 5s). The existing `startPositionMonitor` in `main.go` already polls every 5 minutes -- the ticker replaces and improves this pattern.

**Example:**
```go
package dashboard

import (
    "context"
    "time"
)

type Ticker struct {
    hub             *Hub
    tradingService  interfaces.TradingService
    activityLogger  *services.ActivityLogger
    positionManager *services.PositionManager
}

func (t *Ticker) Run(ctx context.Context) {
    accountTicker := time.NewTicker(30 * time.Second)
    positionTicker := time.NewTicker(10 * time.Second)
    activityTicker := time.NewTicker(5 * time.Second)

    defer accountTicker.Stop()
    defer positionTicker.Stop()
    defer activityTicker.Stop()

    // Send initial state immediately on connect
    t.publishAccount()
    t.publishPositions()
    t.publishActivity()

    for {
        select {
        case <-ctx.Done():
            return
        case <-accountTicker.C:
            t.publishAccount()
        case <-positionTicker.C:
            t.publishPositions()
        case <-activityTicker.C:
            t.publishActivity()
        }
    }
}

func (t *Ticker) publishPositions() {
    positions, err := t.tradingService.GetPositions(context.Background())
    if err != nil {
        return
    }
    t.hub.broadcast <- &Event{
        Topic:   "positions",
        Type:    "snapshot",
        Payload: positions,
        Time:    time.Now().UnixMilli(),
    }
}
```

### Pattern 3: No-Build Frontend with Alpine.js

**What:** Serve plain HTML/JS/CSS from `./web` using Gin's static file server. Use Alpine.js (15KB, vendored) for reactive UI binding. No npm, no webpack, no build step. The Go binary + `web/` folder is the entire deployment.

**When to use:** Single-user dashboards, internal tools, admin panels where developer experience matters more than framework ecosystem. A trading bot dashboard is exactly this use case.

**Trade-offs:** No component reuse across projects, no TypeScript type safety, no hot module replacement during development. These are non-issues for a single-user dashboard with ~6 panels. The simplicity of "edit HTML, refresh browser" massively outweighs the benefits of a React/Vue setup for this scope.

**Why not HTMX:** HTMX is server-rendered -- it swaps HTML fragments from the server. But this dashboard needs WebSocket-pushed JSON data rendered client-side. Alpine.js handles client-side reactivity with `x-data`, `x-for`, `x-text` directives that bind directly to JavaScript state updated by WebSocket messages. HTMX would fight this pattern.

**Why not React/Vue:** Adds npm, node_modules, a build step, and tooling complexity. The dashboard has ~6 panels with straightforward data binding. Alpine.js handles this in vanilla-level simplicity with reactive bindings.

## Data Flow

### Real-time Push Flow (Primary)

```
Alpaca API / SQLite DB
        |
        v
  [Dashboard Ticker]  (polls at intervals: 10s positions, 30s account, 5s activity)
        |
        v  (publishes Event structs)
  [Hub broadcast channel]
        |
        v  (JSON marshal + write to each client)
  [WebSocket Client goroutine]
        |
        v  (WebSocket frame)
  [Browser WebSocket.onmessage]
        |
        v  (update Alpine.js reactive data)
  [UI re-renders affected panels]
```

### Action Flow (User-Initiated)

```
User clicks "Sell All" / "Cancel Order"
        |
        v
  [Browser fetch() to REST API]  POST /api/v1/orders/sell, DELETE /api/v1/orders/:id
        |
        v
  [Existing REST Controller]  (unchanged code)
        |
        v
  [Trading Service -> Alpaca API]
        |
        v  (next ticker cycle picks up the change)
  [Dashboard Ticker sees updated positions/orders]
        |
        v
  [Broadcast updated state via WebSocket]
```

### Initial Load Flow

```
Browser navigates to /dashboard
        |
        v
  [Gin serves index.html + JS/CSS from ./web]
        |
        v
  [Browser JS opens WebSocket to /ws]
        |
        v
  [Hub registers client]
        |
        v
  [Ticker immediately sends current snapshots]
        |
        v
  [Dashboard renders with full state]
```

### Key Data Flows

1. **Portfolio updates:** Ticker polls `tradingService.GetPositions()` + `tradingService.GetAccount()` every 10-30s, broadcasts `{topic: "positions", type: "snapshot"}` and `{topic: "account", type: "snapshot"}` events.

2. **Activity feed:** Ticker polls `activityLogger.GetCurrentLog()` every 5s, diffs against previous state, broadcasts `{topic: "activity", type: "snapshot"}` with latest activities.

3. **Emergency controls:** User clicks button, browser sends REST request directly to existing API endpoints (sell position, cancel order). No WebSocket needed for actions -- REST is simpler and provides confirmation responses. Ticker picks up resulting state changes on next cycle.

4. **News feed:** Ticker polls `newsService.GetLatestNews()` every 60s (news is slow-moving), broadcasts `{topic: "news", type: "snapshot"}`.

5. **Trade history:** Initial load fetches from REST API `GET /api/v1/orders?status=filled`. Updated via WebSocket when new orders fill (detected by ticker comparing order counts).

## Event Message Schema

All WebSocket messages follow a consistent envelope format:

```json
{
    "topic": "positions",
    "type": "snapshot",
    "payload": { ... },
    "time": 1707600000000
}
```

**Topics:**

| Topic | Payload | Interval | Description |
|-------|---------|----------|-------------|
| `account` | `Account` object | 30s | Cash, portfolio value, buying power |
| `positions` | `[]Position` array | 10s | All current positions with P&L |
| `orders` | `[]Order` array | 10s | Open/recent orders |
| `activity` | `DailyActivityLog` object | 5s | Today's activity feed |
| `news` | `[]NewsItem` array | 60s | Latest news headlines |
| `managed` | `[]ManagedPosition` array | 10s | Managed positions with risk levels |
| `status` | `ServerStatus` object | 60s | Bot health, uptime, connection status |

**Types:**
- `snapshot` -- Full state replacement (client replaces local data entirely)
- `error` -- Error notification (display to user)

Using `snapshot` type exclusively keeps the client simple: no delta merging, no out-of-order concerns. The data sizes are small (single user, dozens of positions at most) so bandwidth is irrelevant.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1 user (this project) | Hub + ticker + static files. Everything in-process. No auth, no sessions. |
| 2-5 users | Add basic auth middleware. Hub already handles multiple clients. No changes needed. |
| 50+ users | Would need to rate-limit Alpaca API calls (shared ticker is fine). Consider SSE instead of WebSocket if read-only. Not relevant for this project. |

### Scaling Priorities

1. **First bottleneck:** Alpaca API rate limits. The ticker polls every 10-30s which is well within limits for a single user. If multiple tickers were needed, centralize polling in one ticker and broadcast to all.
2. **Second bottleneck:** SQLite write contention. Already mitigated by GORM's connection pooling. Not a concern at single-user scale.

## Anti-Patterns

### Anti-Pattern 1: Polling from the Browser

**What people do:** Use `setInterval` + `fetch()` in the browser to poll REST endpoints every few seconds.
**Why it's wrong:** Creates N HTTP requests per interval per panel. Each request has TCP overhead, header overhead, and creates load on the Gin router. Wasteful when nothing has changed.
**Do this instead:** Single WebSocket connection. Server pushes only when it has new data. Browser renders from in-memory state.

### Anti-Pattern 2: Sending Actions over WebSocket

**What people do:** Route user actions (buy, sell, cancel) through the WebSocket connection instead of REST.
**Why it's wrong:** WebSocket has no built-in request/response semantics. You need to add correlation IDs, timeouts, and error handling that HTTP already provides. The existing REST controllers already handle all actions perfectly.
**Do this instead:** Use REST (`fetch()`) for all user-initiated actions. Use WebSocket only for server-to-client push. The ticker will pick up state changes caused by actions on the next cycle (within seconds).

### Anti-Pattern 3: Building a React/Vue SPA with npm Build Pipeline

**What people do:** Add a `frontend/` directory with `package.json`, webpack/vite config, hundreds of npm dependencies, TypeScript config, and a build step that produces a `dist/` folder.
**Why it's wrong for this project:** Adds 10x complexity for a 6-panel single-user dashboard. Creates a separate dev server workflow (port 3000 proxy to port 4534). Makes deployment harder (need Node.js to build). The dashboard is an internal tool, not a product.
**Do this instead:** Plain HTML + Alpine.js + vendored dependencies. Edit, refresh, done. Ship the `web/` folder alongside the Go binary.

### Anti-Pattern 4: Streaming Every Price Tick via WebSocket

**What people do:** Connect to Alpaca's streaming WebSocket, pipe every tick through to the browser.
**Why it's wrong:** Creates massive message volume the browser cannot meaningfully render. Positions update at most every few seconds visually. The existing `StreamBars` stub in `alpaca_data.go` is unimplemented for good reason -- periodic snapshots are better for a dashboard.
**Do this instead:** Poll at sensible intervals (10-30s), send complete snapshots. Visually indistinguishable from real-time for a portfolio dashboard.

### Anti-Pattern 5: Complex State Synchronization with Deltas

**What people do:** Send only changed fields (deltas) to minimize bandwidth, requiring client-side merge logic.
**Why it's wrong for this project:** Single user, small data (dozens of positions). Full snapshots are a few KB at most. Delta logic adds complexity for zero benefit. Risk of client state diverging from server state.
**Do this instead:** Always send full snapshots. Client replaces its entire state on each message. Simple, correct, debuggable.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Alpaca Trading API | Polled by existing services, results broadcast via ticker | Rate limits: 200 req/min (paper). 10s polling = 6 req/min, well within limits. |
| Alpaca Market Data | Polled by DataService for quotes/bars | Already working. Ticker wraps existing calls. |
| Google News RSS | Polled by NewsService | Already working. 60s polling interval is plenty. |
| MarketWatch RSS | Polled by NewsService | Already working. Same 60s interval. |
| Gemini AI | Called on-demand via REST, not via ticker | AI analysis is too slow/expensive for periodic polling. User triggers via dashboard button -> REST call. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| main.go <-> dashboard package | Dependency injection at startup | Hub and Ticker receive service interfaces in constructors. Same pattern as existing controllers. |
| dashboard/hub <-> dashboard/ticker | Hub's broadcast channel | Ticker writes events; hub reads and distributes. One-directional, channel-based. |
| dashboard/handler <-> Gin router | Single route registration | `router.GET("/ws", dashboard.HandleWebSocket(hub))` |
| Browser <-> REST API | Standard HTTP fetch | Dashboard JS calls existing `/api/v1/*` endpoints for actions. No changes to controllers needed. |
| Browser <-> WebSocket | Persistent connection at `/ws` | Auto-reconnect with exponential backoff on client side. |

## Build Order (Dependency Chain)

The components have a clear dependency order that determines the build sequence:

```
Phase 1: WebSocket Infrastructure (no UI needed to test)
    dashboard/events.go     -- Event types (no dependencies)
    dashboard/hub.go        -- Hub (depends on events)
    dashboard/client.go     -- Client read/write pumps (depends on hub)
    dashboard/handler.go    -- Gin route handler (depends on hub, client)

Phase 2: Data Bridge (connects existing services to hub)
    dashboard/ticker.go     -- Polls services, publishes events (depends on hub, existing services)
    Modify cmd/bot/main.go  -- Wire hub + ticker into startup

Phase 3: Frontend Shell (can now see live data in browser console)
    web/index.html          -- HTML structure with Alpine.js
    web/js/websocket.js     -- WebSocket client with reconnect
    web/js/app.js           -- Alpine.js data store, message routing

Phase 4: Dashboard Panels (one at a time, each independently useful)
    web/js/components/portfolio.js   -- Portfolio/account panel (most critical)
    web/js/components/controls.js    -- Emergency controls (most urgent safety feature)
    web/js/components/activity.js    -- Activity feed
    web/js/components/market.js      -- Market data
    web/js/components/news.js        -- News feed
    web/js/components/history.js     -- Trade history

Phase 5: Polish
    web/css/dashboard.css   -- Styling, responsive layout
    Error handling, reconnection UI, loading states
```

**Why this order:**
1. WebSocket infra first because it can be tested with `wscat` or browser console before any UI exists.
2. Ticker second because it bridges existing services to the hub, and its output can be verified by watching WebSocket messages.
3. Frontend shell third because it wires up Alpine.js and WebSocket client -- at this point you can see raw JSON in the browser.
4. Panels are independent of each other. Portfolio and emergency controls come first because they are the most operationally critical (seeing positions, killing trades).
5. Polish last because everything is functional before styling matters.

## Sources

- [Gin WebSocket example (gorilla/websocket)](https://github.com/gin-gonic/examples/blob/master/websocket/server/server.go) -- HIGH confidence, official Gin examples repo
- [gorilla/websocket chat hub pattern](https://github.com/gorilla/websocket/blob/main/examples/chat/README.md) -- HIGH confidence, canonical hub pattern
- [Gin static file serving docs](https://gin-gonic.com/en/docs/examples/serving-static-files/) -- HIGH confidence, official docs
- [WebSocket architecture best practices (Ably)](https://ably.com/topic/websocket-architecture-best-practices) -- MEDIUM confidence, well-regarded real-time provider
- [Building WebSocket for Notifications with GoLang and Gin](https://medium.com/@abhishekranjandev/building-a-production-grade-websocket-for-notifications-with-golang-and-gin-a-detailed-guide-5b676dcfbd5a) -- MEDIUM confidence, community article with production patterns
- [coder/websocket as modern gorilla alternative](https://github.com/coder/websocket) -- MEDIUM confidence, maintained fork. Note: gorilla/websocket is still functional and widely used; switching is optional.
- [Alpine.js](https://alpinejs.dev/) -- HIGH confidence, well-established lightweight framework
- [HTMX + Alpine.js comparison](https://www.infoworld.com/article/3856520/htmx-and-alpine-js-how-to-combine-two-great-lean-front-ends.html) -- MEDIUM confidence, informed the Alpine.js-only decision
- [Existing codebase analysis](./cmd/bot/main.go) -- HIGH confidence, direct code review

---
*Architecture research for: Prophet Trader real-time dashboard*
*Researched: 2026-02-11*
