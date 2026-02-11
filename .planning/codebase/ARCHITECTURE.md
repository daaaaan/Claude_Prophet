# Architecture

**Analysis Date:** 2026-02-11

## Pattern Overview

**Overall:** Microservices/Layered Architecture with hybrid Go backend and Node.js AI integration layer

**Key Characteristics:**
- Go backend exposes REST API via Gin framework for trading operations
- Node.js MCP server acts as AI control plane with vector-based decision history
- Clear separation of concerns: Services → Controllers → HTTP Endpoints
- Interface-based design for loose coupling between layers
- Background workers for position monitoring, data cleanup, and activity logging

## Layers

**API Layer (HTTP REST):**
- Purpose: Expose trading, market data, news, and intelligence endpoints
- Location: `cmd/bot/main.go` (setupRouter function)
- Contains: Gin route handlers, request/response handling, CORS middleware
- Depends on: Controllers layer
- Used by: AI agents, external clients, MCP server

**Controller Layer:**
- Purpose: Orchestrate business logic, convert HTTP requests to service calls
- Location: `controllers/` directory
- Contains: `order_controller.go`, `news_controller.go`, `intelligence_controller.go`, `position_controller.go`, `activity_controller.go`
- Depends on: Services layer, interfaces
- Used by: API layer, background workers

**Service Layer (Business Logic):**
- Purpose: Execute domain logic, coordinate external APIs, manage state
- Location: `services/` directory
- Contains: Trading service, data service, news service, Gemini AI service, position manager, activity logger, technical analysis
- Depends on: Interfaces, database layer, external APIs
- Used by: Controllers

**Interface Layer (Contracts):**
- Purpose: Define contracts between layers using Go interfaces
- Location: `interfaces/` directory
- Contains: `trading.go` (TradingService, DataService, StorageService, StrategyExecutor), `options.go` (options trading contracts)
- Depends on: None (interfaces only)
- Used by: All other layers

**Database Layer (Data Persistence):**
- Purpose: Persist orders, positions, bars, trades, signals, account snapshots
- Location: `database/storage.go`
- Contains: SQLite via GORM, migrations, CRUD operations
- Depends on: Models
- Used by: Services, controllers

**Models Layer (Data Structures):**
- Purpose: Define domain entities for database and API
- Location: `models/models.go`
- Contains: DBOrder, DBBar, DBPosition, DBTrade, DBAccountSnapshot, DBSignal, DBManagedPosition
- Depends on: GORM
- Used by: Database, services

**AI Integration Layer (Node.js):**
- Purpose: Provide AI-driven trading decisions via MCP protocol
- Location: `mcp-server.js`, `vectorDB.js`
- Contains: MCP server tools, Gemini API integration, vector embeddings, trade history similarity search
- Depends on: Trading Bot REST API, local SQLite vector DB
- Used by: External AI agents (Claude, etc.)

**Vector Database Layer:**
- Purpose: Store and search trade decisions by semantic similarity
- Location: `vectorDB.js`
- Contains: SQLite with sqlite-vec extension, local embeddings via Xenova transformer
- Depends on: better-sqlite3, sqlite-vec, @xenova/transformers
- Used by: MCP server for finding similar past trades

## Data Flow

**Trading Order Flow:**

1. User/AI calls `POST /api/v1/orders/buy` or `/api/v1/orders/sell`
2. OrderController receives request, validates, applies defaults
3. Controller calls TradingService.PlaceOrder() with Order struct
4. AlpacaTradingService converts to Alpaca API format, submits order
5. Response returned with OrderID and status
6. StorageService saves order to SQLite for audit trail
7. Background position monitor polls for updates

**Position Management Flow (Managed Positions):**

1. User calls `POST /api/v1/positions/managed` with risk parameters
2. PositionController delegates to PositionManager
3. PositionManager creates entry order (market or limit)
4. On entry fill: Automatically creates stop-loss and take-profit orders
5. MonitorPositions() background worker watches price updates
6. On trigger: Updates trailing stops, executes partial exits, closes position
7. Final state saved to database for performance analysis

**Market Intelligence Flow:**

1. Request arrives at IntelligenceController
2. Controller coordinates:
   - NewsService.GetNews() → fetches headlines
   - TechnicalAnalysisService.Analyze() → calculates indicators
   - StockAnalysisService.AnalyzeStock() → combines news + technical
