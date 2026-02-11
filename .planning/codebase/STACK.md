# Technology Stack

**Analysis Date:** 2026-02-11

## Languages

**Primary:**
- Go 1.22 - Core trading bot and API server (located in `cmd/bot/main.go`, `services/`, `controllers/`)
- JavaScript (Node.js) - MCP server and vector database operations (located in `mcp-server.js`, `vectorDB.js`, `backfill_embeddings.js`)

**Secondary:**
- Bash - Deployment and automation scripts (`autonomous_trading.sh`, `run_autonomous.sh`)

## Runtime

**Environment:**
- Go 1.22 (trading bot backend)
- Node.js (MCP server, vector operations)

**Package Manager:**
- npm (Node.js) - `package.json` with dependencies
- Lockfile: `package-lock.json` present
- Go modules - `go.mod` and `go.sum` for Go dependencies

## Frameworks

**Core:**
- Gin v1.10.0 - HTTP web framework for trading API (`cmd/bot/main.go` line 16)
- Model Context Protocol (MCP) @modelcontextprotocol/sdk v1.22.0 - Agent integration framework (`mcp-server.js`)

**Database/ORM:**
- GORM v1.25.12 - Go ORM for SQLite (`database/storage.go` line 13)
- sqlite-vec v0.1.7-alpha.2 - Vector search extension for SQLite (`vectorDB.js`, `package.json`)
- better-sqlite3 v12.4.6 - Node.js SQLite driver (`vectorDB.js`, `package.json`)

**Testing:**
- None detected - test script shows "no test specified" in `package.json` line 12

**Build/Dev:**
- Standard Go build toolchain (1.22)
- Node.js/npm for JavaScript components

## Key Dependencies

**Critical:**

### Go:
- `github.com/alpacahq/alpaca-trade-api-go/v3 v3.5.0` - Alpaca trading API client (`services/alpaca_trading.go` line 12)
- `github.com/gin-gonic/gin v1.10.0` - HTTP API server (`cmd/bot/main.go` line 16)
- `github.com/sirupsen/logrus v1.9.3` - Structured logging (`cmd/bot/main.go` line 17)
- `gorm.io/gorm v1.25.12` - ORM for database operations (`database/storage.go`)
- `gorm.io/driver/sqlite v1.5.6` - SQLite driver for GORM
- `github.com/joho/godotenv v1.5.1` - Environment variable loading (`config/config.go` line 8)
- `github.com/shopspring/decimal v1.3.1` - Precise decimal arithmetic for trading prices

### JavaScript:
- `@google/generative-ai ^0.24.1` - Google Gemini API client for AI analysis (`mcp-server.js` line 4)
- `@modelcontextprotocol/sdk ^1.22.0` - MCP server implementation
- `axios ^1.13.2` - HTTP client for API calls
- `@xenova/transformers ^2.17.2` - Local embedding model for vector operations (`vectorDB.js` line 4)
- `openai ^6.9.1` - OpenAI API support (present but may not be actively used)
- `cors ^2.8.5` - CORS middleware for Express
- `express ^5.1.0` - Express framework (appears to be dependency but not primary - Gin is used instead)

**Infrastructure:**
- SQLite with vector extensions - Local vector database for trade embeddings (`vectorDB.js`)
- Alpaca Markets API - Trading execution and market data
- Google Gemini API - AI-powered news analysis
- Google News RSS feeds - News source (via `services/news_service.go`)
- MarketWatch RSS feeds - Alternative news source

## Configuration

**Environment:**
- `.env` file loading via `github.com/joho/godotenv`
- Configuration structure in `config/config.go` (lines 10-21)
- Environment variables:
  - `ALPACA_API_KEY` - Alpaca trading API key (required)
  - `ALPACA_SECRET_KEY` - Alpaca secret key (required)
  - `ALPACA_BASE_URL` - Base URL (defaults to `https://paper-api.alpaca.markets`)
  - `ALPACA_PAPER` - Paper trading mode flag (defaults to true)
  - `GEMINI_API_KEY` - Google Gemini API key (optional)
  - `DATABASE_PATH` - SQLite database location (defaults to `./data/prophet_trader.db`)
  - `SERVER_PORT` - HTTP server port (defaults to 4534)
  - `ENABLE_LOGGING` - Enable logging (defaults to true)
  - `LOG_LEVEL` - Log level (defaults to info)
  - `TRADING_BOT_URL` - Trading bot API endpoint for MCP server (defaults to `http://localhost:4534`)

**Build:**
- Go: Standard `go build` (no custom config detected)
- Node.js: Standard npm build (no bundler configured)

## Platform Requirements

**Development:**
- Go 1.22 or later
- Node.js (version not specified in package.json)
- SQLite 3.x with sqlite-vec extension
- Bash shell for scripts

**Production:**
- Go 1.22 runtime
- Node.js runtime for MCP server
- SQLite database file storage at `./data/prophet_trader.db`
- Network access to:
  - Alpaca API (paper-api.alpaca.markets or live API)
  - Google News RSS feeds
  - MarketWatch RSS feeds
  - Google Gemini API
- Port 4534 exposed for HTTP API server

## Deployment Notes

- Cross-platform SQLite support via optional dependencies:
  - `sqlite-vec-darwin-arm64` - macOS ARM64
  - `sqlite-vec-darwin-x64` - macOS Intel
  - `sqlite-vec-linux-x64` - Linux x86-64
  - `sqlite-vec-win32-x64` - Windows x86-64
  - (See `package.json` lines 42-47 for optional dependency handling)

- Main entry points:
  - Go bot: `cmd/bot/main.go` - Run with `go run ./cmd/bot/main.go`
  - MCP server: `mcp-server.js` - Run with `npm start`

---

*Stack analysis: 2026-02-11*
