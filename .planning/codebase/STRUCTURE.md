# Codebase Structure

**Analysis Date:** 2026-02-11

## Directory Layout

```
prophet-trader/
├── cmd/bot/                    # Go backend entry point
│   └── main.go                 # Server startup, router setup, workers
├── config/                     # Configuration management
│   └── config.go               # Env var loading, AppConfig struct
├── controllers/                # HTTP request handlers
│   ├── order_controller.go     # Trading operations (buy, sell, manage positions)
│   ├── position_controller.go  # Position queries and management
│   ├── intelligence_controller.go # News, analysis, insights
│   ├── news_controller.go      # News endpoints
│   └── activity_controller.go  # Activity logging
├── database/                   # Data persistence
│   └── storage.go              # SQLite via GORM, migrations, CRUD
├── interfaces/                 # Go interface contracts
│   ├── trading.go              # TradingService, DataService, StorageService
│   └── options.go              # Options trading types
├── models/                     # Database models (GORM)
│   └── models.go               # DBOrder, DBBar, DBPosition, DBTrade, etc.
├── services/                   # Business logic layer
│   ├── alpaca_trading.go       # Alpaca order execution
│   ├── alpaca_data.go          # Alpaca market data queries
│   ├── alpaca_options_data.go  # Alpaca options chain, quotes
│   ├── position_manager.go     # Risk management, position lifecycle
│   ├── news_service.go         # News aggregation
│   ├── gemini_service.go       # Google Gemini AI integration
│   ├── stock_analysis_service.go # Combined analysis logic
│   ├── technical_analysis.go   # Indicator calculations
│   └── activity_logger.go      # Trade activity recording
├── data/                       # SQLite database (generated)
│   └── prophet_trader.db       # Main trading data store
├── news_summaries/             # Generated news summaries (JSON)
├── decisive_actions/           # Generated trading decisions (JSON)
├── activity_logs/              # Session activity logs (JSON)
├── mcp-server.js               # Node.js MCP server (AI interface)
├── vectorDB.js                 # Vector DB for trade similarity search
├── backfill_embeddings.js      # Utility to pre-compute embeddings
├── package.json                # Node.js dependencies
├── go.mod                      # Go module definition
├── go.sum                       # Go dependencies lock file
└── .env                        # Environment configuration (not committed)
```

## Directory Purposes

**cmd/bot/:**
- Purpose: Go application entry point and HTTP server setup
- Contains: Main function, router initialization, background worker launches
- Key files: `main.go`

**config/:**
- Purpose: Centralized configuration management from environment
- Contains: Config struct, env var loading with godotenv, defaults
- Key files: `config.go`

**controllers/:**
- Purpose: HTTP request handlers, orchestration layer
- Contains: Request binding, validation, service orchestration, response formatting
- Key files: All `*_controller.go` files

**database/:**
- Purpose: Data persistence using SQLite and GORM ORM
- Contains: Connection management, migrations, CRUD operations, cleanup routines
- Key files: `storage.go`

**interfaces/:**
- Purpose: Go interface definitions for dependency injection
- Contains: Service contracts, data structures, abstractions
- Key files: `trading.go`, `options.go`

**models/:**
- Purpose: Database entity definitions (GORM structs)
- Contains: Order, Bar, Position, Trade, Account, Signal, ManagedPosition models
- Key files: `models.go`

**services/:**
- Purpose: Business logic, external API integration, domain operations
- Contains: Alpaca API wrappers, news aggregation, analysis, AI integration, position management
- Key files: All service implementations

**data/:**
- Purpose: SQLite database file storage
- Contains: prophet_trader.db (auto-created on startup)
- Generated: Yes, automatically created by GORM migrations

**news_summaries/:**
- Purpose: Cache Gemini-summarized news for analysis
- Contains: JSON files with market sentiment, themes, stock mentions
- Generated: Yes, created by intelligence controller

**decisive_actions/:**
- Purpose: Store AI-generated trading decisions for audit trail
- Contains: JSON decision files with reasoning, strategy, targets
- Generated: Yes, created by MCP server and intelligence controller

**activity_logs/:**
- Purpose: Session-based trading activity recording
- Contains: JSON files with trade events, fills, position changes
- Generated: Yes, created by activity logger service

## Key File Locations

**Entry Points:**
- `cmd/bot/main.go`: Go backend startup
- `mcp-server.js`: Node.js MCP server startup
- `package.json`: Node.js scripts (`npm start` runs mcp-server.js)

**Configuration:**
- `config/config.go`: AppConfig struct, environment loading
- `.env`: Environment variables (ALPACA_API_KEY, GEMINI_API_KEY, etc.) - not committed

**Core Logic:**
- `services/alpaca_trading.go`: Order execution (PlaceOrder, CancelOrder, ListOrders)
- `services/alpaca_data.go`: Market data queries (GetHistoricalBars, GetLatestBar, GetLatestQuote)
- `services/position_manager.go`: Risk management and position lifecycle
- `services/gemini_service.go`: AI-powered news cleaning and analysis

**Testing:**
- No test files currently present (test framework not implemented)

## Naming Conventions

