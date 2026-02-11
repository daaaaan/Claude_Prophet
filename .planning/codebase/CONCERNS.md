# Codebase Concerns

**Analysis Date:** 2026-02-11

## Tech Debt

**Incomplete WebSocket Streaming Implementation:**
- Issue: `StreamBars()` in `services/alpaca_data.go` is a stub that returns an empty channel
- Files: `services/alpaca_data.go` (lines 150-168)
- Impact: Real-time streaming functionality advertised in interfaces is non-functional. Only polling-based data retrieval works.
- Fix approach: Implement proper Alpaca websocket connection using market data streaming API. Replace the empty goroutine with actual subscription management.

**Unimplemented Options Chain Parsing:**
- Issue: Strike price and option type parsing from OCC symbols is incomplete
- Files: `services/alpaca_trading.go` (line 358 - TODO comment)
- Impact: Options position management cannot correctly extract contract details from option symbols. Limits automated position management accuracy.
- Fix approach: Add OCC symbol parser to extract strike, expiration, and call/put from symbols like SPY251219C00679000.

## Known Bugs

**Config Loading Error Swallowing:**
- Symptoms: `.env` file missing throws non-fatal error that is suppressed
- Files: `config/config.go` (lines 26-29)
- Trigger: godotenv.Load() fails if no `.env` file exists (normal in containerized/production environments)
- Impact: Error is returned but not checked in main(), could mask actual configuration problems
- Workaround: Ensure `.env` file always exists (even if empty) or check error in main before logging fatality
- Root cause: `godotenv.Load()` returns error for missing file, but this is normal behavior

**Race Condition in Position Manager Lock Usage:**
- Symptoms: Position state could become inconsistent between read and modification operations
- Files: `services/position_manager.go` (lines 216-218, 290-295, 625-628)
- Trigger: Quick successive calls to modify and read position state without atomic transactions
- Problem: Positions are read with RLock, copied to slice, then released. During iteration over that slice, another goroutine could modify the actual position object in the map while checkPositions() operates on it.
- Impact: Position state inconsistencies during concurrent monitoring and CLI updates
- Safe modification: Use defer for all locks, or redesign to copy position data entirely before releasing lock

**Database Save Failures Not Blocking Operations:**
- Symptoms: Failed database saves log warnings but don't prevent operation continuity
- Files: `services/position_manager.go` (lines 221-223, 349)
- Pattern: `if err := pm.savePositionToDB(position); err != nil { pm.logger.WithError(err).Error(...) }` - continues execution
- Impact: Position state exists in memory but not persisted. Restart loses position management state.
- Risk: User believes position is managed and stopped when it's only tracked in-memory

## Security Considerations

**Environment Variable Exposure in Error Messages:**
- Risk: API keys could leak in error logs if SDK panics
- Files: `config/config.go` uses `os.Getenv()` without validation
- Current mitigation: Errors logged by logrus don't include actual key values
- Recommendations:
  1. Add validation that ALPACA_API_KEY and GEMINI_API_KEY are non-empty before config returns
  2. Never log full credential values in any error message
  3. Use config.Load() return value to fail-fast rather than logging Fatal in main

**Missing Input Validation:**
- Risk: Malformed symbols, negative quantities accepted without validation
- Files: `controllers/order_controller.go` (lines 62-105) - relies on Gin struct binding only
- Impact: Invalid orders sent to Alpaca could cause API errors or unexpected behavior
- Recommendations:
  1. Add symbol format validation (check against known exchanges)
  2. Validate quantity > 0 after parsing (Gin binding helps but not exhaustive)
  3. Add price validation for limit/stop orders (must be positive)

**Gemini API Key Management:**
- Risk: GEMINI_API_KEY optional but used without nil checks in some paths
- Files: `cmd/bot/main.go` (line 84), `mcp-server.js` (line 22)
- Impact: If key not provided, AI summarization will fail at runtime without graceful degradation
- Recommendations:
  1. Make Gemini service initialization optional
  2. Add feature flags for AI capabilities when key missing
  3. Return meaningful error if user attempts AI operations without API key

## Performance Bottlenecks