3. GeminiService.CleanNewsForTrading() → summarizes via AI
4. Results cached and returned

**Vector Search Flow (via MCP):**

1. MCP client calls `find_similar_trades` tool
2. mcp-server.js receives request
3. vectorDB.getEmbedding() creates embedding for query text
4. Database executes vec_distance_search
5. Returns top-N similar past trades with their outcomes
6. AI agent uses history to inform new decisions

**State Management:**

- **Account State:** Cached in memory from Alpaca, updated on trades
- **Positions:** Queried real-time from Alpaca, cached in SQLite snapshots
- **Orders:** Stored in SQLite with sync to Alpaca status
- **Market Data:** Streamed from Alpaca, cached in SQLite by timeframe
- **Trade History:** Embedded and stored in vector DB for similarity search
- **Activity Log:** Streamed JSON to `activity_logs/` directory for audit

## Key Abstractions

**TradingService Interface (`interfaces/trading.go`):**
- Purpose: Abstract away Alpaca API specifics
- Examples: `services/alpaca_trading.go`, `services/alpaca_options_data.go`
- Pattern: Dependency injection via interface; allows swapping implementations

**DataService Interface (`interfaces/trading.go`):**
- Purpose: Abstract market data access (bars, quotes, trades)
- Examples: `services/alpaca_data.go`
- Pattern: Same interface for historical and real-time data

**StorageService Interface (`interfaces/trading.go`):**
- Purpose: Abstract persistence layer
- Examples: `database/storage.go`
- Pattern: GORM-based SQLite implementation

**ManagedPosition Abstraction (`services/position_manager.go`):**
- Purpose: Encapsulate complex position lifecycle with risk management
- Pattern: State machine (PENDING → ACTIVE → PARTIAL → CLOSED/STOPPED_OUT)
- Handles: Entry, stop-loss, take-profit, trailing stops, partial exits

## Entry Points

**Go Backend:**
- Location: `cmd/bot/main.go`
- Triggers: Direct execution (`go run ./cmd/bot`)
- Responsibilities:
  - Load config from `.env`
  - Initialize all services and controllers
  - Start HTTP server on configurable port (default 4534)
  - Launch background workers for position monitoring, data cleanup, activity logging
  - Setup graceful shutdown

**Node.js MCP Server:**
- Location: `mcp-server.js`
- Triggers: Direct execution (`node mcp-server.js`)
- Responsibilities:
  - Implement Model Context Protocol for Claude/AI agents
  - Expose tools: account, positions, orders, news, trade history, managed positions
  - Call trading bot REST API for execution
  - Manage vector DB for trade similarity search

**Background Workers:**
- `MonitorPositions()`: Runs in `services/position_manager.go`, polls Alpaca for position updates every N seconds
- `startPositionMonitor()`: Runs in main.go, polls managed positions
- `startDataCleanup()`: Runs in main.go, deletes old data based on retention policy
- `ActivityLogger.StartSession()`: Records trading activity to JSON files

## Error Handling

**Strategy:** Layered error propagation with contextual logging

**Patterns:**

- Controllers catch service errors, return HTTP 4xx/5xx with message
- Services log errors with logrus including context (symbol, quantity, etc.)
- Database errors wrapped with descriptive messages
- API client errors (Alpaca, Gemini, news sources) caught and re-thrown with context
- Background workers log errors but continue operating (resilient failure)

Example from `controllers/order_controller.go`:
```go
if err != nil {
  oc.logger.WithFields(logrus.Fields{
    "symbol": req.Symbol,
    "qty":    req.Qty,
  }).Error("Order placement failed")
  return nil, err
}
```

## Cross-Cutting Concerns

**Logging:**
- Framework: `github.com/sirupsen/logrus`
- Pattern: Structured logging with fields (symbol, qty, orderID, etc.)
- Configuration: `LOG_LEVEL` env var, configurable timestamp format
- Used in: All services, controllers, database operations

**Validation:**
- Framework: Gin's binding tag validation (`binding:"required,gt=0"`)
- Pattern: Request structs with validation tags; returned 400 Bad Request if invalid
- Example: `Qty float64 \`json:"qty" binding:"required,gt=0"\`` in order requests

**Authentication:**
- Strategy: API key-based (Alpaca, Gemini) via environment variables
- MCP: STDIO transport (inherits security model from calling process)
- No HTTP-level auth on REST API (assumes private network)

---

*Architecture analysis: 2026-02-11*
