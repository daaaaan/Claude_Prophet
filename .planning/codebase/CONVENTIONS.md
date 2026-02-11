# Coding Conventions

**Analysis Date:** 2026-02-11

## Overview

This is a hybrid Go/JavaScript codebase (trading bot with MCP server). Go code is predominant in core business logic; JavaScript handles MCP protocol integration and vector database operations. Conventions differ slightly between languages but maintain consistency within each.

## Naming Patterns

### Go Files

**Files:**
- Snake case: `alpaca_trading.go`, `position_manager.go`, `technical_analysis.go`
- Directories: lowercase, descriptive: `services/`, `controllers/`, `models/`, `interfaces/`, `database/`, `config/`

**Functions and Methods:**
- PascalCase for exported (public): `PlaceOrder()`, `GetPositions()`, `NewOrderController()`
- camelCase for unexported (private): `getEnvOrDefault()`, `parseStrategies()`
- Constructors prefixed with `New`: `NewAlpacaTradingService()`, `NewOrderController()`
- HTTP handlers prefixed with `Handle`: `HandleBuy()`, `HandleGetPositions()`, `HandleCancelOrder()`
- Example: `services/alpaca_trading.go`:
  ```go
  func NewAlpacaTradingService(apiKey, secretKey, baseURL string, isPaper bool) (*AlpacaTradingService, error)
  func (s *AlpacaTradingService) PlaceOrder(ctx context.Context, order *interfaces.Order) (*interfaces.OrderResult, error)
  func (oc *OrderController) Buy(ctx context.Context, req BuyRequest) (*interfaces.OrderResult, error)
  ```

**Variables and Struct Fields:**
- camelCase: `apiKey`, `secretKey`, `orderID`, `buyingPower`
- JSON tags use snake_case: `json:"stop_loss_price"`, `json:"taking_profit_percent"`
- Example from `controllers/order_controller.go`:
  ```go
  type BuyRequest struct {
    Symbol      string   `json:"symbol" binding:"required"`
    Qty         float64  `json:"qty" binding:"required,gt=0"`
    LimitPrice  *float64 `json:"limit_price,omitempty"`
  }
  ```

**Types:**
- PascalCase (exported): `DBOrder`, `TradingService`, `OrderController`, `ManagedPosition`
- Interface names descriptive: `TradingService`, `DataService`, `StorageService`, `StrategyExecutor`
- Example: `interfaces/trading.go`:
  ```go
  type TradingService interface { ... }
  type OrderResult struct { ... }
  type OptionsPosition struct { ... }
  ```

### JavaScript Files

**Files:**
- camelCase or lowercase: `mcp-server.js`, `vectorDB.js`, `backfill_embeddings.js`
- Directories: lowercase: `seed_data/`, `decisive_actions/`

**Functions:**
- camelCase: `callTradingBot()`, `parseFilename()`, `extractTradeInfo()`, `getEmbedding()`, `storeTrade()`
- Async functions: `async function backfillAll()`, `async function getEmbedder()`
- Example from `vectorDB.js`:
  ```js
  export async function getEmbedding(text)
  export async function storeTrade(trade)
  export async function findSimilarTrades(queryText, limit = 5, filters = {})
  export function getTradeStats(filters = {})
  ```

**Variables:**
- camelCase: `embedder`, `dbPath`, `textToEmbed`, `queryEmbedding`, `tradeId`
- Constants: UPPER_SNAKE_CASE: `TRADING_BOT_URL`, `GEMINI_API_KEY`, `SUMMARIES_DIR`, `DECISIONS_DIR`
- Example from `mcp-server.js`:
  ```js
  const TRADING_BOT_URL = process.env.TRADING_BOT_URL || 'http://localhost:4534';
  const GEMINI_API_KEY = process.env.GEMINI_API_KEY;
  const SUMMARIES_DIR = path.join(process.cwd(), 'news_summaries');
  ```