**Synchronous Position Monitoring Loop:**
- Problem: `MonitorPositions()` in `services/position_manager.go` checks all positions every 10 seconds sequentially
- Files: `services/position_manager.go` (lines 271-286)
- Cause: `checkPositions()` iterates through all positions synchronously, blocking until all checks complete
- Impact: With many positions, monitoring lag increases linearly. If price fetch takes 500ms and there are 20 positions, full cycle could take 10+ seconds.
- Improvement path:
  1. Add configurable monitoring interval (currently hardcoded 10s)
  2. Parallelize price updates using goroutines with bounded pool
  3. Implement position state caching to reduce API calls for unchanged positions
  4. Add metrics to track monitoring cycle duration

**Inefficient Database Queries:**
- Problem: No pagination or limits on database queries for historical bars
- Files: `database/storage.go` (lines 97-150)
- Impact: Fetching years of historical data loads entire dataset into memory, causing slowdown and memory spikes
- Improvement: Add limit parameter, implement cursor-based pagination

**Embedding Model Lazy Loading:**
- Problem: Local embedding model loads on first use, blocking first trade storage call
- Files: `vectorDB.js` (lines 11-20)
- Impact: First similarity search or trade storage will block for ~30 seconds (noted in comment) while model downloads
- Improvement: Pre-load embedding model during initialization, or use lazy loading with progress tracking

## Fragile Areas

**Position Manager State Consistency:**
- Files: `services/position_manager.go` (926 lines - largest Go file)
- Why fragile: Complex state machine with PENDING → ACTIVE → PARTIAL/CLOSED/STOPPED_OUT transitions. Multiple order types (entry, stop loss, take profit, partial exit) must be coordinated.
- Safe modification:
  1. Add comprehensive unit tests for state transitions
  2. Use atomic operations or transactions for multi-step updates
  3. Add invariant checks (e.g., verify StopLossOrderID is empty when creating new stop order)
  4. Test failure scenarios (order placement fails, price update fails, etc.)
- Test coverage: No test files found for position manager

**MCP Server Tool Implementation:**
- Files: `mcp-server.js` (1700+ lines)
- Why fragile: Single large file with 50+ tool implementations. Cascading error handling through try-catch blocks.
- Safe modification:
  1. Split tools into separate modules
  2. Add integration tests for each tool endpoint
  3. Test error paths (missing API key, network timeout, invalid symbol)
- Test coverage: No tests present

**Gemini API Integration:**
- Files: `services/gemini_service.go` (228 lines), `mcp-server.js` (lines 1096-1193)
- Why fragile: Prompt injection vulnerability - user queries concatenated directly into prompts
- Safe modification:
  1. Sanitize input before using in prompts
  2. Add length limits on user inputs
  3. Test with adversarial prompts
  4. Use prompt templates instead of string concatenation

**Options Data Parsing:**
- Files: `services/alpaca_options_data.go` (273 lines)
- Why fragile: Raw HTTP calls with manual JSON parsing, no schema validation
- Issues:
  1. Line 117, 177, 246: Raw `string(body)` in error messages could expose large API response data
  2. No timeout on HTTP requests
  3. Assumes API response format matches without validation
- Safe modification:
  1. Add struct tags with json validation
  2. Use custom HTTP client with timeout
  3. Add response schema validation

## Scaling Limits

**Single-Threaded Position Monitoring:**
- Current capacity: Efficient for <10 positions
- Limit: With 20+ positions, 10-second monitoring cycle time grows to 15-30 seconds
- Scaling path:
  1. Implement goroutine worker pool (5-10 workers)
  2. Use channels to distribute position checks
  3. Add metrics for monitoring latency
  4. Make tick interval configurable per position

**SQLite Database Scaling:**
- Current capacity: Good for <1 year of intraday bar data, <10K positions
- Limit: SQLite not optimal for concurrent writes under high-frequency updates
- Scaling path:
  1. For higher frequency trading, migrate to PostgreSQL
  2. Add connection pooling
  3. Implement batch writes to reduce transaction overhead
  4. Add database migration strategy

