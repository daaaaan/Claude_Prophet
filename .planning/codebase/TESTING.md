# Testing Patterns

**Analysis Date:** 2026-02-11

## Current Testing Status

**No automated testing framework is currently configured.**

The `package.json` test script is a stub:
```json
"scripts": {
  "test": "echo \"Error: no test specified\" && exit 1"
}
```

No test files (`.test.js`, `.spec.js`, `.test.go`, etc.) exist in the codebase.

## Recommended Test Framework Setup

### JavaScript/Node.js Tests (for mcp-server.js, vectorDB.js)

**Recommended Framework:** Jest or Vitest

**Setup (Jest example):**
```bash
npm install --save-dev jest @types/jest
```

**package.json configuration:**
```json
{
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage"
  },
  "jest": {
    "testEnvironment": "node",
    "collectCoverageFrom": [
      "*.js",
      "!node_modules/**"
    ]
  }
}
```

**jest.config.js (if needed for ES modules):**
```js
export default {
  testEnvironment: 'node',
  transform: {},
  extensionsToTreatAsEsm: ['.js'],
};
```

### Go Tests (for services/, controllers/, config/)

**Framework:** Go's built-in `testing` package

**Test file naming:** `*_test.go` in same package directory
- Example: `services/alpaca_trading.go` → `services/alpaca_trading_test.go`

**Run tests:**
```bash
go test ./...          # All packages
go test ./services     # Specific package
go test -v ./...       # Verbose
go test -cover ./...   # Coverage
```

## Test Organization Patterns

### JavaScript Test Structure

**Recommended file location:** Co-located with source
```
project/
├── vectorDB.js
├── vectorDB.test.js
├── mcp-server.js
├── mcp-server.test.js
├── backfill_embeddings.js
└── backfill_embeddings.test.js
```

**Test file naming convention:** `<module>.test.js` or `<module>.spec.js`

**Suite organization pattern (Jest):**
```javascript
import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import { storeTrade, findSimilarTrades, getTradeStats } from './vectorDB.js';

describe('vectorDB', () => {
  beforeEach(() => {
    // Setup: clear database, mock embeddings
  });

  afterEach(() => {
    // Cleanup: restore state
  });

  describe('storeTrade', () => {
    it('should store trade with embedding', async () => {
      const trade = {
        id: 'test-1',
        symbol: 'SPY',
        action: 'BUY',
        reasoning: 'Test trade',
        market_context: 'Normal market'
      };

      const result = await storeTrade(trade);
      expect(result).toBe('test-1');
    });

    it('should throw on invalid trade data', async () => {
      await expect(storeTrade({})).rejects.toThrow();
    });
  });
});
```

### Go Test Structure

**Recommended file location:** Same package as code
```
services/
├── alpaca_trading.go
├── alpaca_trading_test.go
├── position_manager.go
└── position_manager_test.go
```

**Test function naming:** `Test<FunctionName>` format

**Test pattern (using standard library):**
```go
package services

import (
  "context"
  "testing"
)

func TestPlaceOrder(t *testing.T) {
  // Setup
  service := setupTestService(t)
  ctx := context.Background()

  order := &interfaces.Order{
    Symbol: "AAPL",
    Qty: 10,
    Side: "buy",
  }

  // Execute
  result, err := service.PlaceOrder(ctx, order)

  // Assert
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if result == nil {
    t.Fatal("expected non-nil result")
  }
  if result.OrderID == "" {
    t.Error("expected non-empty OrderID")
  }
}

func TestPlaceOrder_InvalidQty(t *testing.T) {
  service := setupTestService(t)
  ctx := context.Background()

  order := &interfaces.Order{
    Symbol: "AAPL",
    Qty: -10, // Invalid
    Side: "buy",
  }

  _, err := service.PlaceOrder(ctx, order)
  if err == nil {
    t.Fatal("expected error for negative quantity")
  }
}

func setupTestService(t *testing.T) *AlpacaTradingService {
  // Create mock/test service
  t.Helper()
  // Implementation here
  return &AlpacaTradingService{ /* ... */ }
}
```

## Mocking Strategy

### JavaScript Mocking (Jest)

**Use jest.mock() for external dependencies:**

**Example: Mock axios calls in mcp-server.test.js**
```javascript
import axios from 'axios';
jest.mock('axios');

describe('callTradingBot', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should return data on success', async () => {
    axios.mockResolvedValueOnce({
      data: { cash: 50000, buying_power: 100000 }
    });

    const result = await callTradingBot('/account');

    expect(result).toEqual({
      cash: 50000,
      buying_power: 100000
    });
    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'http://localhost:4534/api/v1/account'
      })
    );
  });

  it('should throw on axios error', async () => {
    axios.mockRejectedValueOnce(new Error('Network error'));

    await expect(callTradingBot('/account')).rejects.toThrow('Trading bot error');
  });
});
```

