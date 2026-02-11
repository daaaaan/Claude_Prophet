# External Integrations

**Analysis Date:** 2026-02-11

## APIs & External Services

**Trading & Market Data:**
- Alpaca Markets - Core trading platform and market data
  - SDK/Client: `github.com/alpacahq/alpaca-trade-api-go/v3`
  - Auth: `ALPACA_API_KEY`, `ALPACA_SECRET_KEY` environment variables
  - Base URL configurable via `ALPACA_BASE_URL` env var (defaults to `https://paper-api.alpaca.markets`)
  - Paper trading supported via `ALPACA_PAPER` flag
  - Usage: `services/alpaca_trading.go`, `services/alpaca_data.go`
  - Endpoints used:
    - Order placement: `/orders` (POST)
    - Order cancellation: `/orders/{id}` (DELETE)
    - Order retrieval: `/orders` (GET)
    - Positions: `/positions` (GET)
    - Account: `/account` (GET)
    - Market data: `/stocks/{symbol}/quotes`, `/stocks/{symbol}/bars`
    - Options: `/options/snapshots/{symbol}` (custom REST calls in `services/alpaca_trading.go` lines 299-327)

**AI & Language Models:**
- Google Gemini API - News analysis and market intelligence
  - SDK/Client: `@google/generative-ai` (Node.js, version ^0.24.1)
  - Auth: `GEMINI_API_KEY` environment variable (optional)
  - Model: `gemini-2.0-flash-exp` (as of `mcp-server.js` and `services/gemini_service.go`)
  - Endpoints: `https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent`
  - Usage:
    - `services/gemini_service.go` - CleanNewsForTrading() function for AI-powered news summarization
    - `mcp-server.js` - Gemini integration for MCP tools
  - Purpose: Token-efficient news summaries optimized for trading decisions

**Local ML Models:**
- Xenova/transformers (local embedding model) - No API key required
  - SDK/Client: `@xenova/transformers` (Node.js, version ^2.17.2)
  - Model: `Xenova/all-MiniLM-L6-v2` (384-dimensional embeddings)
  - Usage: `vectorDB.js` (lines 4, 16)
  - Purpose: Generate embeddings for trade decision similarity search

## Data Storage

**Databases:**
- SQLite with vector extensions
  - Connection: Local file at `./data/prophet_trader.db` (configurable via `DATABASE_PATH` env var)
  - Client (Go): `gorm.io/gorm` (version 1.25.12) with `gorm.io/driver/sqlite` (1.5.6)
  - Client (Node.js): `better-sqlite3` (version 12.4.6) and `sqlite-vec` (0.1.7-alpha.2)
  - Schema: Defined in `database/storage.go` via GORM auto-migration
  - Tables:
    - `DBOrder` - Placed orders
    - `DBBar` - Price bars/OHLCV data
    - `DBPosition` - Position snapshots
    - `DBTrade` - Trade records
    - `DBAccountSnapshot` - Account state history
    - `DBSignal` - Trading signals
    - `DBManagedPosition` - Managed position tracking
    - `trade_embeddings` - Trade decision metadata with embeddings (Node.js)
    - `trade_vectors` - Vector index using vec0 virtual table (Node.js)

**File Storage:**
- Local filesystem only
  - Activity logs: `./activity_logs/` (see `cmd/bot/main.go` line 110)
  - News summaries: `./news_summaries/` (MCP server, `mcp-server.js` line 4)
  - Decisive actions: `./decisive_actions/` (MCP server, `mcp-server.js` line 5)
  - Database: `./data/prophet_trader.db`

**Caching:**
- None detected - no Redis, Memcached, or other caching layer

## Authentication & Identity

**Auth Provider:**
- Custom - No centralized auth provider
- Alpaca API: API key/secret pair authentication
- Google Gemini: API key authentication
- No user authentication implemented - assumes single-user bot deployment

## News & Market Intelligence Sources

**News Feeds (RSS):**
- Google News - Free RSS feeds
  - General news: `https://news.google.com/rss?hl=en-US&gl=US&ceid=US:en`
  - Topic feeds: Business, Technology, Science, Health, Sports, etc.
  - Search feeds: `https://news.google.com/rss/search?q={query}...`
  - Client: `services/news_service.go` (lines 68-104)

- MarketWatch - Financial news feeds
  - Top stories: `https://feeds.content.dowjones.io/public/rss/mw_topstories`
  - Real-time headlines: `https://feeds.content.dowjones.io/public/rss/mw_realtimeheadlines`
  - Breaking news: `https://feeds.content.dowjones.io/public/rss/mw_bulletins`
  - Market pulse: `https://feeds.content.dowjones.io/public/rss/mw_marketpulse`
  - Client: `services/news_service.go` (lines 105-150)

## Monitoring & Observability

**Error Tracking:**
- None detected - No Sentry, Rollbar, or similar service