**Files:**
- Go service files: `{domain}_{type}.go` (e.g., `alpaca_trading.go`, `position_manager.go`)
- Go controller files: `{domain}_controller.go` (e.g., `order_controller.go`)
- Node.js files: camelCase.js (e.g., `mcp-server.js`, `vectorDB.js`)
- Database: `prophet_trader.db` (SQLite)
- Logs: `session-{timestamp}.json` (activity logs)
- Decisions: `decision-{id}.json` (trade decisions)

**Directories:**
- Go packages: lowercase, descriptive (e.g., `services`, `controllers`, `interfaces`)
- Data directories: descriptive plural or snake_case (e.g., `news_summaries`, `activity_logs`)

**Functions:**
- Go: PascalCase (exported) or camelCase (private)
- Node.js: camelCase
- Controllers/Handlers: `Handle{Action}` or `{Action}` (e.g., `HandleBuy`, `Buy`)

**Variables:**
- Go: camelCase (following Go conventions)
- Struct fields: PascalCase (exported)
- JSON tags: snake_case (e.g., `json:"filled_qty"`)

**Types:**
- Go structs: PascalCase with DB/Request/Response suffixes (e.g., DBOrder, BuyRequest, OrderResult)
- Go interfaces: -Service suffix (e.g., TradingService, DataService)

## Where to Add New Code

**New Trading Feature:**
- Primary code: `services/{feature}_service.go` (new service file)
- Tests: `services/{feature}_service_test.go` (when tests are added)
- Controller: Add method to `controllers/order_controller.go` or create `controllers/{feature}_controller.go`
- Routes: Add route handler in `cmd/bot/main.go` setupRouter function
- Models: Add GORM struct to `models/models.go` if persistence needed

**New Trading Strategy:**
- Implementation: Create `services/{strategy_name}_strategy.go` implementing StrategyExecutor interface
- Register: Add initialization in `cmd/bot/main.go` main function
- Controller: Add endpoint in OrderController or create new controller

**New Analytics/Intelligence Feature:**
- Service: Add method to `services/gemini_service.go` or create new service
- Controller: Add method to `controllers/intelligence_controller.go`
- Route: Add to `/api/v1/intelligence` endpoint group in main.go
- Storage: Add model to `models/models.go` if results need persistence

**New External Integration (News, Data, AI):**
- Service: Create `services/{provider}_service.go`
- Interface: If reusable, add interface to `interfaces/trading.go`
- Initialization: Add to main.go service setup
- Error Handling: Implement retry logic and graceful degradation

**Database Changes:**
- Models: Modify `models/models.go` (add struct fields)
- Migration: GORM auto-migration happens on startup via `db.AutoMigrate()`
- Queries: Add methods to `database/storage.go` as needed
- Testing: Manual verification via sqlite3 CLI or GORM logging

**New API Endpoint:**
- Controller Method: Add to appropriate controller file (e.g., OrderController.NewAction)
- Route: Register in setupRouter() under appropriate API group
- Request Struct: Define near controller (e.g., NewActionRequest)
- Response: Use standard OrderResult or create custom struct
- Documentation: Add to MCP server tools if AI-accessible

**Utilities:**
- Shared helpers: Create `services/utils.go` or `utils.go` in root
- Technical helpers: Add to `services/technical_analysis.go`
- Log formatting: Add to existing service or create `services/logging.go`

**MCP Tools (AI Integration):**
- New tool: Add handler in `mcp-server.js` under CallToolRequestSchema
- Tool definition: Add to tools array in ListToolsRequestSchema
- Implementation: Call trading bot API via `callTradingBot()` helper
- Vector features: Use `vectorDB.js` functions for similarity search

## Special Directories

**data/:**
- Purpose: SQLite database storage
- Generated: Yes, created by GORM on first startup
- Committed: No (added to .gitignore typically)
- Size: Grows with trading activity and historical data

**news_summaries/:**
- Purpose: Cache generated news analysis to avoid re-summarizing
- Generated: Yes, by intelligence controller
- Committed: No (typically ignored)
- Cleanup: Manual or via retention policy

**decisive_actions/:**
- Purpose: Audit trail of all trading decisions and reasoning
- Generated: Yes, by MCP server and controllers
- Committed: No (typically ignored)
- Retention: Should be archived periodically for compliance

**activity_logs/:**
- Purpose: Per-session trading activity and event log
- Generated: Yes, by activity logger service
- Committed: No (typically ignored)
- Format: JSON with timestamp, action, details
- Usage: Post-trading analysis, debugging

**.claude/:**
- Purpose: GSD (Get Smart Done) agent state and configuration
- Generated: Yes, by GSD orchestrator
- Committed: No (internal tooling)

**.omc/:**
- Purpose: OMC (OpenMCP) configuration state
- Generated: Yes, by MCP tooling
- Committed: No (internal tooling)

**.planning/codebase/:**
- Purpose: Architecture and structure documentation
- Generated: Yes, by GSD codebase mapper
- Committed: Yes, for reference
- Contents: ARCHITECTURE.md, STRUCTURE.md, CONVENTIONS.md, TESTING.md, CONCERNS.md, STACK.md, INTEGRATIONS.md

---

*Structure analysis: 2026-02-11*