**Embedding Vector Storage:**
- Current capacity: sqlite-vec works for <10K trades
- Limit: Vector search performance degrades above 50K embeddings
- Scaling path:
  1. Consider dedicated vector DB (Qdrant, Weaviate) if scaling beyond 100K trades
  2. Implement periodic vector cleanup/archiving

## Dependencies at Risk

**sqlite-vec Alpha Version:**
- Risk: Package at v0.1.7-alpha.2, pre-release software
- Files: `package.json` (line 40), `optionalDependencies` (lines 42-47)
- Impact: No stable API guarantees. Updates could break vector search functionality.
- Current mitigation: Marked as optionalDependencies, so missing on some platforms won't break
- Migration plan:
  1. Monitor sqlite-vec releases for stable v1.0
  2. Prepare fallback to standard SQLite for vector search if needed
  3. Test upgrade path before committing to stable version

**@xenova/transformers Large Download:**
- Risk: First run requires 30+ seconds to download embedding model
- Files: `vectorDB.js` (lines 14-16)
- Impact: Slow initialization in production environments
- Migration plan:
  1. Pre-build container with model cached
  2. Or switch to smaller embedding model
  3. Or use API-based embeddings (cost trade-off)

**Alpaca SDK Dependency Chain:**
- Risk: `github.com/alpacahq/alpaca-trade-api-go/v3` and `marketdata` client both maintained
- Files: `go.mod` (lines 5-6)
- Impact: Breaking changes in either SDK could require major refactors
- Current health: Both actively maintained as of Feb 2025

## Missing Critical Features

**No Comprehensive Error Recovery:**
- Problem: Order placement failures don't retry or have recovery strategy
- Files: `services/alpaca_trading.go` (lines 83-87), `controllers/order_controller.go` (lines 90-94)
- Blocks: Transient network errors cause order placement to fail permanently
- Recommendation: Add exponential backoff retry for transient errors

**No Order Confirmation Mechanism:**
- Problem: Orders executed immediately without confirmation, no dryrun mode
- Files: `controllers/order_controller.go`, `mcp-server.js`
- Blocks: No safety check before executing real trades
- Recommendation: Add optional confirmation step, implement paper trading mode

**No Trade Execution Audit Trail:**
- Problem: Decisions logged to JSON files but no structured audit log
- Files: `decisive_actions/` (folder) - JSON files only
- Impact: Difficult to audit compliance, hard to trace decision reasoning in investigations
- Recommendation: Add structured audit table in database with immutable records

**No Position Liquidation on Server Shutdown:**
- Problem: If server crashes during market hours, positions remain open
- Files: `cmd/bot/main.go` (lines 131-137)
- Impact: Unmonitored positions could hit stop-loss without automation
- Recommendation: Implement graceful shutdown that closes all positions or alerts user

## Test Coverage Gaps

**No Unit Tests for Position Manager:**
- What's not tested: State transitions, concurrent updates, risk order logic
- Files: `services/position_manager.go` (926 lines)
- Risk: Core trading logic has zero test coverage. State transitions are complex and error-prone.
- Priority: HIGH - Rewrite position manager with TDD approach

**No Unit Tests for Trading Service:**
- What's not tested: Order placement, cancellation, error handling
- Files: `services/alpaca_trading.go` (426 lines)
- Risk: Trading operations depend on external API. Mock tests would catch logic bugs.
- Priority: HIGH

**No Integration Tests:**
- What's not tested: Trading bot -> MCP server -> Go backend communication
- Risk: MCP protocol changes would break Claude Code integration
- Priority: MEDIUM

**No E2E Tests:**
- What's not tested: Full trading workflows (entry -> stop loss -> exit)
- Risk: Real-world trading scenarios may not work correctly
- Priority: HIGH but requires testnet setup

**No Tests for News/Analysis Pipelines:**
- What's not tested: News aggregation, Gemini summarization, technical analysis
- Files: `services/news_service.go` (240 lines), `services/stock_analysis_service.go` (414 lines), `services/technical_analysis.go` (324 lines)
- Risk: Analysis tools could produce incorrect results silently
- Priority: MEDIUM

---

*Concerns audit: 2026-02-11*