**Example: Mock vector database in vectorDB.test.js**
```javascript
jest.mock('better-sqlite3');
jest.mock('@xenova/transformers');

describe('getEmbedding', () => {
  it('should return embedding array', async () => {
    mockPipeline.mockResolvedValueOnce({
      data: new Float32Array([0.1, 0.2, 0.3, /* ... 381 more */ ])
    });

    const embedding = await getEmbedding('test text');

    expect(Array.isArray(embedding)).toBe(true);
    expect(embedding.length).toBe(384);
  });
});
```

### Go Mocking

**Use interface-based mocking (Go best practice):**

**Example: Mock TradingService in order_controller_test.go**
```go
package controllers

import (
  "context"
  "testing"
  "prophet-trader/interfaces"
)

// MockTradingService implements TradingService interface
type MockTradingService struct {
  PlaceOrderFunc func(ctx context.Context, order *interfaces.Order) (*interfaces.OrderResult, error)
}

func (m *MockTradingService) PlaceOrder(ctx context.Context, order *interfaces.Order) (*interfaces.OrderResult, error) {
  if m.PlaceOrderFunc != nil {
    return m.PlaceOrderFunc(ctx, order)
  }
  return nil, nil
}

// Test code
func TestOrderController_Buy_Success(t *testing.T) {
  mockTradingService := &MockTradingService{
    PlaceOrderFunc: func(ctx context.Context, order *interfaces.Order) (*interfaces.OrderResult, error) {
      return &interfaces.OrderResult{
        OrderID: "12345",
        Status: "filled",
      }, nil
    },
  }

  oc := NewOrderController(mockTradingService, nil, nil)
  result, err := oc.Buy(context.Background(), BuyRequest{
    Symbol: "AAPL",
    Qty: 10,
  })

  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if result.OrderID != "12345" {
    t.Errorf("expected OrderID 12345, got %s", result.OrderID)
  }
}
```

## What to Mock / What NOT to Mock

### JavaScript

**DO Mock:**
- External API calls (axios, HTTP requests)
- Database operations (better-sqlite3)
- ML model loading (transformers pipeline)
- File system when dealing with temp files
- Time-dependent functions (use `jest.useFakeTimers()`)

**DO NOT Mock:**
- Pure utility functions (parseFilename, extractTradeInfo)
- Core business logic (should test with real logic)
- Error handling paths (test actual errors)

### Go

**DO Mock:**
- External API clients (Alpaca API client)
- Database connections
- HTTP clients
- Service dependencies in controllers

**DO NOT Mock:**
- Request/Response struct creation
- Config loading (use test files)
- Core calculation logic
- Error handling

## Fixtures and Test Data

### JavaScript Fixtures

**Location:** `test/fixtures/` or co-located `__fixtures__/` directories

**Example: test/fixtures/trade.json**
```json
{
  "id": "2026-02-11-SPY-BUY-123456",
  "decision_file": "2026-02-11T12-00-00-000Z_BUY_SPY.json",
  "symbol": "SPY",
  "action": "BUY",
  "strategy": "SCALP",
  "result_pct": 2.5,
  "result_dollars": 500,
  "date": "2026-02-11",
  "reasoning": "Support level bounce",
  "market_context": "High volume breakout"
}
```

**Usage in tests:**
```javascript
import tradeFixture from './fixtures/trade.json' assert { type: 'json' };

describe('storeTrade', () => {
  it('should store valid trade', async () => {
    const result = await storeTrade(tradeFixture);
    expect(result).toBe(tradeFixture.id);
  });
});
```

### Go Fixtures

**Location:** `testdata/` directories alongside test files

**Example: services/testdata/orders.json**
```json
{
  "orders": [
    {
      "id": "order-123",
      "symbol": "AAPL",
      "qty": 10,
      "side": "buy",
      "status": "filled"
    }
  ]
}
```

**Usage in tests:**
```go
func TestPlaceOrder_WithFixture(t *testing.T) {
  data, err := ioutil.ReadFile("testdata/orders.json")
  if err != nil {
    t.Fatal(err)
  }

  var fixture map[string]interface{}
  json.Unmarshal(data, &fixture)

  // Use fixture in test
}
```

## Coverage Goals

**Recommended targets:**
- Critical business logic (trading, positions): 80%+
- Controllers (HTTP handlers): 70%+
- Utilities and helpers: 60%+
- Configuration: 40%+ (often straightforward)

**View coverage:**

JavaScript (Jest):
```bash
npm test -- --coverage
# Opens coverage/lcov-report/index.html
```

Go:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Types and Scopes

### Unit Tests (Primary)

**Scope:** Individual functions/methods in isolation

**Examples:**
- `storeTrade()` with mocked database
- `PlaceOrder()` with mocked Alpaca client
- `parseFilename()` with various filename formats
- `getTradeStats()` with fixture data

**JavaScript pattern:**
```javascript
describe('storeTrade unit tests', () => {
  it('should create embedding from reasoning text', async () => {
    mockEmbedding.mockResolvedValueOnce([0.1, 0.2, ...]);

    const trade = { reasoning: 'test', market_context: 'test' };
    await storeTrade(trade);

    expect(mockEmbedding).toHaveBeenCalledWith('test\n\nMarket Context: test');
  });
});
```