**Destructuring:** Snake case in JSON payloads, camelCase in JavaScript variables
  ```js
  const { symbol, action, strategy, result_pct, result_dollars } = args; // From JSON
  const symbolData = { symbol, action }; // Internal usage
  ```

## Code Style

### Go Code Style

**Imports:**
- Standard library imports first
- External packages second
- Local imports last
- Organized alphabetically within groups
- Example from `services/alpaca_trading.go`:
  ```go
  import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "prophet-trader/interfaces"
    "time"

    "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
    "github.com/sirupsen/logrus"
  )
  ```

**Formatting:**
- gofmt standard (implicit, language convention)
- 2-space indentation (standard Go)
- Line length: ~100 characters (observed)

**Struct Tags:**
- JSON tags with `binding` validation for request structs: `binding:"required"`, `binding:"gt=0"`
- GORM tags for database models: `gorm:"index"`, `gorm:"uniqueIndex"`, `gorm:"index:idx_symbol_timestamp"`
- Example from `models/models.go`:
  ```go
  type DBOrder struct {
    OrderID    string  `gorm:"uniqueIndex"`
    Symbol     string  `gorm:"index"`
    Status     string  `gorm:"index"`
  }
  ```

### JavaScript Code Style

**Imports:**
- `import` statements (ES6 modules) - file specifies `"type": "module"` in `package.json`
- File extensions included: `import { storeTrade } from './vectorDB.js'`
- Grouped logically: standard library, external packages, local imports
- Example from `mcp-server.js`:
  ```js
  import { Server } from '@modelcontextprotocol/sdk/server/index.js';
  import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
  import { GoogleGenerativeAI } from '@google/generative-ai';
  import axios from 'axios';
  import fs from 'fs/promises';
  import path from 'path';
  import { storeTrade, findSimilarTrades } from './vectorDB.js';
  ```

**Formatting:**
- 2-space indentation (observed consistently)
- Semicolons used (style choice)
- Line length: ~100 characters

**Async/Await:**
- Async functions always use `async`/`await`, no `.then()` chains in main flow
- `try/catch` for error handling in async contexts
- Example from `mcp-server.js`:
  ```js
  async function callTradingBot(endpoint, method = 'GET', data = null) {
    try {
      const response = await axios(config);
      return response.data;
    } catch (error) {
      throw new Error(`Trading bot error: ${error.message}`);
    }
  }
  ```

## Import Organization

### Go

**Order:**
1. Standard library (`context`, `fmt`, `time`, `encoding/json`, etc.)
2. External packages (`github.com/...`)
3. Local packages (`prophet-trader/...`)

Observed in all Go files consistently.

### JavaScript

**Order:**
1. External packages (`@modelcontextprotocol/sdk`, `@google/generative-ai`, `axios`)
2. Node.js built-in (`fs`, `path`)
3. Local imports (`./vectorDB.js`, `../models`)

Example from `vectorDB.js`:
```js
import Database from 'better-sqlite3';
import * as sqliteVec from 'sqlite-vec';
import { pipeline } from '@xenova/transformers';
import path from 'path';
import { fileURLToPath } from 'url';
```

## Error Handling

### Go Pattern

**Standard error return tuple:**
- Functions return `(result, error)` as final return values
- Errors checked immediately with `if err != nil`
- Errors wrapped with context: `fmt.Errorf("error loading .env file: %v", err)`
- Example from `config/config.go`:
  ```go
  func Load() error {
    if err := godotenv.Load(); err != nil {
      return fmt.Errorf("error loading .env file: %v", err)
    }
    // ...
    return nil
  }
  ```

**Logging on error:**
- Use logrus fields for context: `logger.WithFields(logrus.Fields{...})`
- Log level appropriate to severity: `.Info()`, `.Warn()`, `.Error()`
- Example from `controllers/order_controller.go`:
  ```go
  if err := oc.storageService.SaveOrder(order); err != nil {
    oc.logger.WithError(err).Warn("Failed to save order to database")
  }
  ```

### JavaScript Pattern