**Logs:**
- Local file system via logrus
  - Library: `github.com/sirupsen/logrus` (v1.9.3)
  - Format: Text with full timestamps
  - Output: stdout/stderr (not persisted to files by default)
  - Activity logging to disk: `./activity_logs/` directory (see `services/activity_logger.go`)

**Metrics:**
- None detected - No Prometheus, DataDog, or similar

## CI/CD & Deployment

**Hosting:**
- Not specified - Assumes local/self-hosted deployment
- Supports both paper trading (default) and live trading via Alpaca

**CI Pipeline:**
- GitHub repository present (see `package.json` line 16: `https://github.com/JakeNesler/Prophet_Trader.git`)
- No GitHub Actions workflows detected in `.github/workflows/`

**Deployment Scripts:**
- `autonomous_trading.sh` - Shell script for autonomous bot operation
- `run_autonomous.sh` - Alternative startup script
- `backfill_embeddings.js` - Node.js script for pre-computing trade embeddings

## Environment Configuration

**Required env vars (critical):**
- `ALPACA_API_KEY` - Alpaca public/API key
- `ALPACA_SECRET_KEY` - Alpaca secret key
- Validation: `cmd/bot/main.go` lines 41-44 - Bot will not start without these

**Optional env vars:**
- `GEMINI_API_KEY` - For AI news cleaning (optional)
- `ALPACA_BASE_URL` - Defaults to paper trading endpoint
- `ALPACA_PAPER` - Set to "false" for live trading (dangerous!)
- `DATABASE_PATH` - Custom database location
- `SERVER_PORT` - Custom HTTP port (default 4534)
- `ENABLE_LOGGING` - Enable/disable logging
- `LOG_LEVEL` - Logging verbosity
- `TRADING_BOT_URL` - MCP server connects to this bot endpoint

**Secrets location:**
- Environment variables via `.env` file (see `.env.example` for template)
- `.env` file should be in project root and is git-ignored
- Secrets are NEVER committed to version control

## API Endpoints Exposed

**Trading Bot HTTP API (Gin):**
- Port: 4534 (default, configurable via `SERVER_PORT`)
- Base path: `/api/v1`
- Health check: `GET /health`

**Order Management:**
- `POST /api/v1/orders/buy` - Place buy order
- `POST /api/v1/orders/sell` - Place sell order
- `DELETE /api/v1/orders/:id` - Cancel order
- `GET /api/v1/orders` - List orders

**Market Data:**
- `GET /api/v1/market/quote/:symbol` - Get stock quote
- `GET /api/v1/market/bar/:symbol` - Get latest bar
- `GET /api/v1/market/bars/:symbol` - Get multiple bars

**Account & Positions:**
- `GET /api/v1/account` - Account info (cash, buying power, portfolio value)
- `GET /api/v1/positions` - Current positions

**Options Trading:**
- `POST /api/v1/options/order` - Place options order
- `GET /api/v1/options/positions` - List options positions
- `GET /api/v1/options/position/:symbol` - Get specific options position
- `GET /api/v1/options/chain/:symbol` - Get options chain for underlying

**News & Intelligence:**
- `GET /api/v1/news` - Get latest news
- `GET /api/v1/news/topic/:topic` - Get news by topic
- `GET /api/v1/news/search?q=query` - Search news
- `GET /api/v1/news/market` - Get market news
- `GET /api/v1/news/marketwatch/*` - MarketWatch feeds (multiple endpoints)
- `POST /api/v1/intelligence/cleaned-news` - AI-cleaned news summary
- `GET /api/v1/intelligence/quick-market` - Quick market intelligence
- `GET /api/v1/intelligence/analyze/:symbol` - Analyze single stock
- `POST /api/v1/intelligence/analyze-multiple` - Analyze multiple stocks

**Position Management:**
- `POST /api/v1/positions/managed` - Create managed position
- `GET /api/v1/positions/managed` - List managed positions
- `GET /api/v1/positions/managed/:id` - Get managed position details
- `DELETE /api/v1/positions/managed/:id` - Close managed position

**Activity Logging:**
- `GET /api/v1/activity/current` - Current session activity
- `GET /api/v1/activity/:date` - Activity by date
- `GET /api/v1/activity` - List all activity logs
- `POST /api/v1/activity/session/start` - Start session
- `POST /api/v1/activity/session/end` - End session
- `POST /api/v1/activity/log` - Log activity

**Dashboard:**
- `Static /dashboard` - Serves static web dashboard from `./web` directory

## Model Context Protocol (MCP)

**MCP Server:**
- Location: `mcp-server.js`
- Version: 1.0.0
- Tools exposed (via MCP):
  - `get_account` - Account information
  - `get_positions` - Open positions
  - `get_orders` - All orders
  - `place_buy_order` - Execute buy order
  - `place_sell_order` - Execute sell order
  - `cancel_order` - Cancel order
  - And more trading/analysis tools
- Integration: Allows Claude and other AI agents to interact with trading bot via standardized MCP protocol

---

*Integration audit: 2026-02-11*