### Integration Tests (Secondary)

**Scope:** Multiple components working together, with real external services mocked at boundaries

**Examples:**
- OrderController.Buy() → TradingService → order storage (mocked Alpaca)
- MCP server receives tool call → processes → returns result
- VectorDB stores trade → finds similar trades

**Not yet implemented but recommended pattern:**
```javascript
describe('integration: place order flow', () => {
  it('should place order, log it, and persist', async () => {
    // Mock only Alpaca API
    mockAlpacaClient.placeOrder.mockResolvedValueOnce({ id: '123' });

    // Call through controller
    const result = await controller.Buy(request);

    // Verify full chain
    expect(mockAlpacaClient.placeOrder).toHaveBeenCalled();
    expect(mockStorage.saveOrder).toHaveBeenCalled();
    expect(result.OrderID).toBe('123');
  });
});
```

### E2E Tests (Not Currently Used)

**Would test:** Full trading flows with real or simulated trading environment

**Not implemented.** Candidates if added:
- End-to-end order placement and position tracking
- MCP server startup and tool availability
- Vector DB embedding and search accuracy

## Async Testing Patterns

### JavaScript (Jest)

**Async/await in tests:**
```javascript
it('should store and retrieve trade', async () => {
  const trade = { /* ... */ };

  const id = await storeTrade(trade);
  expect(id).toBeDefined();

  const found = await findSimilarTrades('buy', 1);
  expect(found.length).toBeGreaterThan(0);
});
```

**Promise handling:**
```javascript
it('should reject on invalid embedding', () => {
  mockEmbedding.mockRejectedValueOnce(new Error('Model failed'));

  return expect(getEmbedding('test')).rejects.toThrow('Model failed');
});
```

**Timeout for long operations:**
```javascript
it('should load embeddings', async () => {
  // Allow up to 30 seconds (default is 5)
  const result = await getEmbedder();
  expect(result).toBeDefined();
}, 30000);
```

### Go

**Context with timeout in tests:**
```go
func TestPlaceOrder_WithTimeout(t *testing.T) {
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  _, err := service.PlaceOrder(ctx, order)
  if err == context.DeadlineExceeded {
    t.Fatal("order placement timed out")
  }
}
```

## Error Testing Patterns

### JavaScript

**Testing error conditions:**
```javascript
describe('error handling', () => {
  it('should throw on invalid trade ID', async () => {
    const trade = { id: '', symbol: 'SPY' }; // Empty ID invalid

    await expect(storeTrade(trade)).rejects.toThrow('Trade ID required');
  });

  it('should handle embedding model failures', async () => {
    mockEmbedding.mockRejectedValueOnce(new Error('CUDA error'));

    await expect(getEmbedding('test')).rejects.toThrow();
  });

  it('should catch and wrap axios errors', async () => {
    mockAxios.mockRejectedValueOnce(new Error('ECONNREFUSED'));

    const error = await callTradingBot('/account').catch(e => e);
    expect(error.message).toContain('Trading bot error');
  });
});
```

### Go

**Testing error returns:**
```go
func TestPlaceOrder_InvalidSymbol(t *testing.T) {
  ctx := context.Background()
  order := &interfaces.Order{
    Symbol: "", // Invalid
    Qty: 10,
  }

  _, err := service.PlaceOrder(ctx, order)
  if err == nil {
    t.Fatal("expected error for empty symbol")
  }
}

func TestPlaceOrder_NegativeQty(t *testing.T) {
  _, err := service.PlaceOrder(context.Background(), &interfaces.Order{
    Symbol: "AAPL",
    Qty: -10,
  })

  if !strings.Contains(err.Error(), "quantity must be positive") {
    t.Errorf("unexpected error: %v", err)
  }
}
```

## Run Commands

### JavaScript

```bash
npm test                           # Run all tests once
npm run test:watch                # Run in watch mode (re-run on file change)
npm run test:coverage             # Generate coverage report
jest vectorDB.test.js             # Run specific test file
jest -t "storeTrade"              # Run tests matching pattern
jest --verbose                    # Detailed output
```

### Go

```bash
go test ./...                     # Run all package tests
go test -v ./services            # Verbose output, services package only
go test -run TestPlaceOrder ./...  # Run tests matching regex
go test -cover ./...              # Show coverage percentages
go test -coverprofile=out.txt ./... && go tool cover -html=out.txt
go test -race ./...               # Enable race detector
```

---

*Testing analysis: 2026-02-11*

**Note:** This codebase currently has zero automated tests. The patterns above are recommendations for implementation based on the codebase structure and dependencies (Jest for JavaScript, Go testing for Go code). Priority areas for initial test coverage: vector database operations (storeTrade, findSimilarTrades), order placement flow (controllers/order_controller.go), and critical services (alpaca_trading.go, position_manager.go).