**Try/catch blocks:**
- Async functions wrapped in try/catch
- Error message extraction: `error.message`
- Custom error wrapping: `throw new Error(\`context: ${error.message}\`)`
- Example from `mcp-server.js`:
  ```js
  try {
    const response = await axios(config);
    return response.data;
  } catch (error) {
    throw new Error(`Trading bot error: ${error.message}`);
  }
  ```

**Console logging:**
- `console.error()` for errors in scripts
- Direct error passing: `.catch(console.error)` for top-level promises
- Example from `seed_data/load_seed_data.js`:
  ```js
  } catch (error) {
    console.error(`Error storing trade for ${tradeData.symbol}:`, error.message);
    return false;
  }
  ```

## Logging

### Go Logging (logrus)

**Framework:** `github.com/sirupsen/logrus`

**Pattern:**
- Logger initialized per service/controller: `logger := logrus.New()`
- Formatter configured: `logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})`
- Structured logging with fields: `logger.WithFields(logrus.Fields{...})`
- Error attachment: `logger.WithError(err)`
- Log levels: `.Info()`, `.Warn()`, `.Error()`, `.Debug()`

**Examples from `services/alpaca_trading.go`:**
```go
s.logger.WithFields(logrus.Fields{
  "symbol": order.Symbol,
  "side":   order.Side,
  "qty":    order.Qty,
  "type":   order.Type,
}).Info("Processing order")

s.logger.WithError(err).Error("Failed to place order")
```

### JavaScript Logging

**Pattern:**
- Direct `console.log()` for info messages
- `console.error()` for errors
- Progress output via `process.stdout.write()`
- No structured logging framework used

**Examples from `seed_data/load_seed_data.js`:**
```js
console.log(`Loading ${principles.length} trading principles...`);
process.stdout.write(`\rPrinciples loaded: ${principlesLoaded}/${principles.length}`);
console.error(`Error storing trade for ${tradeData.symbol}:`, error.message);
```

**Examples from `vectorDB.js`:**
```js
console.log('🔄 Loading local embedding model (first run may take 30s)...');
console.log('✅ Embedding model loaded');
console.error('Embedding error:', error.message);
```

## Comments

### Go Comments

**Function documentation:** Comments on exported functions/types (GoDoc convention)
- Format: `// FunctionName does what`
- Start with function/type name
- Complete sentence

**Example from `interfaces/trading.go`:**
```go
// TradingService defines the interface for executing trades
type TradingService interface { ... }

// PlaceOrder places a new order
func PlaceOrder(ctx context.Context, order *Order) (*OrderResult, error)
```

**Inline comments:** Explain why, not what
- Used sparingly
- Example from `models/models.go`:
  ```go
  // Metadata for strategy tracking
  StrategyName string
  Metadata     string // JSON string for flexible data
  ```

### JavaScript Comments

**Block comments:** JSDoc-style for exported functions
```javascript
/**
 * Get local embedding for text (no API key required)
 * @param {string} text - Text to embed
 * @returns {Promise<number[]>} - 384-dimensional embedding vector
 */
export async function getEmbedding(text) { ... }
```

**Examples from `vectorDB.js`:**
```js
/**
 * Store trade decision with embedding for similarity search
 * @param {Object} trade - Trade decision object
 * @param {string} trade.id - Unique trade ID
 * @param {string} trade.decision_file - Path to decision JSON file
 * @returns {Promise<void>}
 */
export async function storeTrade(trade) { ... }
```

**Inline comments:** Explain non-obvious logic
- Example from `backfill_embeddings.js`:
  ```js
  /**
   * Parse decision filename to extract metadata
   * Format: YYYY-MM-DDTHH-MM-SS-MMMZ_ACTION.json
   */
  function parseFilename(filename) { ... }
  ```

**Section comments:** Major code blocks
- Example from `mcp-server.js`:
  ```js
  // Configuration
  const TRADING_BOT_URL = ...

  // Initialize Gemini
  const genAI = ...

  // Helper to call trading bot API
  async function callTradingBot(...) { ... }
  ```

## Function Design

### Go Functions

**Size:** Generally 50-100 lines for controller handlers, 20-50 for service methods

**Parameters:**
- Context always first: `func (s *Service) Method(ctx context.Context, ...)`
- Request objects for HTTP handlers: `func (oc *OrderController) Buy(ctx context.Context, req BuyRequest)`
- Avoid >3 positional parameters (use struct for complex data)

**Return values:**
- Errors always last: `func(...) (*Result, error)`
- Named returns NOT used (implicit)
- Multiple returns: `(result, error)` or `([]*Data, error)`

**Example from `controllers/order_controller.go`:**
```go
func (oc *OrderController) Buy(ctx context.Context, req BuyRequest) (*interfaces.OrderResult, error) {
  // Set defaults
  if req.Type == "" {
    req.Type = "market"
  }

  // Log operation
  oc.logger.WithFields(logrus.Fields{
    "symbol": req.Symbol,
    "qty":    req.Qty,
  }).Info("Processing buy order")

  // Convert request to interface
  order := &interfaces.Order{ ... }

  // Call service
  result, err := oc.tradingService.PlaceOrder(ctx, order)
  if err != nil {
    oc.logger.WithError(err).Error("Failed to place buy order")
    return nil, err
  }

  // Persist and return
  return result, nil
}
```

### JavaScript Functions

**Size:** 20-60 lines for async functions

**Parameters:**
- Options object for >2 parameters: `findSimilarTrades(queryText, limit = 5, filters = {})`
- Default values in signature: `async function backfillAll(options = {})`

**Return values:**
- Promises: `async function() { ... }` returns `Promise<T>`
- Direct returns from non-async: `function getTradeStats(filters = {}) { ... return stats; }`

**Example from `vectorDB.js`:**
```js
export async function storeTrade(trade) {
  try {
    // Create embedding
    const textToEmbed = `${trade.reasoning}\n\nMarket Context: ${trade.market_context}`;
    const embedding = await getEmbedding(textToEmbed);

    // Prepare and execute statement
    const insertTrade = db.prepare(`...`);
    insertTrade.run(trade.id, trade.decision_file, ...);

    // Return result
    return trade.id;
  } catch (error) {
    console.error('Embedding error:', error.message);
    throw error;
  }
}
```

## Module Design

### Go Modules/Packages

**Organization:** By domain responsibility
- `services/`: Business logic (trading, data, analysis)
- `controllers/`: HTTP handlers and orchestration
- `models/`: Database models with GORM
- `interfaces/`: Service interfaces and data structures
- `config/`: Configuration loading
- `database/`: Database initialization (if present)

**Exports:** All public (PascalCase) functions/types for interfaces
- `PlaceOrder()` exported so controllers can use
- Request/Response structs exported for JSON binding
- Private helpers (camelCase) within package only

**Example structure from `services/alpaca_trading.go`:**
```go
package services

// Service struct (exported)
type AlpacaTradingService struct { ... }

// Constructor (exported)
func NewAlpacaTradingService(...) (*AlpacaTradingService, error) { ... }

// Public methods (exported, implements interface)
func (s *AlpacaTradingService) PlaceOrder(...) (*interfaces.OrderResult, error) { ... }
```

### JavaScript Modules

**Organization:** By function or domain
- `mcp-server.js`: Main MCP protocol server
- `vectorDB.js`: Vector database operations (export/import)
- `backfill_embeddings.js`: Utility script
- `seed_data/load_seed_data.js`: Data loading utility

**Exports:** Explicit named exports for reusable functions
```js
export async function getEmbedding(text) { ... }
export async function storeTrade(trade) { ... }
export async function findSimilarTrades(...) { ... }
export function getTradeStats(filters = {}) { ... }
export function clearAllEmbeddings() { ... }
```

**Script-style files:** Top-level execution without export
- Example from `seed_data/load_seed_data.js`:
  ```js
  async function loadSeedData() { ... }
  loadSeedData().catch(console.error);
  ```

---

*Convention analysis: 2026-02-11*
